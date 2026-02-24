package cacheservice

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/cleanupservice"
	"github.com/Parallels/prl-devops-service/catalog/common"
	"github.com/Parallels/prl-devops-service/catalog/interfaces"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/compressor"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/notifications"
	"github.com/cjlapao/common-go/helper"
)

const (
	DEFAULT_PACKAGE_SIZE = 60 * 1024  // 60 MB in megabytes
	MAX_CACHE_SIZE       = 400 * 1024 // 400 GB in megabytes
	metadataExtension    = ".meta"
	pvmCacheExtension    = ".pvm"
	macvmCacheExtension  = ".macvm"
)

type CacheRequest struct {
	ApiContext           basecontext.ApiContext
	Manifest             *models.VirtualMachineCatalogManifest
	RemoteStorageService interfaces.RemoteStorageService
}

func NewCacheRequest(ctx basecontext.ApiContext, catalogManifest *models.VirtualMachineCatalogManifest, rss interfaces.RemoteStorageService) CacheRequest {
	return CacheRequest{
		ApiContext:           ctx,
		Manifest:             catalogManifest,
		RemoteStorageService: rss,
	}
}

type CacheService struct {
	notificationChannel chan string
	notifications       *notifications.NotificationService
	cfg                 *config.Config
	baseCtx             basecontext.ApiContext
	rss                 interfaces.RemoteStorageService
	manifest            models.VirtualMachineCatalogManifest
	cacheFolder         string
	packFilename        string
	metadataFilename    string
	packChecksum        string
	packFilePath        string
	metadataFilePath    string
	packExtension       string
	metadataExtension   string
	maxCacheSize        int64
	keepFreeDiskSpace   int64
	CacheType           models.CatalogCacheType
	CacheManifest       models.VirtualMachineCatalogManifest
	cacheData           *models.CacheResponse
	cleanupservice      *cleanupservice.CleanupService
}

func NewCacheService(ctx basecontext.ApiContext) (*CacheService, error) {
	svc := &CacheService{
		baseCtx:        ctx,
		cfg:            config.Get(),
		CacheType:      models.CatalogCacheTypeNone,
		cleanupservice: cleanupservice.NewCleanupService(),
		notifications:  notifications.Get(),
	}

	// checking if we have a size for the vm, if not we will set a default size
	if svc.manifest.Size == 0 {
		svc.manifest.Size = DEFAULT_PACKAGE_SIZE
	}
	keepFreeDiskSpace := svc.cfg.GetIntKey(constants.CATALOG_CACHE_KEEP_FREE_DISK_SPACE_ENV_VAR)
	if keepFreeDiskSpace > 0 {
		svc.keepFreeDiskSpace = int64(keepFreeDiskSpace)
	}
	maxCacheSize := svc.cfg.GetIntKey(constants.CATALOG_CACHE_MAX_SIZE_ENV_VAR)
	if maxCacheSize > 0 {
		svc.maxCacheSize = int64(maxCacheSize)
	}

	cacheFolder, err := svc.cfg.CatalogCacheFolder()
	if err != nil {
		err := errors.NewWithCode("Error getting cache folder", 500)
		return nil, err
	}
	svc.cacheFolder = cacheFolder

	return svc, nil
}

func (cs *CacheService) WithRequest(r CacheRequest) error {
	// We will be caching the catalog item

	cs.manifest = *r.Manifest
	cs.rss = r.RemoteStorageService
	cs.baseCtx = r.ApiContext
	cs.packFilename = r.Manifest.PackFile
	cs.metadataFilename = r.Manifest.MetadataFile
	// getting the checksum of the file from the remote storage provider
	if checksum, err := r.RemoteStorageService.FileChecksum(cs.baseCtx, r.Manifest.Path, r.Manifest.PackFile); err != nil {
		err := errors.NewWithCode("Error getting checksum for file", 500)
		return err
	} else {
		cs.packChecksum = checksum
	}

	cs.packFilePath = filepath.Join(r.Manifest.Path, r.Manifest.PackFile)
	cs.metadataFilePath = filepath.Join(r.Manifest.Path, r.Manifest.MetadataFile)
	cs.packExtension = filepath.Ext(r.Manifest.PackFile)
	cs.metadataExtension = filepath.Ext(r.Manifest.MetadataFile)

	return nil
}

