package local

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog/common"
	"github.com/Parallels/pd-api-service/helpers"

	"github.com/cjlapao/common-go/helper"
)

const providerName = "local-storage"

type LocalProviderConfig struct {
	Path string
}

type LocalProvider struct {
	Config LocalProviderConfig
}

func NewLocalProvider() *LocalProvider {
	return &LocalProvider{
		Config: LocalProviderConfig{},
	}
}

func (s *LocalProvider) Name() string {
	return providerName
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

func (s *LocalProvider) Check(ctx basecontext.ApiContext, connection string) (bool, error) {
	parts := strings.Split(connection, ";")
	provider := ""
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(strings.ToLower(part), common.PROVIDER_VAR_NAME+"=") {
			provider = strings.ReplaceAll(part, common.PROVIDER_VAR_NAME+"=", "")
		}
		if strings.Contains(strings.ToLower(part), "catalog_path=") {
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

	if s.Config.Path == "" {
		return false, errors.New("missing catalog_path")
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

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
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
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalProvider) DeleteFile(ctx basecontext.ApiContext, path string, fileName string) error {
	filePath := filepath.Join(path, fileName)
	if !strings.HasPrefix(filePath, s.Config.Path) {
		filePath = filepath.Join(s.Config.Path, filePath)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// file does not exist, return nil
		return nil
	}
	err := os.Remove(filePath)
	if err != nil {
		return err
	}
	return nil
}

func (s *LocalProvider) FileChecksum(ctx basecontext.ApiContext, path string, fileName string) (string, error) {
	fullPath := filepath.Join(path, fileName)
	if !strings.HasPrefix(fullPath, s.Config.Path) {
		fullPath = filepath.Join(s.Config.Path, fullPath)
	}
	checksum, err := helpers.GetFileChecksum(fullPath)
	if err != nil {
		return "", err
	}
	return checksum, nil
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
		err := os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
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
		// folder does not exist, return nil
		return nil
	}
	err := os.RemoveAll(folderPath)
	if err != nil {
		return err
	}
	return nil
}

func (s *LocalProvider) FolderExists(ctx basecontext.ApiContext, path string, fileName string) (bool, error) {
	fullPath := filepath.Join(path, fileName)
	if !strings.HasPrefix(fullPath, s.Config.Path) {
		fullPath = filepath.Join(s.Config.Path, fullPath)
	}
	exists := helper.FileExists(fullPath)
	return exists, nil
}
