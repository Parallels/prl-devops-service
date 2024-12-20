package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	PROVIDER_VAR_NAME = "provider"
)

// MoveContentsToRoot moves all contents of the provided directory (srcDir)
// into rootDir. It overwrites any existing files with the same name in rootDir.
func MoveContentsToRoot(rootDir, srcDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %q: %w", srcDir, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		destPath := filepath.Join(rootDir, entry.Name())

		// If a file/folder with the same name already exists in rootDir, remove it first.
		if _, err := os.Stat(destPath); err == nil {
			// Remove existing destination file/directory to avoid rename conflicts.
			if err := os.RemoveAll(destPath); err != nil {
				return fmt.Errorf("failed to remove existing destination %q: %w", destPath, err)
			}
		}

		// Move the file or directory
		if err := os.Rename(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to move %q to %q: %w", srcPath, destPath, err)
		}
	}

	return nil
}

// CleanAndFlatten checks the root directory for any folders that end with .macvm or .pvm.
// If found, it moves all their contents into the root directory and removes the original folder.
func CleanAndFlatten(rootDir string) error {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to read root directory %q: %w", rootDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			if strings.HasSuffix(name, ".macvm") || strings.HasSuffix(name, ".pvm") {
				vmDir := filepath.Join(rootDir, name)

				// Move all contents from vmDir to rootDir
				if err := MoveContentsToRoot(rootDir, vmDir); err != nil {
					return fmt.Errorf("failed to move contents of %q to %q: %w", vmDir, rootDir, err)
				}

				// Remove the now-empty directory
				if err := os.RemoveAll(vmDir); err != nil {
					return fmt.Errorf("failed to remove directory %q: %w", vmDir, err)
				}
			}
		}
	}

	return nil
}