func (cs *CacheService) packCacheFilename() string {
	return fmt.Sprintf("%v%v", cs.packChecksum, cs.packExtension)
}

func (cs *CacheService) metadataCacheFileName() string {
	return fmt.Sprintf("%v%v", cs.packChecksum, cs.metadataExtension)
}

func (cs *CacheService) cacheMachineName() string {
	return fmt.Sprintf("%v.%v", cs.packChecksum, cs.manifest.Type)
}

func (cs *CacheService) cachedPackFilePath() string {
	return filepath.Join(cs.cacheFolder, cs.packCacheFilename())
}

func (cs *CacheService) cachedMetadataFilePath() string {
	return filepath.Join(cs.cacheFolder, cs.metadataCacheFileName())
}

func (cs *CacheService) notify(message string) {
	cs.notifications.NotifyInfo(message)
}

func (cs *CacheService) getCacheTotalSize() (int64, error) {
	// We will be getting the total size of the cache folder
	totalSize, err := helpers.DirSize(cs.cacheFolder)
	if err != nil {
		return -1, errors.NewFromErrorWithCode(err, 500)
	}

	return totalSize, nil
}

func (cs *CacheService) checkNeedCleanup() (*CleanupRequirements, error) {
	r := CleanupRequirements{
		NeedsCleaning: false,
		Reason:        "",
	}

	freeDiskSpace, err := cs.getFreeDiskSpace()
	if err != nil {
		return nil, errors.NewFromErrorWithCode(err, 500)
	}
	cacheTotalSize, err := cs.getCacheTotalSize()
	if err != nil {
		return nil, errors.NewFromErrorWithCode(err, 500)
	}

	manifestUsedSize := cs.manifest.Size * 2
	// if we can stream then we do no need the double size request
	if cs.rss.CanStream() && cs.cfg.IsRemoteProviderStreamEnabled() {
		manifestUsedSize = cs.manifest.Size
	}

	// First lets check if we passed the setup thresholds in the system
	if cs.keepFreeDiskSpace > 0 {
		// if we have less free space than what we want to keep, we will return true
		if freeDiskSpace < (cs.keepFreeDiskSpace + manifestUsedSize) {
			r.NeedsCleaning = true
			r.Reason = "Free disk space with the new cached item is less than the keep free disk space threshold"
			r.SpaceNeeded = (cs.keepFreeDiskSpace + manifestUsedSize) - freeDiskSpace
			return &r, nil
		}
		// if the total cache size is bigger than the keep free disk space, we will return true
		if cacheTotalSize > (cs.maxCacheSize + manifestUsedSize) {
			r.NeedsCleaning = true
			r.Reason = "Cache size with the new cached item is bigger than the keep free disk space"
			r.SpaceNeeded = (cs.maxCacheSize + manifestUsedSize) - cacheTotalSize
			return &r, nil
		}
		// if the total cache size plus the current package size is bigger than the keep free disk space, we will return true
		if cacheTotalSize+manifestUsedSize > cs.keepFreeDiskSpace {
			r.NeedsCleaning = true
			r.Reason = "Cache size including the cached item is bigger than the keep free disk space"
			r.SpaceNeeded = cs.keepFreeDiskSpace - (cacheTotalSize + manifestUsedSize)
			return &r, nil
		}
	}
	// We will now check if we passed the max cache size threshold
	if cs.maxCacheSize > 0 {
		// if the total cache size is bigger than the max cache size, we will return true
		if cacheTotalSize > (cs.maxCacheSize + manifestUsedSize) {
			r.NeedsCleaning = true
			r.Reason = "Cache size with the new cached item is bigger than the keep free disk space"
			r.SpaceNeeded = (cs.maxCacheSize + manifestUsedSize) - cacheTotalSize
			return &r, nil
		}
		// if the total cache size plus the current package size is bigger than the max cache size, we will return true
		if cacheTotalSize+manifestUsedSize > cs.maxCacheSize {
			r.NeedsCleaning = true
			r.Reason = "Cache size including the cached item is bigger than the keep free disk space"
			r.SpaceNeeded = cs.maxCacheSize - (cacheTotalSize + manifestUsedSize)
			return &r, nil
		}
	}

	// lastly we will check if we indeed have space to cache the package
	if manifestUsedSize > (freeDiskSpace + cacheTotalSize) {
		r.NeedsCleaning = true
		r.Reason = "Free disk space is less than required to cache the package"
		r.SpaceNeeded = manifestUsedSize - (freeDiskSpace + cacheTotalSize)
		r.IsFatal = true
		return &r, nil
	}

	return &r, nil
}

