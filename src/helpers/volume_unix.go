//go:build linux || darwin
// +build linux darwin

package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

func extractVolumeName(path string) (string, error) {
	if runtime.GOOS == "windows" {
		return "", fmt.Errorf("not supported on windows")
	}
	if runtime.GOOS == "linux" {
		return "", fmt.Errorf("not supported on linux")
	}

	if !strings.HasPrefix(path, "/Volumes/") {
		volume := filepath.Dir(path)
		return volume, nil
	}

	// Clean up the path to remove any redundant separators
	cleanedPath := filepath.Clean(path)

	// Split the path into parts based on the separator
	parts := strings.Split(cleanedPath, string(filepath.Separator))

	// Check if the path has enough parts to contain a volume name
	if len(parts) < 3 || parts[1] != "Volumes" {
		return "", fmt.Errorf("invalid path or volume not found in path: %s", path)
	}

	// Return the volume name (third part in the split path)
	return fmt.Sprintf("/%s/%s", parts[1], parts[2]), nil
}

// Unix-like volume check using syscall
func isSameVolume(src, dst string) (bool, error) {
	// if the destination is a mounted volume, we cannot use the clone command
	srcVolumeName, err := extractVolumeName(src)
	if err != nil {
		return false, err
	}
	dstVolumeName, err := extractVolumeName(dst)
	if err != nil {
		return false, err
	}

	srcInfo, err := os.Stat(srcVolumeName)
	if err != nil {
		return false, err
	}
	dstInfo, err := os.Stat(dstVolumeName)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	// Use Stat_t to compare the device IDs on Unix systems
	srcDev := srcInfo.Sys().(*syscall.Stat_t).Dev
	dstDev := dstInfo.Sys().(*syscall.Stat_t).Dev
	return srcDev == dstDev, nil
}
