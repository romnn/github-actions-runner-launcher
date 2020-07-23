package githubactionsrunnerlauncher

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"

	"github.com/google/go-github/v31/github"
	"github.com/k0kubun/pp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

// Version is incremented using bump2version
const Version = "0.1.1"

// RunnerEnvironment ...
type RunnerEnvironment struct {
	RepoURL       string `yaml:"REPO_URL"`
	AccessToken   string `yaml:"ACCESS_TOKEN"`
	RunnerName    string `yaml:"RUNNER_NAME"`
	RunnerToken   string `yaml:"RUNNER_TOKEN"`
	RunnerWorkdir string `yaml:"RUNNER_WORKDIR"`
	IsOrgRunner   string `yaml:"ORG_RUNNER"`
	OrgName       string `yaml:"ORG_NAME"`
	Labels        string `yaml:"LABELS"`
}

// RunnerConfig ...
type RunnerConfig struct {
	Environment RunnerEnvironment `yaml:"environment"`
}

// LaunchConfig ...
type LaunchConfig struct {
	Services map[string]RunnerConfig `yaml:"services"`
}

// Launcher ...
type Launcher struct {
	Config           LaunchConfig
	RunnerVersion    string
	RunnerArch       string
	ForceReconfigure bool
	ForceRemoveExisting bool
	configPath       string
	apiClient        *github.Client
	aptMux 			 sync.Mutex
}

// NewWithConfig ...
func NewWithConfig(file string) (*Launcher, error) {
	l := Launcher{
		RunnerArch:    "x64",
		RunnerVersion: "2.169.1",
	}
	if err := l.ParseConfigFile(file); err != nil {
		return nil, err
	}
	return &l, nil
}

// Run ...
func (l *Launcher) Run(run bool) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-shutdown
		log.Info("Stopping runners...")
		cancel()
	}()

	var wg sync.WaitGroup
	for runnerName, runnerConfig := range l.Config.Services {

		// Check for name override
		if nameOverride := runnerConfig.Environment.RunnerName; nameOverride != "" {
			runnerName = nameOverride
		}

		rLog := log.WithFields(log.Fields{"runner": runnerName})
		rLog.Info("Starting")

		runnerToken := runnerConfig.Environment.RunnerToken
		if runnerToken == "" {
			// Check for access token
			accessToken := runnerConfig.Environment.AccessToken
			if accessToken == "" {
				return errors.New("One of RUNNER_TOKEN or ACCESS_TOKEN is required to start a runner")
			}
			// Obtain a runner token using the access token
			var expireTime *github.Timestamp
			runnerToken, expireTime, err = l.ObtainRunnerToken(context.Background(), runnerConfig, accessToken)
			if err != nil {
				return fmt.Errorf("Failed to create runner token using access token: %v", err)
			}
			rLog.Infof("Obtained RUNNER_TOKEN=%s (expires on %v)", runnerToken, expireTime)
		}

		go func(rConf RunnerConfig, rToken string) {
			wg.Add(1)
			if err := l.configureRunner(rLog, rConf, rToken); err != nil {
				rLog.Error(err)
			}
			if run {
				l.startRunner(ctx, rLog, &wg, rConf, rToken)
			}
			wg.Done()
		}(runnerConfig, runnerToken) 
	}
	wg.Wait()
	return nil
}

func (l *Launcher) startRunner(ctx context.Context, rLog *log.Entry, wg *sync.WaitGroup, runnerConfig RunnerConfig, runnerToken string) {
	// wg.Add(1)
	// defer wg.Done()

	workDir, err := l.GetWorkDirForRunner(runnerConfig)
	if err != nil {
		rLog.Error(err)
		return
	}

	if err := l.configureRunner(rLog, runnerConfig, runnerToken); err != nil {
		rLog.Error(err)
		return
	}
	cmd := exec.CommandContext(ctx, filepath.Join(workDir, "./run.sh"))
	cmd.Dir = workDir

	// Stream stdout and stderr
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		rLog.Error(err)
		return
	}
	err = cmd.Start()
	if err != nil {
		rLog.Error(err)
		return
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m := scanner.Text()
			log.Info(m)
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			m := scanner.Text()
			log.Error(m)
		}
	}()

	if err := cmd.Wait(); err != nil && ctx.Err() == nil {
		rLog.Errorf("Failed to run runner: %v", err)
	}
}

