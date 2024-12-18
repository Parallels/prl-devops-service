package catalog

import (
	"archive/tar"
	"bufio"
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
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"

	"github.com/cjlapao/common-go/helper"
)

type CompressorType int

const (
	CompressorTypeGzip CompressorType = iota
	CompressorTypeTar
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
	s.sendPushStepInfo(r, "Compressing manifest files")
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
	manifest.PackSize = fileInfo.Size()

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
	return s.decompressMachine(ctx, machineFilePath, destination)
}

func (s *CatalogManifestService) decompressMachine(ctx basecontext.ApiContext, machineFilePath string, destination string) error {
	staringTime := time.Now()
	filePath := filepath.Clean(machineFilePath)
	compressedFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer compressedFile.Close()

	fileType, err := s.detectFileType(filePath)
	if err != nil {
		return err
	}

	var fileReader io.Reader

	switch fileType {
	case "tar":
		fileReader = compressedFile
	case "gzip":
		// Create a gzip reader
		bufferReader := bufio.NewReader(compressedFile)
		gzipReader, err := gzip.NewReader(bufferReader)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		fileReader = gzipReader
	case "tar.gz":
		// Create a gzip reader
		bufferReader := bufio.NewReader(compressedFile)
		gzipReader, err := gzip.NewReader(bufferReader)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		fileReader = gzipReader
	}

	tarReader := tar.NewReader(fileReader)
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
			ctx.LogDebugf("Directory type found for file %v (byte %v, rune %v)", machineFilePath, header.Typeflag, string(header.Typeflag))
			if _, err := os.Stat(machineFilePath); os.IsNotExist(err) {
				if err := os.MkdirAll(machineFilePath, os.FileMode(header.Mode)); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			ctx.LogDebugf("HardFile type found for file %v (byte %v, rune %v): size %v", machineFilePath, header.Typeflag, string(header.Typeflag), header.Size)
			file, err := os.OpenFile(filepath.Clean(machineFilePath), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()

			if err := helpers.CopyTarChunks(file, tarReader, header.Size); err != nil {
				return err
			}
		case tar.TypeGNUSparse:
			ctx.LogDebugf("Sparse File type found for file %v (byte %v, rune %v): size %v", machineFilePath, header.Typeflag, string(header.Typeflag), header.Size)
			file, err := os.OpenFile(filepath.Clean(machineFilePath), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()

			if err := helpers.CopyTarChunks(file, tarReader, header.Size); err != nil {
				return err
			}
		case tar.TypeSymlink:
			ctx.LogDebugf("Symlink File type found for file %v (byte %v, rune %v)", machineFilePath, header.Typeflag, string(header.Typeflag))
			os.Symlink(header.Linkname, machineFilePath)
			realLinkPath, err := filepath.EvalSymlinks(filepath.Join(destination, header.Linkname))
			if err != nil {
				ctx.LogWarnf("Error resolving symlink path: %v", header.Linkname)
				if err := os.Remove(machineFilePath); err != nil {
					return fmt.Errorf("failed to remove invalid symlink: %v", err)
				}
			} else {
				relLinkPath, err := filepath.Rel(destination, realLinkPath)
				if err != nil || strings.HasPrefix(filepath.Clean(relLinkPath), "..") {
					return fmt.Errorf("invalid symlink path: %v", header.Linkname)
				}
				os.Symlink(realLinkPath, machineFilePath)
			}
		default:
			ctx.LogWarnf("Unknown type found for file %v, ignoring (byte %v, rune %v)", machineFilePath, header.Typeflag, string(header.Typeflag))
		}
	}

	endingTime := time.Now()
	ctx.LogInfof("Finished decompressing machine from %s to %s, in %v", machineFilePath, destination, endingTime.Sub(staringTime))
	return nil
}

func handleSparseFile(header *tar.Header, tarReader *tar.Reader, destDir string) error {
	outFile, err := os.Create(destDir + "/" + header.Name)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer outFile.Close()

	fmt.Printf("Writing sparse file: %s (%d bytes)\n", header.Name, header.Size)

	// Copy exactly the size of the sparse file
	if _, err := io.CopyN(outFile, tarReader, header.Size); err != nil && err != io.EOF {
		return fmt.Errorf("error writing sparse file content: %v", err)
	}
	return nil
}

func handleRegularFile(header *tar.Header, tarReader *tar.Reader, destDir string) error {
	outFile, err := os.Create(destDir + "/" + header.Name)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer outFile.Close()

	fmt.Printf("Writing regular file: %s (%d bytes)\n", header.Name, header.Size)

	// Copy exactly the size of the regular file
	if _, err := io.CopyN(outFile, tarReader, header.Size); err != nil && err != io.EOF {
		return fmt.Errorf("error writing file content: %v", err)
	}
	return nil
}

func (s *CatalogManifestService) sendPullStepInfo(r *models.PullCatalogManifestRequest, msg string) {
	if r.StepChannel != nil {
		r.StepChannel <- msg
	}
}

func (s *CatalogManifestService) sendPushStepInfo(r *models.PushCatalogManifestRequest, msg string) {
	if r.StepChannel != nil {
		r.StepChannel <- msg
	}
}
