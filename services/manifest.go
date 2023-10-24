package services

import (
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/models"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ManifestService struct {
}

func NewManifestService() *ManifestService {
	return &ManifestService{}
}

func (s *ManifestService) GenerateManifest(name, localPath string) (*models.RemoteVirtualMachineManifest, error) {
	result := models.RemoteVirtualMachineManifest{}
	result.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	result.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	result.Name = name
	result.Path = localPath

	_, file := filepath.Split(localPath)
	ext := filepath.Ext(file)
	result.Type = ext[1:]

	isDir, err := helpers.IsDirectory(localPath)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, fmt.Errorf("The path %v is not a directory", localPath)
	}

	files, err := s.getManifestFiles(localPath, "")
	if err != nil {
		return nil, err
	}

	result.Files = files
	return &result, nil
}

func (s *ManifestService) getManifestFiles(path string, relativePath string) ([]models.RemoteVirtualMachineFile, error) {
	result := make([]models.RemoteVirtualMachineFile, 0)
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			files, err := s.getManifestFiles(filepath.Join(path, file.Name()), "")
			if err != nil {
				return nil, err
			}
			result = append(result, files...)
			continue
		}

		manifestFile := models.RemoteVirtualMachineFile{
			Path: filepath.Join(path, file.Name()),
		}
		fileInfo, err := file.Info()
		if err != nil {
			return nil, err
		}
		manifestFile.Size = fileInfo.Size()
		manifestFile.CreatedAt = fileInfo.ModTime().Format(time.RFC3339Nano)
		manifestFile.UpdatedAt = fileInfo.ModTime().Format(time.RFC3339Nano)
		checksum, err := helpers.GetFileChecksum(manifestFile.Path)
		if err != nil {
			return nil, err
		}
		manifestFile.Checksum = checksum
		result = append(result, manifestFile)
	}

	return result, nil
}
