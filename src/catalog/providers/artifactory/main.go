package artifactory

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/common"
	"github.com/Parallels/prl-devops-service/serviceprovider/download"
	"github.com/Parallels/prl-devops-service/writers"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/config"
)

var globalClient *artifactory.ArtifactoryServicesManager

type authenticationMethod string

const (
	ApiKeyMethod   authenticationMethod = "api_key"
	UserPassMethod authenticationMethod = "user_pass"
)

type ArtifactoryRepo struct {
	Host     string
	Port     string
	RepoName string
	ApiKey   string
	UserName string
	Password string
}

const providerName = "artifactory"

type ArtifactoryProvider struct {
	Repo            ArtifactoryRepo
	ProgressChannel chan int
	FileNameChannel chan string
}

func NewArtifactoryProvider() *ArtifactoryProvider {
	return &ArtifactoryProvider{}
}

func (s *ArtifactoryProvider) Name() string {
	return providerName
}

func (s *ArtifactoryProvider) CanStream() bool {
	return false
}

func (s *ArtifactoryProvider) GetProviderMeta(ctx basecontext.ApiContext) map[string]string {
	return map[string]string{
		common.PROVIDER_VAR_NAME: providerName,
		"url":                    s.Repo.Host,
		"port":                   s.Repo.Port,
		"repo":                   s.Repo.RepoName,
		"access_key":             s.Repo.ApiKey,
		"username":               s.Repo.UserName,
		"password":               s.Repo.Password,
	}
}

func (s *ArtifactoryProvider) SetProgressChannel(fileNameChannel chan string, progressChannel chan int) {
	s.ProgressChannel = progressChannel
	s.FileNameChannel = fileNameChannel
}

func (s *ArtifactoryProvider) GetProviderRootPath(ctx basecontext.ApiContext) string {
	return "/"
}

func (s *ArtifactoryProvider) getAuthenticationMethod() authenticationMethod {
	if s.Repo.ApiKey != "" {
		return ApiKeyMethod
	}
	return UserPassMethod
}

// Check checks the connection to the Artifactory provider.
// It parses the connection string and sets the necessary fields in the ArtifactoryProvider struct.
// If the provider is not the expected provider or any required field is missing, it returns false and an error.
// Otherwise, it returns true and no error.
//
// Parameters:
// - ctx: The API context for logging and other operations.
// - connection: The connection string to the Artifactory provider.
//
// Returns:
// - bool: Indicates whether the connection is valid or not.
// - error: An error if any required field is missing or if the provider is not the expected provider.
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
		if strings.Contains(strings.ToLower(part), "username=") {
			s.Repo.UserName = strings.ReplaceAll(part, "username=", "")
		}
		if strings.Contains(strings.ToLower(part), "password=") {
			s.Repo.Password = strings.ReplaceAll(part, "password=", "")
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
	if s.Repo.ApiKey == "" && s.Repo.UserName == "" && s.Repo.Password == "" {
		return false, fmt.Errorf("missing artifactory api key or username and password")
	}
	if s.Repo.ApiKey != "" && s.Repo.UserName != "" && s.Repo.Password != "" {
		return false, fmt.Errorf("artifactory api key and username and password are mutually exclusive")
	}
	if s.Repo.UserName != "" && s.Repo.Password == "" {
		return false, fmt.Errorf("artifactory username requires password")
	}
	if s.Repo.UserName == "" && s.Repo.Password != "" {
		return false, fmt.Errorf("artifactory password requires username")
	}

	return true, nil
}