func (cs *CacheService) loadCacheManifest(metadataPath string) (*models.VirtualMachineCatalogManifest, error) {
	// We will be loading the cache manifest file from the disk
	if metadataPath == "" {
		return nil, errors.NewWithCode("Metadata file path is empty", 500)
	}
	if !helper.FileExists(metadataPath) {
		return nil, errors.NewWithCode("Metadata file does not exist", 404)
	}

	content, err := helper.ReadFromFile(metadataPath)
	if err != nil {
		return nil, errors.NewWithCode("Error reading metadata file", 500)
	}

	var r models.VirtualMachineCatalogManifest
	if err := json.Unmarshal(content, &r); err != nil {
		return nil, errors.NewWithCode("Error unmarshalling metadata file", 500)
	}

	return &r, nil
}

func (cs *CacheService) saveCacheManifest(cacheManifest models.VirtualMachineCatalogManifest, metadataPath string) error {
	// We will be saving the cache manifest file to the disk
	if cacheManifest.CacheLocalFullPath == "" {
		return errors.NewWithCode("CacheLocalFullPath is empty", 500)
	}

	if !helper.FileExists(cacheManifest.CacheLocalFullPath) {
		return errors.NewWithCode("CacheLocalFullPath does not exist", 404)
	}

	content, err := json.Marshal(cacheManifest)
	if err != nil {
		return errors.NewWithCode("Error marshalling cache manifest", 500)
	}

	if err := helper.WriteToFile(string(content), metadataPath); err != nil {
		return errors.NewWithCode("Error writing metadata file", 500)
	}

	return nil
}

func (cs *CacheService) updateCacheManifest(metadataPath string) (*models.VirtualMachineCatalogManifest, error) {
	baseDir := filepath.Dir(metadataPath)
	fileName := filepath.Base(metadataPath)
	extension := filepath.Ext(metadataPath)
	if extension != metadataExtension {
		return nil, errors.NewWithCode("Invalid metadata file extension", 500)
	}

	name := strings.TrimSuffix(fileName, extension)

	metadata, err := cs.loadCacheManifest(metadataPath)
	if err != nil {
		return nil, errors.NewFromErrorWithCode(err, 500)
	}

	packFileName := fmt.Sprintf("%v.%v", name, metadata.Type)
	packFilePath := filepath.Join(baseDir, packFileName)
	if !helper.FileExists(packFilePath) {
		return nil, nil
	}

	if metadata.CachedDate == "" {
		metadata.CachedDate = time.Now().Format(time.RFC3339)
	}
	if metadata.CacheLastUsed == "" {
		metadata.CacheLastUsed = time.Unix(0, 0).Format(time.RFC3339)
	}
	metadata.CacheLocalFullPath = baseDir
	metadata.CacheMetadataName = fmt.Sprintf("%v%v", name, extension)
	metadata.CacheFileName = fmt.Sprintf("%v.%v", name, metadata.Type)
	metadata.IsCompressed = false

	// updating the cache type
	cacheInfo, err := os.Stat(packFilePath)
	if err != nil {
		return nil, err
	}

	if cacheInfo.IsDir() {
		metadata.CacheType = "folder"
	} else {
		metadata.CacheType = "file"
	}

	// updating the cache size
	cacheSize, err := cs.getCacheSize(metadata)
	if err != nil {
		return nil, errors.NewFromErrorWithCode(err, 500)
	}
	metadata.CacheSize = cacheSize

	if err := cs.saveCacheManifest(*metadata, metadataPath); err != nil {
		return nil, errors.NewFromErrorWithCode(err, 500)
	}

	return metadata, nil
}

