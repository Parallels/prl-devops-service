package local

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/common"
	"github.com/Parallels/prl-devops-service/compressor"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/writers"

	"github.com/cjlapao/common-go/helper"
)

const providerName = "local-storage"

type LocalProviderConfig struct {
	Path string
}

type LocalProvider struct {
	Config        LocalProviderConfig
	JobId         string
	currentAction string
}

func NewLocalProvider() *LocalProvider {
	return &LocalProvider{
		Config: LocalProviderConfig{},
	}
}

func (s *LocalProvider) Name() string {
	return providerName
}

func (s *LocalProvider) CanStream() bool {
	return false
}

func (s *LocalProvider) GetProviderMeta(ctx basecontext.ApiContext) map[string]string {
	result := map[string]string{
		common.PROVIDER_VAR_NAME: providerName,
	}
	if s.Config.Path != "" {
		result["catalog_path"] = s.Config.Path
	}

	return result
}

func (s *LocalProvider) GetProviderRootPath(ctx basecontext.ApiContext) string {
	return s.Config.Path
}

func (s *LocalProvider) SetJobId(jobId string) {
	s.JobId = jobId
}

func (s *LocalProvider) SetCurrentAction(action string) {
	s.currentAction = action
}

func (s *LocalProvider) Check(ctx basecontext.ApiContext, connection string) (bool, error) {
	parts := strings.Split(connection, ";")
	provider := ""
	for _, part := range parts {
		part = strings.TrimSpace(part)
		lowered := strings.ToLower(part)
		if strings.Contains(lowered, common.PROVIDER_VAR_NAME+"=") {
			provider = strings.ReplaceAll(lowered, common.PROVIDER_VAR_NAME+"=", "")
		}
		if strings.Contains(lowered, "catalog_path=") {
			s.Config.Path = strings.ReplaceAll(part, "catalog_path=", "")
		}
	}
	if provider == "" || !strings.EqualFold(provider, providerName) {
		ctx.LogDebugf("Provider %s is not %s, skipping", providerName, provider)
		return false, nil
	}

	if s.Config.Path == "" {
		dir, err := os.Getwd()
		if err != nil {
			ctx.LogErrorf("Error getting current directory: %v", err)
			return false, err
		}

		fullPath := filepath.Join(dir, "catalog")
		if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
			ctx.LogErrorf("Error creating catalog directory: %v", err)
			return false, err
		}

		s.Config.Path = fullPath
	}

	return true, nil
}

func (s *LocalProvider) PushFile(ctx basecontext.ApiContext, localRoot, path, filename string) error {
	srcPath := filepath.Join(localRoot, filename)
	destPath := filepath.Join(path, filename)
	if !strings.HasPrefix(destPath, s.Config.Path) {
		destPath = filepath.Join(s.Config.Path, destPath)
	}

	srcFile, err := os.Open(filepath.Clean(srcPath))
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(filepath.Clean(destPath))
	if err != nil {
		return err
	}
	defer destFile.Close()

	fileInfo, err := srcFile.Stat()
	if err != nil {
		_, err = io.Copy(destFile, srcFile)
		return err
	}

	action := s.currentAction
	if action == "" {
		action = constants.ActionUploadingPackFile
	}
	pr := writers.NewProgressFileReader(srcFile, fileInfo.Size(), action)
	pr.SetJobId(s.JobId)
	pr.SetCorrelationId(s.JobId)
	pr.SetPrefix("Uploading")
	_, err = io.Copy(destFile, pr)
	return err
}

