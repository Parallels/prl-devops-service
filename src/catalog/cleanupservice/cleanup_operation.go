package cleanupservice

import (
	"path/filepath"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog/interfaces"
	"github.com/Parallels/pd-api-service/errors"

	"github.com/cjlapao/common-go/helper"
)

type CleanupOperationType int

const (
	CleanupOperationTypeLocal CleanupOperationType = iota
	CleanupOperationTypeRemote
	CleanupOperationTypeRestApiCall
)

type CleanupOperation struct {
	RemoteStorageService interfaces.RemoteStorageService `json:"provider"`
	IsDirectory          bool                            `json:"is_directory"`
	FilePath             string                          `json:"file_path"`
	Type                 CleanupOperationType            `json:"type"`
	Host                 string                          `json:"host"`
	Port                 string                          `json:"port"`
	UrlPath              string                          `json:"path"`
	User                 string                          `json:"user"`
	Password             string                          `json:"password"`
	ApiKey               string                          `json:"api_key"`
	Error                error                           `json:"error"`
}

func (r *CleanupOperation) Clean(ctx basecontext.ApiContext) error {
	if r.Type == CleanupOperationTypeRemote {
		if err := r.cleanRemote(ctx); err != nil {
			return err
		}
	}

	if r.Type == CleanupOperationTypeLocal {
		if err := r.cleanLocal(); err != nil {
			return err
		}
	}

	if r.Type == CleanupOperationTypeRestApiCall {
		return nil
	}

	r.Error = errors.Newf("Unknown cleanup operation type %d", r.Type)
	return r.Error
}

func (r *CleanupOperation) cleanRemote(ctx basecontext.ApiContext) error {
	if r.RemoteStorageService == nil {
		r.Error = errors.Newf("RemoteStorageService is nil, cannot clean %s", r.FilePath)
		return r.Error
	}
	path := filepath.Dir(r.FilePath)
	fileName := filepath.Base(r.FilePath)

	if r.IsDirectory {
		exists, err := r.RemoteStorageService.FolderExists(ctx, path, fileName)
		if err != nil {
			r.Error = errors.Newf("Error checking if folder %s exists: %s", r.FilePath, err.Error())
			return r.Error
		}
		if exists {
			if err := r.RemoteStorageService.DeleteFolder(ctx, path, fileName); err != nil {
				r.Error = errors.Newf("Error deleting folder %s: %s", r.FilePath, err.Error())
				return r.Error
			}
		}
	} else {
		exists, err := r.RemoteStorageService.FileExists(ctx, path, fileName)
		if err != nil {
			r.Error = errors.Newf("Error checking if file %s exists: %s", r.FilePath, err.Error())
			return r.Error
		}
		if exists {
			if err := r.RemoteStorageService.DeleteFile(ctx, path, fileName); err != nil {
				r.Error = errors.Newf("Error deleting file %s: %s", r.FilePath, err.Error())
				return r.Error
			}
		}
	}

	return nil
}

func (r *CleanupOperation) cleanLocal() error {
	if r.IsDirectory {
		if !helper.FileExists(r.FilePath) {
			return nil
		}
		if err := helper.DeleteAllFiles(r.FilePath); err != nil {
			r.Error = errors.Newf("Error deleting folder %s: %s", r.FilePath, err.Error())
			return r.Error
		}

		if helper.FileExists(r.FilePath) {
			if err := helper.DeleteFile(r.FilePath); err != nil {
				r.Error = errors.Newf("Error deleting folder %s: %s", r.FilePath, err.Error())
				return r.Error
			}
		}
	} else {
		if !helper.FileExists(r.FilePath) {
			return nil
		}
		if err := helper.DeleteFile(r.FilePath); err != nil {
			r.Error = errors.Newf("Error deleting file %s: %s", r.FilePath, err.Error())
			return r.Error
		}
	}

	return nil
}

func (r *CleanupOperation) HasError() bool {
	return r.Error != nil
}
