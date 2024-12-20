package cleanupservice

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/interfaces"
)

type CleanupService struct {
	RemoteStorageService interfaces.RemoteStorageService `json:"provider"`
	Operations           []CleanupOperation              `json:"operations"`
}

func NewCleanupService() *CleanupService {
	return &CleanupService{
		Operations: []CleanupOperation{},
	}
}

func (r *CleanupService) NeedsCleanup() bool {
	return len(r.Operations) > 0
}

func (r *CleanupService) AddCleanupOperation(operation CleanupOperation) {
	r.Operations = append(r.Operations, operation)
}

func (r *CleanupService) Clean(ctx basecontext.ApiContext) []error {
	errors := []error{}
	for _, operation := range r.Operations {
		_ = operation.Clean(ctx)
		if operation.HasError() {
			errors = append(errors, operation.Error)
		}
	}

	return errors
}

func (r *CleanupService) AddLocalFileCleanupOperation(filePath string, isFolder bool) {
	r.Operations = append(r.Operations, CleanupOperation{
		FilePath:    filePath,
		IsDirectory: isFolder,
		Type:        CleanupOperationTypeLocal,
	})
}

func (r *CleanupService) RemoveLocalFileCleanupOperation(filePath string) {
	for i, operation := range r.Operations {
		if operation.FilePath == filePath && operation.Type == CleanupOperationTypeLocal {
			r.Operations = append(r.Operations[:i], r.Operations[i+1:]...)
		}
	}
}

func (r *CleanupService) AddRemoteFileCleanupOperation(filePath string, isFolder bool) {
	r.Operations = append(r.Operations, CleanupOperation{
		RemoteStorageService: r.RemoteStorageService,
		IsDirectory:          isFolder,
		FilePath:             filePath,
		Type:                 CleanupOperationTypeRemote,
	})
}

func (r *CleanupService) RemoveRemoteFileCleanupOperation(filePath string) {
	for i, operation := range r.Operations {
		if operation.FilePath == filePath && operation.Type == CleanupOperationTypeRemote {
			r.Operations = append(r.Operations[:i], r.Operations[i+1:]...)
		}
	}
}

func (r *CleanupService) AddRestApiCallCleanupOperation(host string, port string, urlPath string, user string, password string, apiKey string) {
	r.Operations = append(r.Operations, CleanupOperation{
		Type:     CleanupOperationTypeRestApiCall,
		Host:     host,
		Port:     port,
		UrlPath:  urlPath,
		User:     user,
		Password: password,
		ApiKey:   apiKey,
	})
}

func (r *CleanupService) RemoveRestApiCallCleanupOperation(host string, port string, urlPath string, user string, password string, apiKey string) {
	for i, operation := range r.Operations {
		if operation.Host == host && operation.Port == port && operation.UrlPath == urlPath && operation.User == user && operation.Password == password && operation.ApiKey == apiKey && operation.Type == CleanupOperationTypeRestApiCall {
			r.Operations = append(r.Operations[:i], r.Operations[i+1:]...)
		}
	}
}