func (cs *CacheService) updateCacheManifestUsedCount(metadataPath string) (*models.VirtualMachineCatalogManifest, error) {
	metadata, err := cs.loadCacheManifest(metadataPath)
	if err != nil {
		return nil, errors.NewFromErrorWithCode(err, 500)
	}

	metadata.CacheLastUsed = time.Now().Format(time.RFC3339)
	metadata.CacheUsedCount++

	cs.CacheManifest = *metadata
	if err := cs.saveCacheManifest(*metadata, metadataPath); err != nil {
		return nil, errors.NewFromErrorWithCode(err, 500)
	}

	return metadata, nil
}

func (cs *CacheService) setCacheCompleted(metadataPath string) (*models.VirtualMachineCatalogManifest, error) {
	metadata, err := cs.loadCacheManifest(metadataPath)
	if err != nil {
		return nil, errors.NewFromErrorWithCode(err, 500)
	}

	metadata.CacheCompleted = true

	cs.CacheManifest = *metadata
	if err := cs.saveCacheManifest(*metadata, metadataPath); err != nil {
		return nil, errors.NewFromErrorWithCode(err, 500)
	}

	return metadata, nil
}

func (cs *CacheService) getCacheSize(metadata *models.VirtualMachineCatalogManifest) (int64, error) {
	path := filepath.Join(metadata.CacheLocalFullPath, metadata.CacheFileName)
	switch metadata.CacheType {
	case models.CatalogCacheTypeFile.String():
		if info, err := os.Stat(path); err == nil {
			size := info.Size() / (1024 * 1024) // size in MB
			return size, nil
		} else {
			return -1, errors.NewFromErrorWithCode(err, 500)
		}
	case models.CatalogCacheTypeFolder.String():
		folderSizeInMB, err := helpers.DirSize(path)
		if err != nil {
			return -1, errors.NewFromErrorWithCode(err, 500)
		}
		return folderSizeInMB, nil
	case models.CatalogCacheTypeNone.String():
		return -1, errors.NewWithCode("Cache type is none", 500)
	default:
		return -1, errors.NewWithCode("Unknown cache type", 500)
	}
}

func (cs *CacheService) processMetadataCacheItemFile(item CacheItemFile, cleanerSvc *cleanupservice.CleanupService) (*models.VirtualMachineCatalogManifest, error) {
	if item.NeedsCleaning() {
		for _, file := range item.InvalidFiles {
			isFolder := false
			if info, err := os.Stat(file); err == nil && info.IsDir() {
				isFolder = true
			}
			cleanerSvc.AddLocalFileCleanupOperation(file, isFolder)
		}
	}
	if item.NeedsRenaming {
		if item.MetadataFileName != "" {
			newMetadataFileName := filepath.Join(cs.cacheFolder, fmt.Sprintf("%v%v", item.BaseName, metadataExtension))
			if err := os.Rename(item.MetadataFileName, newMetadataFileName); err != nil {
				return nil, err
			}
			item.MetadataFileName = newMetadataFileName
		}
	}
	if item.IsValid() {
		manifest, err := cs.updateCacheManifest(item.MetadataFileName)
		if err != nil {
			return nil, err
		}
		if manifest != nil {
			return manifest, nil
		}
	}
	// This will deal with dangling cache files from older versions as these did not have any caching metadata and
	// are no longer needed as they are not valid
	if item.MetadataFileName == "" && item.CacheFileName != "" {
		cleanerSvc.AddLocalFileCleanupOperation(item.CacheFileName, item.IsCachedFolder)
	}

	return nil, nil
}

func (cs *CacheService) processMetadataCacheFolderItem() (map[string]CacheItemFile, error) {
	cachedContent := make(map[string]CacheItemFile)

	files, err := os.ReadDir(cs.cacheFolder)
	if err != nil {
		return nil, errors.NewFromErrorWithCodef(err, 500, "there was an error checking the cache folder")
	}

	for _, file := range files {
		path := filepath.Join(cs.cacheFolder, file.Name())
		fileParts := strings.Split(file.Name(), ".")
		baseFilename := fileParts[0]
		extension := fmt.Sprintf(".%v", strings.Join(fileParts[1:], "."))
		var cachedItem CacheItemFile
		if _, ok := cachedContent[baseFilename]; ok {
			cachedItem = cachedContent[baseFilename]
		} else {
			cachedItem = CacheItemFile{
				MetadataFileName: "",
				BaseName:         baseFilename,
				CacheFileName:    "",
			}
		}

		if strings.EqualFold(extension, fmt.Sprintf("%v%v", pvmCacheExtension, metadataExtension)) || strings.EqualFold(extension, fmt.Sprintf("%v%v", macvmCacheExtension, metadataExtension)) {
			cachedItem.MetadataFileName = path
			cachedItem.NeedsRenaming = true
		} else if strings.EqualFold(extension, metadataExtension) {
			cachedItem.MetadataFileName = path
		} else if strings.EqualFold(extension, pvmCacheExtension) || strings.EqualFold(extension, macvmCacheExtension) {
			if !file.IsDir() {
				cachedItem.IsCompressed = true
			} else {
				cachedItem.IsCachedFolder = true
			}
			cachedItem.CacheFileName = path
		} else {
			cachedItem.InvalidFiles = append(cachedItem.InvalidFiles, path)
		}
		cachedContent[baseFilename] = cachedItem
	}

	return cachedContent, nil
}

