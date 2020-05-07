package githubactionsrunnerlauncher

import (
	"errors"
	"os"
	"path/filepath"
)

// GetWorkDirForRunner ...
func (l *Launcher) GetWorkDirForRunner(runner RunnerConfig) (string, error) {
	wd := runner.Environment.RunnerWorkdir
	if wd == "" {
		return "", errors.New("missing workdir for runner")
	}
	if filepath.IsAbs(wd) {
		return wd, nil
	}
	// Combine relative with base
	if l.configPath == "" {
		return "", errors.New("missing config dir to obtain runner workdir")
	}
	return filepath.Join(l.configPath, wd), nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
