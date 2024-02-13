package artifactory

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog/common"
	"github.com/Parallels/pd-api-service/serviceprovider/download"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/config"
)

type ArtifactoryRepo struct {
	Host     string
	Port     string
	RepoName string
	ApiKey   string
}

const providerName = "artifactory"

type ArtifactoryProvider struct {
	Repo ArtifactoryRepo
}

func NewArtifactoryProvider() *ArtifactoryProvider {
	return &ArtifactoryProvider{}
}

func (s *ArtifactoryProvider) Name() string {
	return providerName
}

func (s *ArtifactoryProvider) GetProviderMeta(ctx basecontext.ApiContext) map[string]string {
	return map[string]string{
		common.PROVIDER_VAR_NAME: providerName,
		"url":                    s.Repo.Host,
		"port":                   s.Repo.Port,
		"repo":                   s.Repo.RepoName,
		"access_key":             s.Repo.ApiKey,
	}
}

func (s *ArtifactoryProvider) GetProviderRootPath(ctx basecontext.ApiContext) string {
	return "/"
}

func (s *ArtifactoryProvider) Check(ctx basecontext.ApiContext, connection string) (bool, error) {
	parts := strings.Split(connection, ";")
	provider := ""
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(strings.ToLower(part), common.PROVIDER_VAR_NAME+"=") {
			provider = strings.ReplaceAll(part, common.PROVIDER_VAR_NAME+"=", "")
		}
		if strings.Contains(strings.ToLower(part), "url=") {
			s.Repo.Host = strings.ReplaceAll(part, "url=", "")
		}
		if strings.Contains(strings.ToLower(part), "port=") {
			s.Repo.Port = strings.ReplaceAll(part, "port=", "")
		}
		if strings.Contains(strings.ToLower(part), "repo=") {
			s.Repo.RepoName = strings.ReplaceAll(part, "repo=", "")
		}
		if strings.Contains(strings.ToLower(part), "access_key=") {
			s.Repo.ApiKey = strings.ReplaceAll(part, "access_key=", "")
		}
	}
	if provider == "" || !strings.EqualFold(provider, providerName) {
		ctx.LogDebugf("Provider %s is not %s, skipping", providerName, provider)
		return false, nil
	}

	if s.Repo.RepoName == "" {
		return false, fmt.Errorf("missing artifactory repo name")
	}
	if s.Repo.Host == "" {
		return false, fmt.Errorf("missing artifactory host")
	}
	if s.Repo.ApiKey == "" {
		return false, fmt.Errorf("missing artifactory api key")
	}

	return true, nil
}

// uploadFile uploads a file to an S3 bucket
func (s *ArtifactoryProvider) PushFile(ctx basecontext.ApiContext, rootLocalPath string, path string, filename string) error {
	ctx.LogInfof("[%s] Pushing file %s", s.Name(), filename)
	localFilePath := filepath.Join(rootLocalPath, filename)
	remoteFilePath := filepath.Join(s.Repo.RepoName, path, filename)
	if !strings.HasPrefix(remoteFilePath, "/") {
		remoteFilePath = "/" + remoteFilePath
	}

	manager, err := s.getClient(ctx)
	if err != nil {
		return err
	}

	params := services.NewUploadParams()
	params.Pattern = localFilePath
	params.Target = remoteFilePath
	params.IncludeDirs = true
	params.ChecksumsCalcEnabled = true

	totalUploaded, totalFailed, err := manager.UploadFiles(params)
	if err != nil {
		return err
	}
	if totalFailed > 0 {
		return fmt.Errorf("failed to upload %s", filename)
	}
	if totalUploaded != 1 {
		return fmt.Errorf("failed to upload %s", filename)
	}

	ctx.LogInfof("[%s] Uploaded %s", s.Name(), filename)

	return nil
}

func (s *ArtifactoryProvider) PullFile(ctx basecontext.ApiContext, path string, filename string, destination string) error {
	ctx.LogInfof("[%s] Pulling file %s", s.Name(), filename)
	destinationFilePath := filepath.Join(destination, filename)
	remoteFilePath := filepath.Join(s.Repo.RepoName, path, filename)
	remoteFilePath = strings.TrimPrefix(remoteFilePath, "/")
	remoteFilePath = strings.TrimSuffix(remoteFilePath, "/")

	host := s.getHost()

	url := fmt.Sprintf("%s/%s", host, remoteFilePath)

	downloadSrv := download.NewDownloadService()
	headers := make(map[string]string, 0)
	headers["X-JFrog-Art-Api"] = s.Repo.ApiKey
	headers["Content-Type"] = "application/json"
	err := downloadSrv.DownloadFile(url, headers, destinationFilePath)
	if err != nil {
		return err
	}

	return nil
}

func (s *ArtifactoryProvider) DeleteFile(ctx basecontext.ApiContext, path string, fileName string) error {
	ctx.LogInfof("[%s] Deleting file %s", s.Name(), fileName)
	fullPath := filepath.Join(s.Repo.RepoName, path, fileName)
	fullPath = strings.TrimPrefix(fullPath, "/")
	fullPath = strings.TrimSuffix(fullPath, "/")

	host := s.getHost()

	url := fmt.Sprintf("%s/%s", host, fullPath)
	client := http.DefaultClient
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	request.Header.Add("X-JFrog-Art-Api", s.Repo.ApiKey)
	request.Header.Add("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 204 {
		return fmt.Errorf("error deleting file %s, status code: %d", fullPath, response.StatusCode)
	}

	return nil
}

