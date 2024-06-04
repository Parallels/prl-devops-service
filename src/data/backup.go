package data

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/cjlapao/common-go/helper"
)

func (j *JsonDatabase) Backup(ctx basecontext.ApiContext) error {
	backupFiles, err := findBackupFiles(j.filename)
	if err != nil {
		ctx.LogErrorf("[Database] Error finding backup files: %v", err)
		return err
	}

	if len(backupFiles) >= j.Config.NumberOfBackupFiles {
		// Delete the oldest backup file
		oldestFile := backupFiles[0]
		err := os.Remove(oldestFile)
		if err != nil {
			ctx.LogErrorf("[Database] Error deleting backup file: %v", err)
			return err
		}
	}

	// Create a new backup file with timestamp
	timestamp := time.Now().Format("20060102150405")
	newBackupFile := fmt.Sprintf("%s.save.bak.%s", j.filename, timestamp)
	err = helper.CopyFile(j.filename, newBackupFile)
	if err != nil {
		ctx.LogErrorf("[Database] Error creating new backup file: %v", err)
		return err
	}

	return nil
}

func findBackupFiles(filename string) ([]string, error) {
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)
	pattern := fmt.Sprintf("%s.save.bak.*", base)
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, err
	}

	// Sort the backup files by timestamp
	sort.Slice(matches, func(i, j int) bool {
		return extractTimestamp(matches[i]) < extractTimestamp(matches[j])
	})

	return matches, nil
}

func extractTimestamp(filename string) string {
	parts := strings.Split(filename, ".")
	return parts[len(parts)-1]
}
