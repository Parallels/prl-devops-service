package catalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
	"github.com/cjlapao/common-go/helper"
)

type CacheFile struct {
	IsDir bool
	Path  string
}

func (s *CatalogManifestService) CleanAllCache(ctx basecontext.ApiContext) error {
	cfg := config.Get()
	cacheLocation, err := cfg.CatalogCacheFolder()
	if err != nil {
		return err
	}

	clearFiles := []CacheFile{}
	entries, err := os.ReadDir(cacheLocation)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		clearFiles = append(clearFiles, CacheFile{
			IsDir: entry.IsDir(),
			Path:  filepath.Join(cacheLocation, entry.Name()),
		})
	}

	for _, file := range clearFiles {
		if file.IsDir {
			if err := os.RemoveAll(file.Path); err != nil {
				return err
			}
		} else {
			if err := os.Remove(file.Path); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *CatalogManifestService) CleanCacheFile(ctx basecontext.ApiContext, catalogId string, version string) error {
	cfg := config.Get()
	cacheLocation, err := cfg.CatalogCacheFolder()
	if err != nil {
		return err
	}

	allCache, err := s.GetCacheItems(ctx)
	if err != nil {
		return err
	}

	found := false
	for _, cache := range allCache.Manifests {
		if strings.EqualFold(cache.CatalogId, catalogId) {
			if strings.EqualFold(cache.Version, version) || version == "" {
				found = true
				if cache.CacheType == "folder" {
					if err := os.RemoveAll(filepath.Join(cacheLocation, cache.CacheFileName)); err != nil {
						return err
					}
				} else {
					if err := os.Remove(cache.CacheLocalFullPath); err != nil {
						return err
					}
				}

				if cache.CacheMetadataName != "" {
					if _, err := os.Stat(filepath.Join(cacheLocation, cache.CacheMetadataName)); os.IsNotExist(err) {
						continue
					}

					if err := os.Remove(filepath.Join(cacheLocation, cache.CacheMetadataName)); err != nil {
						return err
					}
				}
			}
		}
	}

	if !found {
		return errors.NewWithCodef(404, "Cache not found for catalog %s and version %s", catalogId, version)
	}

	return nil
}

func (s *CatalogManifestService) GetCacheItems(ctx basecontext.ApiContext) (models.VirtualMachineCatalogManifestList, error) {
	response := models.VirtualMachineCatalogManifestList{
		Manifests: make([]models.VirtualMachineCatalogManifest, 0),
	}
	cfg := config.Get()
	cacheLocation, err := cfg.CatalogCacheFolder()
	if err != nil {
		return response, err
	}

	totalSize := int64(0)
	err = filepath.Walk(cacheLocation, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".meta" {
			if err := s.processMetadataCache(path, info, &response, &totalSize); err != nil {
				return err
			}
		} else {
			if err := s.processOldCache(ctx, path, info, &response, &totalSize); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return response, err
	}

	response.TotalSize = totalSize
	if response.Manifests == nil {
		response.Manifests = make([]models.VirtualMachineCatalogManifest, 0)
	}

	return response, nil
}

func (s *CatalogManifestService) processMetadataCache(path string, info os.FileInfo, response *models.VirtualMachineCatalogManifestList, totalSize *int64) error {
	var metaContent models.VirtualMachineCatalogManifest
	manifestBytes, err := helper.ReadFromFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(manifestBytes, &metaContent)
	if err != nil {
		return err
	}
	cacheName := strings.TrimSuffix(path, filepath.Ext(path))
	cacheInfo, err := os.Stat(cacheName)
	if err != nil {
		return err
	}
	if cacheInfo.IsDir() {
		metaContent.CacheType = "folder"
		var folderSize int64
		err = filepath.Walk(cacheName, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				folderSize += info.Size()
			}
			return nil
		})
		if err != nil {
			return err
		}
		metaContent.CacheSize = int64(float64(folderSize) / (1024 * 1024))
	} else {
		metaContent.CacheType = "file"
		metaContent.CacheSize = int64(float64(cacheInfo.Size()) / (1024 * 1024))
	}

	metaContent.CacheLocalFullPath = cacheName
	metaContent.CacheFileName = filepath.Base(cacheName)
	metaContent.CacheMetadataName = filepath.Base(path)
	metaContent.CacheDate = info.ModTime().Format("2006-01-02 15:04:05")
	response.Manifests = append(response.Manifests, metaContent)
	*totalSize += metaContent.CacheSize
	return nil
}

func (s *CatalogManifestService) processOldCache(ctx basecontext.ApiContext, path string, info os.FileInfo, response *models.VirtualMachineCatalogManifestList, totalSize *int64) error {
	if filepath.Ext(path) == ".pvm" || filepath.Ext(path) == ".macvm" {
		metaPath := path + ".meta"
		if _, err := os.Stat(metaPath); err == nil {
			return nil
		}

		srvCtl := system.Get()
		arch, err := srvCtl.GetArchitecture(ctx)
		if err != nil {
			arch = "unknown"
		}

		cacheSize := info.Size() / 1024 / 1024
		cacheType := "file"
		if info.IsDir() {
			cacheType = "folder"
			var folderSize int64
			err = filepath.Walk(path, func(p string, i os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !i.IsDir() {
					folderSize += i.Size()
				}
				return nil
			})
			if err != nil {
				return err
			}

			cacheSize = folderSize / 1024 / 1024
		}

		oldCacheManifest := models.VirtualMachineCatalogManifest{
			ID:                 filepath.Base(path),
			CatalogId:          filepath.Base(path),
			Version:            "unknown",
			Architecture:       arch,
			CacheType:          cacheType,
			CacheSize:          cacheSize,
			CacheLocalFullPath: path,
			CacheFileName:      filepath.Base(path),
			CacheMetadataName:  filepath.Base(path),
			CacheDate:          info.ModTime().Format("2006-01-02 15:04:05"),
			IsCompressed:       false,
			Size:               cacheSize,
		}
		if filepath.Ext(path) == ".pvm" {
			oldCacheManifest.Type = "pvm"
		} else {
			oldCacheManifest.Type = "macvm"
		}

		response.Manifests = append(response.Manifests, oldCacheManifest)
		*totalSize += cacheSize
	}
	return nil
}
