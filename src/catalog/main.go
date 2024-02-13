package catalog

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog/cleanupservice"
	"github.com/Parallels/pd-api-service/catalog/interfaces"
	"github.com/Parallels/pd-api-service/catalog/models"
	"github.com/Parallels/pd-api-service/catalog/providers/artifactory"
	"github.com/Parallels/pd-api-service/catalog/providers/aws_s3_bucket"
	"github.com/Parallels/pd-api-service/catalog/providers/azurestorageaccount"
	"github.com/Parallels/pd-api-service/catalog/providers/local"
	"github.com/Parallels/pd-api-service/helpers"

	"github.com/cjlapao/common-go/helper"
)

type CatalogManifestService struct {
	remoteServices []interfaces.RemoteStorageService
}

func NewManifestService(ctx basecontext.ApiContext) *CatalogManifestService {
	manifestService := &CatalogManifestService{}
	manifestService.remoteServices = make([]interfaces.RemoteStorageService, 0)
	manifestService.AddRemoteService(aws_s3_bucket.NewAwsS3Provider())
	manifestService.AddRemoteService(local.NewLocalProvider())
	manifestService.AddRemoteService(azurestorageaccount.NewAzureStorageAccountProvider())
	manifestService.AddRemoteService(artifactory.NewArtifactoryProvider())
	return manifestService
}

func (s *CatalogManifestService) GetProviders(ctx basecontext.ApiContext) []interfaces.RemoteStorageService {
	return s.remoteServices
}

func (s *CatalogManifestService) AddRemoteService(service interfaces.RemoteStorageService) {
	exists := false
	for _, remoteService := range s.remoteServices {
		if remoteService.Name() == service.Name() {
			exists = true
			break
		}
	}

	if exists {
		return
	}

	s.remoteServices = append(s.remoteServices, service)
}

func (s *CatalogManifestService) GenerateManifestContent(ctx basecontext.ApiContext, r *models.PushCatalogManifestRequest, manifest *models.VirtualMachineCatalogManifest) error {
	ctx.LogInfof("Generating manifest content for %v", r.CatalogId)
	if manifest == nil {
		manifest = models.NewVirtualMachineCatalogManifest()
	}

	manifest.CleanupRequest = cleanupservice.NewCleanupRequest()
	manifest.CreatedAt = helpers.GetUtcCurrentDateTime()
	manifest.UpdatedAt = helpers.GetUtcCurrentDateTime()

	manifest.Name = r.CatalogId
	manifest.Path = r.LocalPath
	manifest.Architecture = r.Architecture
	manifest.ID = helpers.GenerateId()
	manifest.CatalogId = helpers.NormalizeString(r.CatalogId)
	manifest.Description = r.Description
	manifest.Version = helpers.NormalizeString(r.Version)
	manifest.Name = fmt.Sprintf("%v-%v-%v", manifest.CatalogId, manifest.Architecture, manifest.Version)
	manifestPackFileName := s.getPackFilename(manifest.Name)

	if r.RequiredRoles != nil {
		manifest.RequiredRoles = r.RequiredRoles
	}
	if r.RequiredClaims != nil {
		manifest.RequiredClaims = r.RequiredClaims
	}
	if r.Tags != nil {
		manifest.Tags = r.Tags
	}

	_, file := filepath.Split(r.LocalPath)
	ext := filepath.Ext(file)
	manifest.Type = ext[1:]

	isDir, err := helpers.IsDirectory(r.LocalPath)
	if err != nil {
		return err
	}
	if !isDir {
		return fmt.Errorf("the path %v is not a directory", r.LocalPath)
	}

	ctx.LogInfof("Getting manifest files for %v", r.CatalogId)
	files, err := s.getManifestFiles(r.LocalPath, "")
	if err != nil {
		return err
	}

	ctx.LogInfof("Compressing manifest files for %v", r.CatalogId)
	packFilePath, err := s.compressMachine(ctx, r.LocalPath, manifestPackFileName, "/tmp")
	if err != nil {
		return err
	}

	// Adding the zip file to the cleanup request
	manifest.CleanupRequest.AddLocalFileCleanupOperation(packFilePath, false)
	manifest.CompressedPath = packFilePath
	manifest.PackFile = "/tmp/" + manifestPackFileName

	fileInfo, err := os.Stat(packFilePath)
	if err != nil {
		return err
	}

	manifest.Size = fileInfo.Size()

	ctx.LogInfof("Getting manifest package checksum for %v", r.CatalogId)
	checksum, err := helpers.GetFileMD5Checksum(packFilePath)
	if err != nil {
		return err
	}
	manifest.CompressedChecksum = checksum

	manifest.VirtualMachineContents = files
	ctx.LogInfof("Finished generating manifest content for %v", r.CatalogId)
	return nil
}

