package githubactionsrunnerlauncher

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/google/go-github/v31/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// Version is incremented using bump2version
const Version = "0.1.1"

// Launcher ...
type Launcher struct {
	Config         LaunchConfig
	RunnerVersion  string
	RunnerArch     string
	Reconfigure    bool
	RemoveExisting bool
	configPath     string
	apiClient      *github.Client
	aptMux         sync.Mutex
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
			if err := l.ConfigureRunner(rLog, rConf, rToken); err != nil {
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
	workDir, err := l.GetWorkDirForRunner(runnerConfig)
	if err != nil {
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

// ConfigureRunner ...
func (l *Launcher) ConfigureRunner(rLog *log.Entry, runner RunnerConfig, runnerToken string) error {
	workDir, err := l.GetWorkDirForRunner(runner)
	if err != nil {
		return err
	}
	if err := l.PrepareRunnerFiles(rLog, runner); err != nil {
		return fmt.Errorf("Failed to prepare runner: %v", err)
	}
	cmd := exec.Command(filepath.Join(workDir, "./config.sh"), "--url", runner.Environment.RepoURL, "--token", runnerToken, "--name", runner.Environment.RunnerName, "--work", workDir, "--labels", runner.Environment.Labels, "--unattended", "--replace")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	rLog.Debug(string(out))
	if l.Reconfigure && err != nil {
		rLog.Error(cmd.String())
		rLog.Error(string(out))
		return fmt.Errorf("Failed to configure runner and Reconfigure option is set: %v", err)
	}
	return nil
}

// CreateGitHubAPIClient ...
func CreateGitHubAPIClient(ctx context.Context, accessToken string) *github.Client {
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
