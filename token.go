package githubactionsrunnerlauncher

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/go-github/v31/github"
	log "github.com/sirupsen/logrus"
)

// ObtainRunnerToken ...
func (l *Launcher) ObtainRunnerToken(ctx context.Context, runnerCfg RunnerConfig, accessToken string) (string, *github.Timestamp, error) {
	if l.apiClient == nil {
		l.apiClient = CreateGitHubAPIClient(ctx, accessToken)
	}
	r := regexp.MustCompile(`^.*github\.com/(?P<Acc>.*)/(?P<Repo>.*)$`)
	matches := r.FindStringSubmatch(runnerCfg.Environment.RepoURL)
	if len(matches) != 3 {
		return "", nil, fmt.Errorf("Failed to extract github account and repo name from \"%s\"", runnerCfg.Environment.RepoURL)
	}
	acc, repo := matches[1], matches[2]
	log.Debugf("acc=%s, repo=%s", acc, repo)
	if acc == "" || repo == "" {
		return "", nil, fmt.Errorf("Failed to extract github account and repo name from \"%s\" (acc=%s, repo=%s)", runnerCfg.Environment.RepoURL, acc, repo)
	}
	if l.RemoveExisting {
		log.Infof("Removing runner %s", runnerCfg.Environment.RunnerName)
		removeToken, _, err := l.apiClient.Actions.CreateRemoveToken(ctx, acc, repo)
		if err != nil {
			log.Warnf("Failed to get a remove token: %v", err)
		}
		if err := l.RemoveRunner(runnerCfg, removeToken); err != nil {
			log.Warnf("Failed to remove runner %s: %v", runnerCfg.Environment.RunnerName, err)
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