func (s *LocalProvider) PullFile(ctx basecontext.ApiContext, path, filename, destination string) error {
	srcPath := filepath.Join(path, filename)
	destPath := filepath.Join(destination, filename)
	if !strings.HasPrefix(srcPath, s.Config.Path) {
		srcPath = filepath.Join(s.Config.Path, srcPath)
	}

	srcFile, err := os.Open(filepath.Clean(srcPath))
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(filepath.Clean(destPath))
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

func (s *LocalProvider) PullFileAndDecompress(ctx basecontext.ApiContext, path, filename, destination string) error {
	srcPath := filepath.Join(path, filename)
	if !strings.HasPrefix(srcPath, s.Config.Path) {
		srcPath = filepath.Join(s.Config.Path, srcPath)
	}

	tempFile, err := os.CreateTemp("", "local-pull-decompress-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	srcFile, err := os.Open(filepath.Clean(srcPath))
	if err != nil {
		return err
	}
	defer srcFile.Close()

	tmpFile, err := os.Create(filepath.Clean(tempPath))
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	if _, err = io.Copy(tmpFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy to temporary file: %w", err)
	}
	tmpFile.Close()

	if err := compressor.DecompressFileWithStepChannel(ctx, tempPath, destination, nil, s.JobId, constants.ActionDecompressingPackFile); err != nil {
		return fmt.Errorf("decompression failed: %w", err)
	}

	return nil
}

func (s *LocalProvider) PullFileToMemory(ctx basecontext.ApiContext, path string, filename string) ([]byte, error) {
	ctx.LogInfof("Pulling file %s", filename)
	maxFileSize := 0.5 * 1024 * 1024 // 0.5MB

	srcPath := filepath.Join(path, filename)
	if !strings.HasPrefix(srcPath, s.Config.Path) {
		srcPath = filepath.Join(s.Config.Path, srcPath)
	}

	srcFile, err := os.Open(filepath.Clean(srcPath))
	if err != nil {
		return nil, err
	}
	defer srcFile.Close()

	fileInfo, err := srcFile.Stat()
	if err != nil {
		return nil, err
	}

	if fileInfo.Size() > int64(maxFileSize) {
		return nil, errors.New("file is too large to be read into memory")
	}

	fileContent := make([]byte, fileInfo.Size())
	if _, err = io.ReadFull(srcFile, fileContent); err != nil {
		return nil, err
	}

	return fileContent, nil
}

func (s *LocalProvider) DeleteFile(ctx basecontext.ApiContext, path string, fileName string) error {
	filePath := filepath.Join(path, fileName)
	if !strings.HasPrefix(filePath, s.Config.Path) {
		filePath = filepath.Join(s.Config.Path, filePath)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(filePath)
}

func (s *LocalProvider) FileChecksum(ctx basecontext.ApiContext, path string, fileName string) (string, error) {
	fullPath := filepath.Join(path, fileName)
	if !strings.HasPrefix(fullPath, s.Config.Path) {
		fullPath = filepath.Join(s.Config.Path, fullPath)
	}
	return helpers.GetFileMD5Checksum(fullPath)
}

func (s *LocalProvider) FileExists(ctx basecontext.ApiContext, path string, fileName string) (bool, error) {
	fullPath := filepath.Join(path, fileName)
	if !strings.HasPrefix(fullPath, s.Config.Path) {
		fullPath = filepath.Join(s.Config.Path, fullPath)
	}
	exists := helper.FileExists(fullPath)
	return exists, nil
}

func (s *LocalProvider) CreateFolder(ctx basecontext.ApiContext, path string, folderName string) error {
	folderPath := filepath.Join(path, folderName)
	if !strings.HasPrefix(folderPath, s.Config.Path) {
		folderPath = filepath.Join(s.Config.Path, folderPath)
	}
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func (s *LocalProvider) DeleteFolder(ctx basecontext.ApiContext, path string, folderName string) error {
	folderPath := filepath.Join(path, folderName)
	if !strings.HasPrefix(folderPath, s.Config.Path) {
		folderPath = filepath.Join(s.Config.Path, folderPath)
	}

	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(folderPath)
}

func (s *LocalProvider) FolderExists(ctx basecontext.ApiContext, path string, folderName string) (bool, error) {
	fullPath := filepath.Join(path, folderName)
	if !strings.HasPrefix(fullPath, s.Config.Path) {
		fullPath = filepath.Join(s.Config.Path, fullPath)
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func (s *LocalProvider) FileSize(ctx basecontext.ApiContext, path string, fileName string) (int64, error) {
	fullPath := filepath.Join(path, fileName)
	if !strings.HasPrefix(fullPath, s.Config.Path) {
		fullPath = filepath.Join(s.Config.Path, fullPath)
	}

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}