func (cs *CacheService) processCacheFileWithStream() (string, error) {
	destinationFolder := filepath.Join(cs.cacheFolder, fmt.Sprintf("%v.%v", cs.packChecksum, cs.manifest.Type))
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("temp-%v", cs.packChecksum))
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	if err := cs.rss.PullFileAndDecompress(cs.baseCtx, cs.manifest.Path, cs.manifest.PackFile, tempDir); err != nil {
		return destinationFolder, err
	}

	// Now we move the folder to the destination folder
	if err := os.Rename(tempDir, destinationFolder); err != nil {
		// if the rename fails we might have a case where we are trying to rename a folder to a folder that already exists
		// so we will try to move the content of the folder to the destination folder
		if err := helpers.CopyDir(tempDir, destinationFolder); err != nil {
			return destinationFolder, err
		}
	}

	cs.cleanupservice.AddLocalFileCleanupOperation(destinationFolder, true)
	return destinationFolder, nil
}

func (cs *CacheService) processCacheFileWithoutStream() (string, error) {
	destinationFolder := filepath.Join(cs.cacheFolder, fmt.Sprintf("%v.%v", cs.packChecksum, cs.manifest.Type))
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("temp-%v", cs.packChecksum))
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	destinationFile := filepath.Join(tempDir, cs.manifest.PackFile)

	if err := cs.rss.PullFile(cs.baseCtx, cs.manifest.Path, cs.manifest.PackFile, tempDir); err != nil {
		cs.cleanupservice.Clean(cs.baseCtx)
		return destinationFolder, err
	}
	// checking if the pack file is compressed or not if it is we will decompress it to the destination folder
	// and remove the pack file from the cache folder if not we will just rename the pack file to the checksum
	if cs.manifest.IsCompressed || strings.HasSuffix(cs.manifest.PackFile, ".pdpack") {
		if err := compressor.DecompressFile(cs.baseCtx, destinationFile, tempDir); err != nil {
			return destinationFolder, err
		}

		// removing the compressed file
		if err := os.Remove(destinationFile); err != nil {
			return destinationFolder, err
		}

		// moving the folder to the destination folder
		if err := os.Rename(tempDir, destinationFolder); err != nil {
			// if the rename fails we might have a case where we are trying to rename a folder to a folder that already exists
			// so we will try to move the content of the folder to the destination folder
			if err := helpers.CopyDir(tempDir, destinationFolder); err != nil {
				return destinationFolder, err
			}
		}
		cs.cleanupservice.AddLocalFileCleanupOperation(destinationFolder, true)
	} else {
		// rename the pack to the checksum
		if err := os.Rename(destinationFile, cs.cachedPackFilePath()); err != nil {
			// if the rename fails we might have a case where we are trying to rename a file to a file that already exists
			// so we will try to move the content of the file to the destination file
			if err := helpers.CopyFile(destinationFile, cs.cachedPackFilePath()); err != nil {
				cs.cleanupservice.Clean(cs.baseCtx)
				return destinationFolder, err
			}
		}
		cs.cleanupservice.AddLocalFileCleanupOperation(cs.cachedPackFilePath(), false)
	}

	return destinationFolder, nil
}

