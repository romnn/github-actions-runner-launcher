package githubactionsrunnerlauncher

import (
	"fmt"
	"io/ioutil"

	"github.com/k0kubun/pp"
	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

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
