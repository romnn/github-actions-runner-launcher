package githubactionsrunnerlauncher

import (
	"fmt"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// ParseConfigFile ...
func (l *Launcher) prepareRunnerFiles(rLog *log.Entry, runner RunnerConfig) error {
	workDir, err := l.GetWorkDirForRunner(runner)
	if err != nil {
		return err
	}

	var runFile, configFile, runnerArchive string
	var runFileExists, configFileExists, runnerArchiveExists bool

	runFile = filepath.Join(workDir, "./run.sh")
	configFile = filepath.Join(workDir, "./config.sh")
	runnerArchive = filepath.Join(workDir, "actions.tar.gz")

	runFileExists, err = fileExists(runFile)
	configFileExists, err = fileExists(configFile)
	runnerArchiveExists, err = fileExists(runnerArchive)

	if err != nil {
		return fmt.Errorf("Failed to check for necessary runner files")
	}

	// Check for unarchived files
	if !runFileExists || !configFileExists {
		rLog.Infof("Not finding runner files in %s", workDir)
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
		cmd := exec.Command("tar", "-zxf", runnerArchive, "-C", workDir)
		err = cmd.Start()
		err = cmd.Wait()
		if err != nil {
			return fmt.Errorf("Failed to untar runner archive: %v", err)
		}

		// Install deps
		rLog.Info("Installing dependencies")
		if err := exec.Command(filepath.Join(workDir, "./bin/installdependencies.sh")).Run(); err != nil {
			return fmt.Errorf("Failed to install runner dependencies: %v", err)
		}
	}
	return nil
}
