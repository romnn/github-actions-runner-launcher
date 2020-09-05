package githubactionsrunnerlauncher

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/google/go-github/v31/github"
	log "github.com/sirupsen/logrus"
)

// RemoveRunner ....
func (l *Launcher) RemoveRunner(runner RunnerConfig, removeToken *github.RemoveToken) error {
	workDir, err := l.GetWorkDirForRunner(runner)
	if err != nil {
		return err
	}

	configFile := filepath.Join(workDir, "./config.sh")
	configFileExists, err := fileExists(configFile)
	if err != nil {
		return fmt.Errorf("failed to check for existing config file: %v", err)
	}
	if !configFileExists {
		return errors.New("cannot remove configured runner without config file")
	}
	removeCmd := exec.Command(configFile, "remove", "--token", removeToken.GetToken())
	err = removeCmd.Start()
	err = removeCmd.Wait()
	if err != nil {
		return fmt.Errorf("Failed to remove runner: %v", err)
	}
	return nil
}

// PrepareRunnerFiles ...
func (l *Launcher) PrepareRunnerFiles(rLog *log.Entry, runner RunnerConfig) error {
	workDir, err := l.GetWorkDirForRunner(runner)
	if err != nil {
		return err
	}

	runFile := filepath.Join(workDir, "./run.sh")
	configFile := filepath.Join(workDir, "./config.sh")
	runnerArchive := filepath.Join(workDir, "actions.tar.gz")

	runFileExists, runErr := fileExists(runFile)
	configFileExists, configErr := fileExists(configFile)
	runnerArchiveExists, archiveErr := fileExists(runnerArchive)

	if runErr != nil || configErr != nil || archiveErr != nil {
		return fmt.Errorf("failed to check for necessary runner files")
	}

	// Check for unarchived files
	if !runFileExists || !configFileExists {
		rLog.Infof("no runner files in %s", workDir)
		// Check if archive exists
		if !runnerArchiveExists {
			// Download it
			url := fmt.Sprintf("https://github.com/actions/runner/releases/download/v%s/actions-runner-linux-%s-%s.tar.gz", l.RunnerVersion, l.RunnerArch, l.RunnerVersion)
			rLog.Infof("Downloading from %s", url)
			if err := downloadFile(runnerArchive, url); err != nil {
				return fmt.Errorf("Failed to download the actions runner archive: %v", err)
			}
		}

		// Untar the archive
		rLog.Infof("Extracting %s", runnerArchive)
		tarCmd := exec.Command("tar", "-zxf", runnerArchive, "-C", workDir)
		err = tarCmd.Start()
		err = tarCmd.Wait()
		if err != nil {
			return fmt.Errorf("Failed to untar runner archive: %v", err)
		}

		// Install deps
		if false {
			// Only one runner can install at the same time
			l.aptMux.Lock()
			rLog.Info("Installing dependencies")
			cmdScript := filepath.Join(workDir, "bin/installdependencies.sh")
			rLog.Info(cmdScript)
			depCmd := exec.Command(cmdScript)
			// depCmd.Dir = workDir
			if out, err := depCmd.CombinedOutput(); err != nil {
				rLog.Warning(cmdScript)
				rLog.Warning(string(out))
				rLog.Warningf("Failed to install runner dependencies: %v", err)
				// return fmt.Errorf("Failed to install runner dependencies: %v", err)
			}
			l.aptMux.Unlock()
		}
	}
	return nil
}
