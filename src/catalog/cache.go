package catalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/cjlapao/common-go/helper"
)

func (s *CatalogManifestService) CleanAllCache(ctx basecontext.ApiContext) error {
	cfg := config.Get()
	cacheLocation, err := cfg.CatalogCacheFolder()
	if err != nil {
		return err
	}

	err = filepath.Walk(cacheLocation, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return os.Remove(path)
		}
		return os.RemoveAll(path)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *CatalogManifestService) CleanCacheFile(ctx basecontext.ApiContext, catalogId string) error {
	cfg := config.Get()
	cacheLocation, err := cfg.CatalogCacheFolder()
	if err != nil {
		return err
	}

	allCache, err := s.GetCacheItems(ctx)
	if err != nil {
		return err
	}

	for _, cache := range allCache.Manifests {
		if strings.EqualFold(cache.CatalogId, catalogId) {
			if cache.CacheType == "folder" {
				if err := os.RemoveAll(filepath.Join(cacheLocation, cache.CacheFileName)); err != nil {
					return err
				}
			} else {
				if err := os.Remove(cache.CacheLocalFullPath); err != nil {
					return err
				}
			}

			if err := os.Remove(filepath.Join(cacheLocation, cache.CacheMetadataName)); err != nil {
				return err
			}
		}
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
				var totalSize int64
				err = filepath.Walk(cacheName, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() {
						totalSize += info.Size()
					}
					return nil
				})
				if err != nil {
					return err
				}
				metaContent.CacheSize = totalSize
			} else {
				metaContent.CacheType = "file"
				metaContent.CacheSize = cacheInfo.Size()
			}

			metaContent.CacheLocalFullPath = cacheName
			metaContent.CacheFileName = filepath.Base(cacheName)
			metaContent.CacheMetadataName = filepath.Base(path)
			metaContent.CacheDate = info.ModTime().Format("2006-01-02 15:04:05")
			response.Manifests = append(response.Manifests, metaContent)
			totalSize += metaContent.CacheSize
		}
		return nil
	})
	if err != nil {
		return response, err
	}

	response.TotalSize = totalSize
	return response, nil
}