func (s *ArtifactoryProvider) PushFile(ctx basecontext.ApiContext, rootLocalPath string, path string, filename string) error {
	ctx.LogInfof("[%s] Pushing file %s", s.Name(), filename)
	localFilePath := filepath.Join(rootLocalPath, filename)
	remoteFilePath := filepath.Join(s.Repo.RepoName, path, filename)
	if !strings.HasPrefix(remoteFilePath, "/") {
		remoteFilePath = "/" + remoteFilePath
	}

	if s.FileNameChannel != nil {
		s.FileNameChannel <- filename
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
	if s.getAuthenticationMethod() == ApiKeyMethod {
		headers["X-JFrog-Art-Api"] = s.Repo.ApiKey
	} else {
		headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(s.Repo.UserName+":"+s.Repo.Password))
	}
	headers["Content-Type"] = "application/json"

	fileSize, err := s.GetFileSize(ctx, path, filename)
	if err != nil {
		return err
	}
	progressReporter := writers.NewProgressReporter(fileSize, s.ProgressChannel)
	err = downloadSrv.DownloadFile(url, headers, destinationFilePath, progressReporter)
	if err != nil {
		return err
	}

	return nil
}

func (s *ArtifactoryProvider) PullFileAndDecompress(ctx basecontext.ApiContext, path string, filename string, destination string) error {
	ctx.LogInfof("[%s] Pulling file %s", s.Name(), filename)
	destinationFilePath := filepath.Join(destination, filename)
	remoteFilePath := filepath.Join(s.Repo.RepoName, path, filename)
	remoteFilePath = strings.TrimPrefix(remoteFilePath, "/")
	remoteFilePath = strings.TrimSuffix(remoteFilePath, "/")

	host := s.getHost()

	url := fmt.Sprintf("%s/%s", host, remoteFilePath)

	downloadSrv := download.NewDownloadService()
	headers := make(map[string]string, 0)
	if s.getAuthenticationMethod() == ApiKeyMethod {
		headers["X-JFrog-Art-Api"] = s.Repo.ApiKey
	} else {
		headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(s.Repo.UserName+":"+s.Repo.Password))
	}
	headers["Content-Type"] = "application/json"

	fileSize, err := s.GetFileSize(ctx, path, filename)
	if err != nil {
		return err
	}
	progressReporter := writers.NewProgressReporter(fileSize, s.ProgressChannel)
	err = downloadSrv.DownloadFile(url, headers, destinationFilePath, progressReporter)
	if err != nil {
		return err
	}

	return nil
}

func (s *ArtifactoryProvider) PullFileToMemory(ctx basecontext.ApiContext, path string, filename string) ([]byte, error) {
	ctx.LogInfof("[%s] Pulling file %s", s.Name(), filename)
	maxFileSize := 0.5 * 1024 * 1024 // 0.5MB

	remoteFilePath := filepath.Join(s.Repo.RepoName, path, filename)
	remoteFilePath = strings.TrimPrefix(remoteFilePath, "/")
	remoteFilePath = strings.TrimSuffix(remoteFilePath, "/")

	host := s.getHost()

	url := fmt.Sprintf("%s/%s", host, remoteFilePath)

	downloadSrv := download.NewDownloadService()
	headers := make(map[string]string, 0)
	if s.getAuthenticationMethod() == ApiKeyMethod {
		headers["X-JFrog-Art-Api"] = s.Repo.ApiKey
	} else {
		headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(s.Repo.UserName+":"+s.Repo.Password))
	}
	headers["Content-Type"] = "application/json"

	fileSize, err := s.GetFileSize(ctx, path, filename)
	if err != nil {
		return nil, err
	}

	if fileSize > int64(maxFileSize) {
		return nil, fmt.Errorf("file size is too large to pull to memory")
	}

	progressReporter := writers.NewProgressReporter(fileSize, s.ProgressChannel)
	data, err := downloadSrv.DownloadFileToBytes(url, headers, progressReporter)
	if err != nil {
		return nil, err
	}

	return data, nil
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

	if s.getAuthenticationMethod() == ApiKeyMethod {
		request.Header.Add("X-JFrog-Art-Api", s.Repo.ApiKey)
	} else {
		request.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.Repo.UserName+":"+s.Repo.Password)))
	}
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

	if s.getAuthenticationMethod() == ApiKeyMethod {
		request.Header.Add("X-JFrog-Art-Api", s.Repo.ApiKey)
	} else {
		request.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.Repo.UserName+":"+s.Repo.Password)))
	}
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

	if s.getAuthenticationMethod() == ApiKeyMethod {
		request.Header.Add("X-JFrog-Art-Api", s.Repo.ApiKey)
	} else {
		request.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.Repo.UserName+":"+s.Repo.Password)))
	}
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

