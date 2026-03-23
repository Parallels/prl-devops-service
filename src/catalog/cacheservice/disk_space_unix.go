//go:build linux || darwin
// +build linux darwin

package cacheservice

import (
	"github.com/Parallels/prl-devops-service/helpers"
)

func (cs *CacheService) getFreeDiskSpace() (int64, error) {
	return helpers.GetFreeDiskSpace(cs.cacheFolder)
}
