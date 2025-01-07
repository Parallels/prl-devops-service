//go:build linux || darwin
// +build linux darwin

package cacheservice

import (
	"syscall"

	"github.com/Parallels/prl-devops-service/errors"
)

func (cs *CacheService) getFreeDiskSpace() (int64, error) {
	// We will be getting the free disk space from the disk
	var stat syscall.Statfs_t
	err := syscall.Statfs(cs.cacheFolder, &stat)
	if err != nil {
		return 0, errors.NewFromErrorWithCode(err, 500)
	}

	// Available blocks * size per block = available space in bytes
	diskFreeSpaceInBytes := int64(stat.Bavail) * int64(stat.Bsize)
	diskFreeSpaceInMb := diskFreeSpaceInBytes / (1024 * 1024)

	return diskFreeSpaceInMb, nil
}