func (s *ArtifactoryProvider) GetFileSize(ctx basecontext.ApiContext, path string, fileName string) (int64, error) {
	ctx.LogInfof("[%s] Checking if file %s exists", s.Name(), fileName)
	fullPath := filepath.Join(s.Repo.RepoName, path, fileName)
	fullPath = strings.TrimPrefix(fullPath, "/")
	fullPath = strings.TrimSuffix(fullPath, "/")

	host := s.getHost()

	url := fmt.Sprintf("%s/%s", host, fullPath)
	client := http.DefaultClient
	request, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return -1, err
	}

	if s.getAuthenticationMethod() == ApiKeyMethod {
		request.Header.Add("X-JFrog-Art-Api", s.Repo.ApiKey)
	} else {
		request.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.Repo.UserName+":"+s.Repo.Password)))
	}
	request.Header.Add("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return -1, err
	}

	if response.StatusCode != 200 {
		return -1, nil
	}

	if response.ContentLength > 0 {
		return response.ContentLength, nil
	}
	return -1, nil
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

	if s.getAuthenticationMethod() == ApiKeyMethod {
		request.Header.Add("X-JFrog-Art-Api", s.Repo.ApiKey)
	} else {
		request.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.Repo.UserName+":"+s.Repo.Password)))
	}
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

	if s.getAuthenticationMethod() == ApiKeyMethod {
		request.Header.Add("X-JFrog-Art-Api", s.Repo.ApiKey)
	} else {
		request.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.Repo.UserName+":"+s.Repo.Password)))
	}
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

	if s.getAuthenticationMethod() == ApiKeyMethod {
		request.Header.Add("X-JFrog-Art-Api", s.Repo.ApiKey)
	} else {
		request.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.Repo.UserName+":"+s.Repo.Password)))
	}
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

func (s *ArtifactoryProvider) FileSize(ctx basecontext.ApiContext, path string, fileName string) (int64, error) {
	ctx.LogInfof("[%s] Checking file %s size", s.Name(), fileName)

	headers := make(map[string]string, 0)
	if s.getAuthenticationMethod() == ApiKeyMethod {
		headers["X-JFrog-Art-Api"] = s.Repo.ApiKey
	} else {
		headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(s.Repo.UserName+":"+s.Repo.Password))
	}
	headers["Content-Type"] = "application/json"

	fileSize, err := s.GetFileSize(ctx, path, fileName)
	if err != nil {
		return -1, err
	}
	return fileSize, nil
}

func (s *ArtifactoryProvider) getHost() string {
	host := s.Repo.Host
	if !strings.HasSuffix(host, "/artifactory") {
		host += "/artifactory"
	}

	return host
}

func (s *ArtifactoryProvider) getClient(ctx basecontext.ApiContext) (artifactory.ArtifactoryServicesManager, error) {
	authDetails := auth.NewArtifactoryDetails()
	host := s.getHost()

	authDetails.SetUrl(host)
	if s.getAuthenticationMethod() == ApiKeyMethod {
		authDetails.SetApiKey(s.Repo.ApiKey)
	} else {
		authDetails.SetUser(s.Repo.UserName)
		authDetails.SetPassword(s.Repo.Password)
	}

	serviceConfig, err := config.NewConfigBuilder().
		SetServiceDetails(authDetails).
		SetDryRun(false).
		SetContext(ctx.Context()).
		SetDialTimeout(180 * time.Second).
		SetOverallRequestTimeout(60 * time.Minute).
		SetHttpRetries(8).
		SetThreads(1).
		Build()
	if err != nil {
		return nil, err
	}

	manager, err := artifactory.New(serviceConfig)
	if err != nil {
		return nil, err
	}

	globalClient = &manager
	return manager, nil
}
