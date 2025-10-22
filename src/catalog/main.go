package catalog

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/cleanupservice"
	"github.com/Parallels/prl-devops-service/catalog/interfaces"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/catalog/providers/artifactory"
	"github.com/Parallels/prl-devops-service/catalog/providers/aws_s3_bucket"
	"github.com/Parallels/prl-devops-service/catalog/providers/azurestorageaccount"
	"github.com/Parallels/prl-devops-service/catalog/providers/local"
	"github.com/Parallels/prl-devops-service/catalog/providers/minio"
	"github.com/Parallels/prl-devops-service/compressor"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/notifications"

	"github.com/cjlapao/common-go/helper"
)

type CompressorType int

const (
	CompressorTypeGzip CompressorType = iota
	CompressorTypeTar
)

type CatalogManifestService struct {
	ns             *notifications.NotificationService
	ctx            basecontext.ApiContext
	remoteServices []interfaces.RemoteStorageService
}

func NewManifestService(ctx basecontext.ApiContext) *CatalogManifestService {
	manifestService := &CatalogManifestService{
		ctx: ctx,
		ns:  notifications.Get(),
	}
	// Adding remote services to the catalog service
	manifestService.remoteServices = make([]interfaces.RemoteStorageService, 0)
	manifestService.AddRemoteService(aws_s3_bucket.NewAwsS3Provider())
	manifestService.AddRemoteService(local.NewLocalProvider())
	manifestService.AddRemoteService(azurestorageaccount.NewAzureStorageAccountProvider())
	manifestService.AddRemoteService(artifactory.NewArtifactoryProvider())
	manifestService.AddRemoteService(minio.NewMinioProvider())
	return manifestService
}

func (s *CatalogManifestService) GetProviders() []interfaces.RemoteStorageService {
	return s.remoteServices
}