func (s *CatalogManifestService) getManifestFiles(path string, relativePath string) ([]models.VirtualMachineManifestContentItem, error) {
	if relativePath == "" {
		relativePath = "/"
	}

	result := make([]models.VirtualMachineManifestContentItem, 0)
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		// fullPath := filepath.Join(path, file.Name())
		if file.IsDir() {
			result = append(result, models.VirtualMachineManifestContentItem{
				IsDir: true,
				Name:  file.Name(),
				Path:  relativePath,
			})
			files, err := s.getManifestFiles(filepath.Join(path, file.Name()), filepath.Join(relativePath, file.Name()))
			if err != nil {
				return nil, err
			}
			result = append(result, files...)
			continue
		}

		manifestFile := models.VirtualMachineManifestContentItem{
			Path: relativePath,
		}
		fileInfo, err := file.Info()
		if err != nil {
			return nil, err
		}
		manifestFile.Name = file.Name()
		manifestFile.Size = fileInfo.Size()
		manifestFile.CreatedAt = fileInfo.ModTime().Format(time.RFC3339Nano)
		manifestFile.UpdatedAt = fileInfo.ModTime().Format(time.RFC3339Nano)
		// checksum, err := helpers.GetFileMD5Checksum(fullPath)
		// if err != nil {
		// 	return nil, err
		// }
		// manifestFile.Checksum = checksum
		result = append(result, manifestFile)
	}

	return result, nil
}

func (s *CatalogManifestService) readManifestFromFile(path string) (*models.VirtualMachineCatalogManifest, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}

	manifestBytes, err := helper.ReadFromFile(path)
	if err != nil {
		return nil, err
	}

	manifest := &models.VirtualMachineCatalogManifest{}
	err = json.Unmarshal(manifestBytes, manifest)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

func (s *CatalogManifestService) getConformName(name string) string {
	return helpers.NormalizeString(name)
}

func (s *CatalogManifestService) getMetaFilename(name string) string {
	name = s.getConformName(name)
	if !strings.HasSuffix(name, ".meta") {
		name = name + ".meta"
	}

	return name
}

func (s *CatalogManifestService) getPackFilename(name string) string {
	name = s.getConformName(name)
	if !strings.HasSuffix(name, ".pdpack") {
		name = name + ".pdpack"
	}

	return name
}

func (s *CatalogManifestService) compressMachine(ctx basecontext.ApiContext, path string, machineFileName string, destination string) (string, error) {
	startingTime := time.Now()
	tarFilename := machineFileName
	tarFilePath := filepath.Join(destination, filepath.Clean(tarFilename))

	tarFile, err := os.Create(filepath.Clean(tarFilePath))
	if err != nil {
		return "", err
	}
	defer tarFile.Close()

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	countFiles := 0
	if err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		countFiles += 1
		return nil
	}); err != nil {
		return "", err
	}

	compressed := 1
	err = filepath.Walk(path, func(machineFilePath string, info os.FileInfo, err error) error {
		ctx.LogInfof("[%v/%v] Compressing file %v", compressed, countFiles, machineFilePath)
		compressed += 1
		if err != nil {
			return err
		}

		if info.IsDir() {
			compressed -= 1
			return nil
		}

		f, err := os.Open(filepath.Clean(machineFilePath))
		if err != nil {
			return err
		}
		defer f.Close()

		relPath := strings.TrimPrefix(machineFilePath, path)
		hdr := &tar.Header{
			Name: relPath,
			Mode: int64(info.Mode()),
			Size: info.Size(),
		}
		if err := tarWriter.WriteHeader(hdr); err != nil {
			return err
		}

		_, err = io.Copy(tarWriter, f)
		return err
	})

	if err != nil {
		return "", err
	}

	endingTime := time.Now()
	ctx.LogInfof("Finished compressing machine from %s to %s in %v", path, tarFilePath, endingTime.Sub(startingTime))
	return tarFilePath, nil
}

func (s *CatalogManifestService) decompressMachine(ctx basecontext.ApiContext, machineFilePath string, destination string) error {
	staringTime := time.Now()
	tarFile, err := os.Open(filepath.Clean(machineFilePath))
	if err != nil {
		return err
	}
	defer tarFile.Close()

	tarReader := tar.NewReader(tarFile)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		machineFilePath, err := helpers.SanitizeArchivePath(destination, header.Name)
		if err != nil {
			return err
		}

		// Creating the basedir if it does not exist
		baseDir := filepath.Dir(machineFilePath)
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			if err := os.MkdirAll(baseDir, 0o750); err != nil {
				return err
			}
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(machineFilePath); os.IsNotExist(err) {
				if err := os.MkdirAll(machineFilePath, os.FileMode(header.Mode)); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			file, err := os.OpenFile(filepath.Clean(machineFilePath), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()

			if err := helpers.CopyTarChunks(file, tarReader); err != nil {
				return err
			}
		}
	}

	endingTime := time.Now()
	ctx.LogInfof("Finished decompressing machine from %s to %s, in %v", machineFilePath, destination, endingTime.Sub(staringTime))
	return nil
}