func (s *ArtifactoryProvider) FileChecksum(ctx basecontext.ApiContext, path string, fileName string) (string, error) {
	ctx.LogInfof("[%s] Getting checksum for file %s", s.Name(), fileName)
	fullPath := filepath.Join(s.Repo.RepoName, path, fileName)
	fullPath = strings.TrimPrefix(fullPath, "/")
	fullPath = strings.TrimSuffix(fullPath, "/")

	host := s.getHost()

	url := fmt.Sprintf("%s/%s", host, fullPath)
	client := http.DefaultClient
	request, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return "", err
	}

	request.Header.Add("X-JFrog-Art-Api", s.Repo.ApiKey)
	request.Header.Add("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("error getting file checksum %s, status code: %d", fullPath, response.StatusCode)
	}

	if response.Header.Get("X-Checksum-Md5") != "" {
		return response.Header.Get("X-Checksum-Md5"), nil
	}

	// return checksum, nil
	return "", nil
}

func (s *ArtifactoryProvider) FileExists(ctx basecontext.ApiContext, path string, fileName string) (bool, error) {
	ctx.LogInfof("[%s] Checking if file %s exists", s.Name(), fileName)
	fullPath := filepath.Join(s.Repo.RepoName, path, fileName)
	fullPath = strings.TrimPrefix(fullPath, "/")
	fullPath = strings.TrimSuffix(fullPath, "/")

	host := s.getHost()

	url := fmt.Sprintf("%s/%s", host, fullPath)
	client := http.DefaultClient
	request, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false, err
	}

	request.Header.Add("X-JFrog-Art-Api", s.Repo.ApiKey)
	request.Header.Add("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return false, err
	}

	if response.StatusCode != 200 {
		return false, nil
	}

	return true, nil
}

func (s *ArtifactoryProvider) CreateFolder(ctx basecontext.ApiContext, folderPath string, folderName string) error {
	ctx.LogInfof("[%s] Creating folder %s", s.Name(), folderName)
	fullPath := filepath.Join(s.Repo.RepoName, folderPath, folderName)
	fullPath = strings.TrimPrefix(fullPath, "/")
	if !strings.HasSuffix(fullPath, "/") {
		fullPath = fullPath + "/"
	}

	host := s.getHost()

	url := fmt.Sprintf("%s/%s", host, fullPath)
	client := http.DefaultClient
	request, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}

	request.Header.Add("X-JFrog-Art-Api", s.Repo.ApiKey)
	request.Header.Add("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("error creating folder %s, status code: %d", fullPath, response.StatusCode)
	}

	return nil
}

func (s *ArtifactoryProvider) DeleteFolder(ctx basecontext.ApiContext, folderPath string, folderName string) error {
	ctx.LogInfof("[%s] Deleting folder %s", s.Name(), folderName)
	fullPath := filepath.Join(s.Repo.RepoName, folderPath, folderName)
	fullPath = strings.TrimPrefix(fullPath, "/")
	if !strings.HasSuffix(fullPath, "/") {
		fullPath = fullPath + "/"
	}

	host := s.getHost()

	url := fmt.Sprintf("%s/%s", host, fullPath)
	client := http.DefaultClient
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	request.Header.Add("X-JFrog-Art-Api", s.Repo.ApiKey)
	request.Header.Add("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("error creating folder %s, status code: %d", fullPath, response.StatusCode)
	}
	return nil
}

func (s *ArtifactoryProvider) FolderExists(ctx basecontext.ApiContext, folderPath string, folderName string) (bool, error) {
	ctx.LogInfof("[%s] Checking if folder %s exists", s.Name(), folderName)
	fullPath := filepath.Join(s.Repo.RepoName, folderPath, folderName)
	fullPath = strings.TrimPrefix(fullPath, "/")
	if !strings.HasSuffix(fullPath, "/") {
		fullPath = fullPath + "/"
	}

	host := s.getHost()

	url := fmt.Sprintf("%s/%s", host, fullPath)
	client := http.DefaultClient
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	request.Header.Add("X-JFrog-Art-Api", s.Repo.ApiKey)
	request.Header.Add("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return false, err
	}

	if response.StatusCode != 200 {
		return false, nil
	}

	// If the folder does not exist, return false
	return true, nil
}

func (s *ArtifactoryProvider) getHost() string {
	host := s.Repo.Host
	if !strings.HasSuffix(host, "/artifactory") {
		host = host + "/artifactory"
	}

	return host
}

func (s *ArtifactoryProvider) getClient(ctx basecontext.ApiContext) (artifactory.ArtifactoryServicesManager, error) {
	authDetails := auth.NewArtifactoryDetails()
	host := s.getHost()

	authDetails.SetUrl(host)
	authDetails.SetApiKey(s.Repo.ApiKey)

	serviceConfig, err := config.NewConfigBuilder().
		SetServiceDetails(authDetails).
		SetDryRun(false).
		SetContext(ctx.Context()).
		SetDialTimeout(180 * time.Second).
		SetOverallRequestTimeout(60 * time.Minute).
		SetHttpRetries(8).
		Build()
	if err != nil {
		return nil, err
	}

	manager, err := artifactory.New(serviceConfig)
	if err != nil {
		return nil, err
	}

	return manager, nil
}