func (s *CatalogManifestService) GetProviderFromConnection(connectionString string) (interfaces.RemoteStorageService, error) {
	for _, rs := range s.remoteServices {
		check, checkErr := rs.Check(s.ctx, connectionString)
		if checkErr != nil {
			s.ns.NotifyErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			return nil, checkErr
		}

		if !check {
			continue
		}

		return rs, nil
	}

	return nil, errors.NewWithCode("remote storage service was not found", 404)
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

func (s *CatalogManifestService) GenerateManifestContent(r *models.PushCatalogManifestRequest, manifest *models.VirtualMachineCatalogManifest) error {
	s.ns.NotifyInfof("Generating manifest content for %v", r.CatalogId)
	if manifest == nil {
		manifest = models.NewVirtualMachineCatalogManifest()
	}

	manifest.CleanupRequest = cleanupservice.NewCleanupService()
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

	s.ns.NotifyInfof("Getting manifest files for %v", r.CatalogId)
	files, err := s.getManifestFiles(r.LocalPath, "")
	if err != nil {
		return err
	}

	s.ns.NotifyInfof("Compressing manifest files for %v", r.CatalogId)
	s.sendPushStepInfo(r, "Compressing manifest files")
	packFilePath, err := s.compressMachine(r.LocalPath, manifestPackFileName, "/tmp", r.CompressPack, r.CompressPackLevel)
	if err != nil {
		return err
	}

	// Adding the zip file to the cleanup request
	manifest.CleanupRequest.AddLocalFileCleanupOperation(packFilePath, false)
	manifest.CompressedPath = packFilePath
	manifest.PackFile = "/tmp/" + manifestPackFileName
	manifest.IsCompressed = r.CompressPack

	fileInfo, err := os.Stat(packFilePath)
	if err != nil {
		return err
	}

	// Getting the total size of the original folder
	var totalSize int64 = 0
	err = filepath.Walk(r.LocalPath, func(_ string, info os.FileInfo, err error) error {
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
	manifest.Size = totalSize
	manifest.PackSize = fileInfo.Size()
	differenceInSize := manifest.Size - manifest.PackSize
	compressionPercentage := 0.0
	if manifest.Size > 0 {
		// calculating compression percentage and rounding to 2 decimal places
		compressionPercentage = helpers.RoundFloat((float64(differenceInSize)/float64(manifest.Size))*100, 2)
	} else {
		compressionPercentage = 0.0
	}
	manifest.CompressedSize = manifest.PackSize
	manifest.CompressedRatio = compressionPercentage
	if r.CompressPack {
		s.ns.NotifyInfof("Original size: %v bytes, Pack size: %v bytes, compressed percentage: %v%%", manifest.Size, manifest.PackSize, compressionPercentage)
	} else {
		s.ns.NotifyInfof("Original size: %v bytes, Pack size: %v bytes, compression not applied", manifest.Size, manifest.PackSize)
	}

	s.ns.NotifyInfof("Getting manifest package checksum for %v", r.CatalogId)
	checksum, err := helpers.GetFileMD5Checksum(packFilePath)
	if err != nil {
		return err
	}
	manifest.CompressedChecksum = checksum

	manifest.VirtualMachineContents = files
	s.ns.NotifyInfof("Finished generating manifest content for %v", r.CatalogId)
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

	return s.readManifestFromBytes(manifestBytes)
}

func (s *CatalogManifestService) readManifestFromBytes(value []byte) (*models.VirtualMachineCatalogManifest, error) {
	manifest := &models.VirtualMachineCatalogManifest{}
	err := json.Unmarshal(value, manifest)
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

func (s *CatalogManifestService) compressMachine(path string, machineFileName string, destination string, enableCompression bool, compressLevel int) (string, error) {
	compressLevelStr, compressLevelErr := helpers.GetCompressRatioEnvValue(compressLevel)

	// recovering to best compression if error
	if compressLevelErr != nil {
		compressLevel = gzip.BestCompression
		compressLevelStr = "best_compression"
	}

	startingTime := time.Now()
	tarFilename := machineFileName
	tarFilePath := filepath.Join(destination, filepath.Clean(tarFilename))

	tarFile, err := os.Create(filepath.Clean(tarFilePath))
	if err != nil {
		return "", err
	}
	defer tarFile.Close()

	targetWriter := io.Writer(tarFile)
	var gzipWriter *gzip.Writer

	if enableCompression {
		s.ns.NotifyInfof("Using gzip compression for %s with level %s (%v)", tarFilePath, compressLevelStr, compressLevel)
		gzipWriter, err = gzip.NewWriterLevel(tarFile, compressLevel)
		if err != nil {
			return "", err
		}
		targetWriter = gzipWriter
	}

	tarWriter := tar.NewWriter(targetWriter)
	defer tarWriter.Close()
	if gzipWriter != nil {
		defer gzipWriter.Close()
	}

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
		s.ns.NotifyInfof("[%v/%v] Compressing file %v", compressed, countFiles, machineFilePath)
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
	s.ns.NotifyInfof("Finished compressing machine from %s to %s in %v", path, tarFilePath, endingTime.Sub(startingTime))
	return tarFilePath, nil
}

// detectFileType determines whether a file is gzip, tar, tar.gz, or unknown.
func (s *CatalogManifestService) detectFileType(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "unknown", err
	}
	defer file.Close()

	// Read the first 512 bytes
	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("could not read file header: %w", err)
	}
	header = header[:n]

	// Reset file offset to the beginning
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("could not reset file offset: %w", err)
	}

	// Check for Gzip magic number
	if n >= 2 && header[0] == 0x1F && header[1] == 0x8B {
		// It's a gzip file, but is it a compressed tar?
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return "gzip", nil
		}
		defer gzipReader.Close()

		// Read the first 512 bytes of the decompressed data
		tarHeader := make([]byte, 512)
		n, err := gzipReader.Read(tarHeader)
		if err != nil && err != io.EOF {
			return "gzip", nil // It's a gzip file, but not a tar archive
		}
		tarHeader = tarHeader[:n]

		// Check for tar magic string in decompressed data
		if n > 262 {
			tarMagic := string(tarHeader[257 : 257+5])
			if tarMagic == "ustar" || tarMagic == "ustar\x00" {
				return "tar.gz", nil
			}
		}
		return "gzip", nil
	}

	// Check for Tar magic string at offset 257
	if n > 262 {
		tarMagic := string(header[257 : 257+5])
		if tarMagic == "ustar" || tarMagic == "ustar\x00" {
			return "tar", nil
		}
	}

	// If none of the above, return unknown
	return "unknown", errors.New("file format not recognized as gzip or tar")
}

func (s *CatalogManifestService) Unzip(ctx basecontext.ApiContext, machineFilePath string, destination string) error {
	return compressor.DecompressFile(ctx, machineFilePath, destination)
}

func (s *CatalogManifestService) sendPushStepInfo(r *models.PushCatalogManifestRequest, msg string) {
	if r.StepChannel != nil {
		r.StepChannel <- msg
	}
}
