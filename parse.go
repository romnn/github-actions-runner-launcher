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
