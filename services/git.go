package services

import (
	"Parallels/pd-api-service/helpers"
	"fmt"
	"os"
	"strings"

	"github.com/cjlapao/common-go/commands"
)

var globalGitService *GitService

type GitService struct {
	executable string
}

func NewGitService() *GitService {
	if globalGitService != nil {
		return globalGitService
	}

	globalGitService = &GitService{}

	globalServices.Logger.Info("Getting Git executable")
	out, err := commands.ExecuteWithNoOutput("which", "git")
	path := strings.ReplaceAll(strings.TrimSpace(out), "\n", "")
	if err != nil || path == "" {
		globalServices.Logger.Warn("Git executable not found, trying to find it in the default locations")
	}

	if path != "" {
		globalGitService.executable = path
	} else {
		if _, err := os.Stat("/opt/homebrew/bin/git"); err == nil {
			globalGitService.executable = "/opt/homebrew/bin/git"
		} else {
			globalServices.Logger.Warn("Git executable not found")
			return nil
		}

		globalServices.Logger.Info("Git executable: " + globalGitService.executable)
	}

	return globalGitService
}

func (s *GitService) Clone(repoURL string, localPath string) (string, error) {
	path := fmt.Sprintf("/tmp/%s", localPath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := helpers.CreateDirIfNotExist(path)
		if err != nil {
			return "", err
		}

		_, err = commands.ExecuteWithNoOutput(s.executable, "clone", repoURL, path)
		if err != nil {
			buildError := fmt.Errorf("Failed to pull repository %v, error: %v", path, err.Error())
			return "", buildError
		}
	} else {
		_, err := helpers.ExecuteWithNoOutput(helpers.Command{
			Command:          s.executable,
			Args:             []string{"pull"},
			WorkingDirectory: path,
		})
		if err != nil {
			buildError := fmt.Errorf("Failed to pull repository %v, error: %v", path, err.Error())

			return "", buildError
		}
	}

	return path, nil
}
