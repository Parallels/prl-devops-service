package cleanupservice

import (
	"Parallels/pd-api-service/basecontext"
	"Parallels/pd-api-service/catalog/interfaces"
)

type CleanupRequest struct {
	RemoteStorageService interfaces.RemoteStorageService `json:"provider"`
	Operations           []CleanupOperation              `json:"operations"`
}

func NewCleanupRequest() *CleanupRequest {
	return &CleanupRequest{
		Operations: []CleanupOperation{},
	}
}

func (r *CleanupRequest) NeedsCleanup() bool {
	return len(r.Operations) > 0
}

func (r *CleanupRequest) AddCleanupOperation(operation CleanupOperation) {
	r.Operations = append(r.Operations, operation)
}

func (r *CleanupRequest) Clean(ctx basecontext.ApiContext) []error {
	errors := []error{}
	for _, operation := range r.Operations {
		operation.Clean(ctx)
		if operation.HasError() {
			errors = append(errors, operation.Error)
		}
	}

	return errors
}

func (r *CleanupRequest) AddLocalFileCleanupOperation(filePath string, isFolder bool) {
	r.Operations = append(r.Operations, CleanupOperation{
		FilePath:    filePath,
		IsDirectory: isFolder,
		Type:        CleanupOperationTypeLocal,
	})
}

func (r *CleanupRequest) AddRemoteFileCleanupOperation(filePath string, isFolder bool) {
	r.Operations = append(r.Operations, CleanupOperation{
		RemoteStorageService: r.RemoteStorageService,
		IsDirectory:          isFolder,
		FilePath:             filePath,
		Type:                 CleanupOperationTypeRemote,
	})
}

func (r *CleanupRequest) AddRestApiCallCleanupOperation(host string, port string, urlPath string, user string, password string, apiKey string) {
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