func (cs *CacheService) GetAllCacheItems() (models.CachedManifests, error) {
	response := models.CachedManifests{
		Manifests: make([]models.VirtualMachineCatalogManifest, 0),
	}
	cleanerSvc := cleanupservice.NewCleanupService()
	if cleanerSvc == nil {
		return response, errors.NewWithCode("Error creating cleanup service", 500)
	}

	// Getting all the cache items from the cache folder and updating the cache manifest
	totalSize := int64(0)
	cachedContent, err := cs.processMetadataCacheFolderItem()
	if err != nil {
		return response, errors.NewFromErrorWithCodef(err, 500, "Error processing metadata cache folder")
	}

	// Processing the cache items and doing some potential cleanup
	for _, item := range cachedContent {
		manifest, err := cs.processMetadataCacheItemFile(item, cleanerSvc)
		if err != nil {
			return response, errors.NewFromErrorWithCodef(err, 500, "Error processing metadata cache file")
		}
		if manifest != nil {
			response.Manifests = append(response.Manifests, *manifest)
			totalSize += manifest.CacheSize
		}
	}

	// Generating the response
	response.TotalSize = totalSize
	if response.Manifests == nil {
		response.Manifests = make([]models.VirtualMachineCatalogManifest, 0)
	}

	response.SortManifestsByCachedDate()
	// Executing the cleanup process if any
	cleanerSvc.Clean(cs.baseCtx)
	return response, nil
}

func (cs *CacheService) IsCached() bool {
	if cs.cacheData == nil {
		cs.Get()
	}

	if cs.cacheData == nil {
		return false
	}

	if cs.cacheData.IsCached {
		return true
	}

	return false
}

func (cs *CacheService) RemoveAllCacheItems() error {
	cacheItems, err := cs.GetAllCacheItems()
	if err != nil {
		return err
	}

	for _, cache := range cacheItems.Manifests {
		if err := cs.RemoveCacheItem(cache.CatalogId, cache.Version); err != nil {
			return err
		}
	}

	return nil
}