// ParseConfigFile ...
func (l *Launcher) configureRunner(rLog *log.Entry, runner RunnerConfig, runnerToken string) error {
	workDir, err := l.GetWorkDirForRunner(runner)
	if err != nil {
		return err
	}
	if err := l.prepareRunnerFiles(rLog, runner); err != nil {
		return fmt.Errorf("Failed to prepare runner: %v", err)
	}
	cmd := exec.Command(filepath.Join(workDir, "./config.sh"), "--url", runner.Environment.RepoURL, "--token", runnerToken, "--name", runner.Environment.RunnerName, "--work", workDir, "--labels", runner.Environment.Labels, "--unattended", "--replace")
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); l.ForceReconfigure && err != nil {
		rLog.Error(cmd.String())
		rLog.Error(string(out))
		return fmt.Errorf("Failed to configure runner and ForceReconfigure option is set: %v", err)
	}
	return nil
}

// ParseConfigFile ...
func (l *Launcher) ParseConfigFile(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Failed to read config file %v: %v", file, err)
	}
	if err := yaml.Unmarshal(data, &l.Config); err != nil {
		return fmt.Errorf("Failed to parse config: %v", err)
	}
	if log.IsLevelEnabled(log.DebugLevel) {
		pp.Print(l.Config)
	}
	l.configPath = file
	log.Info("Parsed config")
	return nil
}

// ObtainRunnerToken ...
func (l *Launcher) ObtainRunnerToken(ctx context.Context, runner RunnerConfig, accessToken string) (string, *github.Timestamp, error) {
	if l.apiClient == nil {
		l.apiClient = createGitHubAPIClient(ctx, accessToken)
	}
	r := regexp.MustCompile(`^.*github\.com/(?P<Acc>.*)/(?P<Repo>.*)$`)
	matches := r.FindStringSubmatch(runner.Environment.RepoURL)
	if len(matches) != 3 {
		return "", nil, fmt.Errorf("Failed to extract github account and repo name from \"%s\"", runner.Environment.RepoURL)
	}
	acc, repo := matches[1], matches[2]
	log.Debugf("acc=%s, repo=%s", acc, repo)
	if acc == "" || repo == "" {
		return "", nil, fmt.Errorf("Failed to extract github account and repo name from \"%s\" (acc=%s, repo=%s)", runner.Environment.RepoURL, acc, repo)
	}

	runners, _, err := l.apiClient.Actions.ListRunners(ctx, acc, repo, &github.ListOptions{
		Page: 0,
		PerPage: 100,
	})
	if err != nil {
		log.Warnf("Failed to check for any existing runners: %v", err)
	} else if l.ForceRemoveExisting {
		log.Infof("%d existing runners will be removed", runners.TotalCount)
		for _, runner := range runners.Runners {
			log.Infof("Removing runner %s (%d) [os=%s, status=%s]", runner.GetName(), runner.GetID(), runner.GetOS(), runner.GetStatus())
			if _, err := l.apiClient.Actions.RemoveRunner(ctx, acc, repo, runner.GetID()); err != nil {
				log.Warnf("Failed to remove runner %d: %v", runner.GetID(), err)
			}
		}
	}
	token, response, err := l.apiClient.Actions.CreateRegistrationToken(ctx, acc, repo)
	log.Debugf("runner_token=%v", token)
	if err != nil || token.Token == nil {
		return "", nil, fmt.Errorf("Failed to obtain runner token from the GitHub API: %v", err)
	}
	log.Debugf("%d of %d GitHub API requests left. Will be reset on %v", response.Rate.Remaining, response.Rate.Limit, response.Rate.Reset)
	if response.Rate.Remaining < 50 {
		log.Warningf("Only %d of %d GitHub API requests left. Will be reset on %v", response.Rate.Remaining, response.Rate.Limit, response.Rate.Reset)
	}
	return *token.Token, token.ExpiresAt, nil
}

func createGitHubAPIClient(ctx context.Context, accessToken string) *github.Client {
	tokenSource := &TokenSource{
		AccessToken: accessToken,
	}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	apiClient := github.NewClient(oauthClient)
	if user, _, err := apiClient.Users.Get(ctx, ""); err == nil {
		log.Infof("Authenticated as %s", user.GetName())
	}
	return apiClient
}
