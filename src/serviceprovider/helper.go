package serviceprovider

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/helpers"
)

func TestCacheFolderAccess(ctx basecontext.ApiContext) error {
	cfg := config.Get()
	if !cfg.IsApi() {
		ctx.LogInfof("Not in API mode, skipping cache folder access test")
		return nil
	}
	cacheFolder, err := cfg.CatalogCacheFolder()
	if err != nil {
		return err
	}
	touchCommand := helpers.Command{
		Command: "touch",
		Args:    []string{filepath.Join(cacheFolder, "test_cache_access.txt")},
	}
	_, stderr, exitCode, err := helpers.ExecuteWithOutput(context.Background(), touchCommand, time.Second*1)
	if err != nil || exitCode != 0 {
		return fmt.Errorf("Error creating test file in cache folder: %v, stderr: %s", err, stderr)
	}
	removeCommand := helpers.Command{
		Command: "rm",
		Args:    []string{filepath.Join(cacheFolder, "test_cache_access.txt")},
	}
	_, stderr, exitCode, err = helpers.ExecuteWithOutput(context.Background(), removeCommand, time.Second*1)
	if err != nil || exitCode != 0 {
		return fmt.Errorf("Error removing test file in cache folder: %v, stderr: %s ", err, stderr)
	}
	return nil
}
