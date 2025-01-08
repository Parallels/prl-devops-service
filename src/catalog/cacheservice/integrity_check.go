package cacheservice

import (
	"os"
	"path/filepath"

	"github.com/Parallels/prl-devops-service/errors"
	"github.com/cjlapao/common-go/helper"
)

var requiredFileList = map[string][]string{
	"pvm":   {"config.pvs", "NVRAM.dat"},
	"macvm": {"aux.bin", "config.pvs", "macid.bin", "machw.bin"},
}

func (cs *CacheService) checkCacheItemIntegrity(cachePath string) error {
	metadata, err := cs.loadCacheManifest(cs.cachedMetadataFilePath())
	if err != nil {
		return err
	}

	if !metadata.CacheCompleted {
		return errors.NewWithCode("Cache is not completed", 400)
	}

	switch metadata.Type {
	case "pvm":
		for _, file := range requiredFileList["pvm"] {
			filePath := filepath.Join(cachePath, file)
			if !helper.FileExists(filePath) {
				return errors.NewWithCodef(400, "Cache is not completed, missing file %v ", file)
			}
		}
	case "macvm":
		for _, file := range requiredFileList["macvm"] {
			filePath := filepath.Join(cachePath, file)
			if !helper.FileExists(filePath) {
				return errors.NewWithCodef(400, "Cache is not completed, missing file %v ", file)
			}
		}
	default:
		return errors.NewWithCode("Invalid cache type", 400)
	}

	// Now checking if we have at least one disk file
	foundHDD := false
	err = filepath.Walk(cachePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".hdd" {
			foundHDD = true
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return errors.NewWithCodef(500, "Error traversing cache path: %v", err)
	}

	if !foundHDD {
		return errors.NewWithCode("Cache is not completed, missing .hdd file", 400)
	}

	return nil
}