func (cs *CacheService) RemoveCacheItem(catalogId string, version string) error {
	// We will be removing the cache item from the cache folder
	cacheItems, err := cs.GetAllCacheItems()
	if err != nil {
		return err
	}

	if catalogId == "" {
		return errors.NewWithCode("Catalog ID is empty", 500)
	}

	// If the version is empty we will set it to ALL so we can clean all versions
	if version == "" {
		version = "ALL"
	}

	found := false
	for _, cache := range cacheItems.Manifests {
		if strings.EqualFold(cache.CatalogId, catalogId) {
			if strings.EqualFold(cache.Version, version) || version == "ALL" {
				found = true
				packFilePath := filepath.Join(cache.CacheLocalFullPath, cache.CacheFileName)
				metadataFullPath := filepath.Join(cache.CacheLocalFullPath, cache.CacheMetadataName)
				if cache.CacheType == models.CatalogCacheTypeFolder.String() {
					if err := helper.DeleteAllFiles(packFilePath); err != nil {
						return err
					}
					if err := os.Remove(packFilePath); err != nil {
						return err
					}
				} else {
					if err := os.Remove(packFilePath); err != nil {
						return err
					}
				}

				if cache.CacheMetadataName != "" {
					if _, err := os.Stat(metadataFullPath); os.IsNotExist(err) {
						continue
					}

					if err := os.Remove((metadataFullPath)); err != nil {
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

func (cs *CacheService) Get() (*models.CacheResponse, error) {
	// Returning false if the cache is disabled so we can force pull the catalog
	if !cs.cfg.IsCatalogCachingEnable() {
		r := models.CacheResponse{
			Type:             models.CatalogCacheTypeNone,
			IsCached:         false,
			MetadataFilePath: cs.metadataFilePath,
			PackFilePath:     cs.packFilePath,
			Checksum:         cs.packChecksum,
		}

		cs.cacheData = &r
		return cs.cacheData, nil
	}

	r := models.CacheResponse{}

	metadataCacheFilePath := filepath.Join(cs.cacheFolder, cs.metadataCacheFileName())
	packCacheFilePath := filepath.Join(cs.cacheFolder, cs.packCacheFilename())
	machineCacheFilePath := filepath.Join(cs.cacheFolder, cs.cacheMachineName())

	// checking if we can find the cache metadata file in the cache folder
	if helper.FileExists(metadataCacheFilePath) {
		cs.baseCtx.LogDebugf("Metadata file %v already exists in cache", cs.metadataFilename)
		r.MetadataFilePath = metadataCacheFilePath
	}

	// checking if the pack file is in the cache folder
	if helper.FileExists(packCacheFilePath) {
		cs.baseCtx.LogDebugf("Compressed File %v already exists in cache", cs.packCacheFilename())
		if info, err := os.Stat(packCacheFilePath); err == nil && info.IsDir() {
			cs.baseCtx.LogDebugf("Cache file %v is a directory, treating it as a folder", cs.packCacheFilename())
			r.Type = models.CatalogCacheTypeFolder
			// Checking the integrity of the cache file
			if err := cs.checkCacheItemIntegrity(packCacheFilePath); err != nil {
				cs.RemoveCacheItem(cs.manifest.CatalogId, cs.manifest.Version)
				r.IsCached = false
				cs.cacheData = &r
				return cs.cacheData, err
			}
		} else {
			r.Type = models.CatalogCacheTypeFile
		}
		r.PackFilePath = packCacheFilePath
	} else if helper.FileExists(machineCacheFilePath) {
		cs.baseCtx.LogDebugf("Machine Folder %v already exists in cache", cs.cacheMachineName())
		if info, err := os.Stat(machineCacheFilePath); err == nil && info.IsDir() {
			cs.baseCtx.LogDebugf("Cache file %v is a directory, treating it as a folder", cs.cacheMachineName())
			r.Type = models.CatalogCacheTypeFolder

			// Checking the integrity of the cache file
			if err := cs.checkCacheItemIntegrity(machineCacheFilePath); err != nil {
				cs.RemoveCacheItem(cs.manifest.CatalogId, cs.manifest.Version)
				r.IsCached = false
				cs.cacheData = &r
				return cs.cacheData, err
			}
		} else {
			r.Type = models.CatalogCacheTypeFile
		}
		r.PackFilePath = machineCacheFilePath
	}

	if r.MetadataFilePath != "" && r.PackFilePath != "" {
		r.IsCached = true
	}

	cs.cacheData = &r
	return cs.cacheData, nil
}

func (cs *CacheService) UpdateCacheManifest() error {
	// We will be updating the cache manifest file to set the last used and used_count fields
	return nil
}

func (cs *CacheService) Clean() error {
	cleanupRequirement, err := cs.checkNeedCleanup()
	if err != nil {
		cs.cleanupservice.Clean(cs.baseCtx)
		return err
	}
	// Checking if we have any fatal requirements in the cleanup as that means we
	// cannot continue with the process
	if cleanupRequirement.IsFatal {
		return errors.NewWithCodef(500, "Fatal cleanup requirement: %v", cleanupRequirement.Reason)
	}

	// We have enough space, no need to cleanup the cache
	// We will return nil to allow the process to continue
	if !cleanupRequirement.NeedsCleaning {
		cs.notify("No cleanup needed, we have enough space")
		return nil
	}
	cs.baseCtx.LogDebugf("Cleanup needed, reason: %v", cleanupRequirement.Reason)

	// We need to do some cleanup to get enough space
	// The first step is to get all of the current cache items, even old ones
	cacheItems, err := cs.GetAllCacheItems()
	if err != nil {
		return err
	}

	// We will sort the cache items by ranking so we can delete the least used ones first
	cacheItems.SortManifestsByRanking()

	// we will now calculate the amount of items we need to delete based on the space needed
	itemToRemove := []models.VirtualMachineCatalogManifest{}
	for _, item := range cacheItems.Manifests {
		if cleanupRequirement.SpaceNeeded <= 0 {
			break
		}
		itemToRemove = append(itemToRemove, item)
		cleanupRequirement.SpaceNeeded -= item.CacheSize
	}

	// If we would clear all the cache items, and we still need space, we
	if cleanupRequirement.SpaceNeeded > 0 {
		allowedAboveFreeDiskSpace := cs.cfg.GetBoolKey(constants.CATALOG_CACHE_ALLOW_CACHE_ABOVE_FREE_DISK_SPACE_ENV_VAR)
		if !allowedAboveFreeDiskSpace {
			cs.notify("Not enough space for the cached item even after cleaning the cache due to required free disk space rule, set the override flag to allow cache above free disk space")
			return errors.NewWithCodef(500, "Not enough space for the cached item even after cleaning the cache due to required free disk space rule, set the override flag to allow cache above free disk space")
		}
	}

	for _, item := range itemToRemove {
		cs.notify(fmt.Sprintf("Removing cache item %v", item.Name))
		if err := cs.RemoveCacheItem(item.CatalogId, item.Version); err != nil {
			return err
		}
	}

	return nil
}

func (cs *CacheService) Cache() error {
	// First let's get the catalog manifest file from the remote storage into memory as the local cache will be a different file based on this
	cs.notify("Downloading catalog manifest file")

	if err := cs.rss.PullFile(cs.baseCtx, cs.manifest.Path, cs.manifest.MetadataFile, cs.cacheFolder); err != nil {
		return err
	}
	cs.cleanupservice.AddLocalFileCleanupOperation(filepath.Join(cs.cacheFolder, cs.manifest.MetadataFile), false)

	// rename the metadata to the checksum
	if err := os.Rename(filepath.Join(cs.cacheFolder, cs.manifest.MetadataFile), cs.cachedMetadataFilePath()); err != nil {
		cs.cleanupservice.Clean(cs.baseCtx)
		return err
	}

	cs.cleanupservice.AddLocalFileCleanupOperation(cs.cachedMetadataFilePath(), false)
	cs.cleanupservice.RemoveLocalFileCleanupOperation(filepath.Join(cs.cacheFolder, cs.manifest.MetadataFile))

	if err := cs.Clean(); err != nil {
		cs.cleanupservice.Clean(cs.baseCtx)
		return errors.NewFromErrorWithCodef(err, 500, "Error cleaning cache")
	}

	cs.notify("Downloading catalog pack file")

	// Checking if the cached file is compressed or not and if we can stream the file and decompress on the fly
	// if not we will need to process this the old way, pulling the file first and then decompressing it
	if (cs.manifest.IsCompressed || strings.HasSuffix(cs.manifest.PackFile, ".pdpack")) && cs.rss.CanStream() {
		destinationFolder, err := cs.processCacheFileWithStream()
		if err != nil {
			cs.cleanupservice.Clean(cs.baseCtx)
			return err
		}
		cs.cleanupservice.AddLocalFileCleanupOperation(destinationFolder, true)
		if err := common.CleanAndFlatten(destinationFolder); err != nil {
			cs.cleanupservice.Clean(cs.baseCtx)
			return err
		}
	} else {
		destinationFolder, err := cs.processCacheFileWithoutStream()
		if err != nil {
			cs.cleanupservice.Clean(cs.baseCtx)
			return err
		}
		cs.cleanupservice.AddLocalFileCleanupOperation(destinationFolder, true)
		if err := common.CleanAndFlatten(destinationFolder); err != nil {
			cs.cleanupservice.Clean(cs.baseCtx)
			return err
		}
	}

	// we will need to update the cache manifest file first so we can check if indeed we can cache the package
	cs.notify("Updating cache manifest")
	_, err := cs.updateCacheManifest(cs.cachedMetadataFilePath())
	if err != nil {
		cs.cleanupservice.Clean(cs.baseCtx)
		return err
	}

	// Setting the cache item as completed for integrity checks
	manifest, err := cs.setCacheCompleted(cs.cachedMetadataFilePath())
	if err != nil {
		cs.cleanupservice.Clean(cs.baseCtx)
		return err
	}

	if manifest == nil {
		cs.cleanupservice.Clean(cs.baseCtx)
		return errors.NewWithCode("Error updating cache manifest", 500)
	}

	cacheItemFolder := filepath.Join(manifest.CacheLocalFullPath, manifest.CacheFileName)
	if err := cs.checkCacheItemIntegrity(cacheItemFolder); err != nil {
		cs.cleanupservice.Clean(cs.baseCtx)
		return err
	}

	cs.CacheManifest = *manifest
	switch cs.CacheManifest.CacheType {
	case models.CatalogCacheTypeFile.String():
		cs.CacheType = models.CatalogCacheTypeFile
	case models.CatalogCacheTypeFolder.String():
		cs.CacheType = models.CatalogCacheTypeFolder
	default:
		cs.CacheType = models.CatalogCacheTypeNone
	}
	return nil
}
