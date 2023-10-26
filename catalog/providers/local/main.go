package local

import (
	"Parallels/pd-api-service/catalog/common"
	global_common "Parallels/pd-api-service/common"
	"Parallels/pd-api-service/helpers"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cjlapao/common-go/helper"
)

const providerName = "local-storage"

var logger = global_common.Logger

type LocalProviderConfig struct {
	Path string
}

type LocalProvider struct {
	Config LocalProviderConfig
}

func NewLocalProviderService() *LocalProvider {
	return &LocalProvider{
		Config: LocalProviderConfig{},
	}
}

func (s *LocalProvider) Name() string {
	return providerName
}

func (s *LocalProvider) GetProviderMeta() map[string]string {
	result := map[string]string{
		"provider": providerName,
	}
	if s.Config.Path != "" {
		result["catalog_path"] = s.Config.Path
	}

	return result
}

func (s *LocalProvider) GetProviderRootPath() string {
	return s.Config.Path
}

func (s *LocalProvider) Check(connection string) (bool, error) {
	parts := strings.Split(strings.ToLower(connection), ";")
	provider := ""
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, common.PROVIDER_VAR_NAME+"=") {
			provider = strings.ReplaceAll(part, common.PROVIDER_VAR_NAME+"=", "")
		}
		if strings.Contains(part, "catalog_path=") {
			s.Config.Path = strings.ReplaceAll(part, "catalog_path=", "")
		}
	}
	if provider != "" && provider != providerName {
		logger.Info("Provider %s is not %s, skipping", providerName, provider)
		return false, nil
	}

	if s.Config.Path == "" {
		dir, err := os.Getwd()
		if err != nil {
			logger.Error("Error getting current directory: %v", err)
			return false, err
		}

		fullPath := filepath.Join(dir, "catalog")
		if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
			logger.Error("Error creating catalog directory: %v", err)
			return false, err
		}

		s.Config.Path = fullPath
	}

	if s.Config.Path == "" {
		return false, errors.New("missing catalog_path")
	}

	return true, nil
}

func (s *LocalProvider) PushFile(localRoot, path, filename string) error {
	srcPath := filepath.Join(localRoot, filename)
	destPath := filepath.Join(path, filename)
	if !strings.HasPrefix(destPath, s.Config.Path) {
		destPath = filepath.Join(s.Config.Path, destPath)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
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

func (s *LocalProvider) PullFile(path, filename, destination string) error {
	srcPath := filepath.Join(path, filename)
	destPath := filepath.Join(destination, filename)
	if !strings.HasPrefix(srcPath, s.Config.Path) {
		srcPath = filepath.Join(s.Config.Path, srcPath)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
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

func (s *LocalProvider) DeleteFile(path string, fileName string) error {
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

func (s *LocalProvider) FileChecksum(path string, fileName string) (string, error) {
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

func (s *LocalProvider) FileExists(path string, fileName string) (bool, error) {
	fullPath := filepath.Join(path, fileName)
	if !strings.HasPrefix(fullPath, s.Config.Path) {
		fullPath = filepath.Join(s.Config.Path, fullPath)
	}
	exists := helper.FileExists(fullPath)
	return exists, nil
}

func (s *LocalProvider) CreateFolder(path string, folderName string) error {
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

func (s *LocalProvider) DeleteFolder(path string, folderName string) error {
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

func (s *LocalProvider) FolderExists(path string, fileName string) (bool, error) {
	fullPath := filepath.Join(path, fileName)
	if !strings.HasPrefix(fullPath, s.Config.Path) {
		fullPath = filepath.Join(s.Config.Path, fullPath)
	}
	exists := helper.FileExists(fullPath)
	return exists, nil
}
