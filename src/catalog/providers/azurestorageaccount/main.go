package azurestorageaccount

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/common"
	"github.com/Parallels/prl-devops-service/compressor"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/notifications"
	"github.com/Parallels/prl-devops-service/writers"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

type AzureStorageAccount struct {
	Name            string `json:"storage_account_name"`
	Key             string `json:"storage_account_key"`
	ContainerName   string `json:"container_name"`
	ProgressChannel chan int
}

const providerName = "azure-storage-account"

type AzureStorageAccountProvider struct {
	StorageAccount  AzureStorageAccount
	ProgressChannel chan int
	FileNameChannel chan string
}

func NewAzureStorageAccountProvider() *AzureStorageAccountProvider {
	return &AzureStorageAccountProvider{}
}

func (s *AzureStorageAccountProvider) Name() string {
	return providerName
}

func (s *AzureStorageAccountProvider) CanStream() bool {
	return false
}

func (s *AzureStorageAccountProvider) GetProviderMeta(ctx basecontext.ApiContext) map[string]string {
	return map[string]string{
		common.PROVIDER_VAR_NAME: providerName,
		"storage_account_name":   s.StorageAccount.Name,
		"storage_account_key":    s.StorageAccount.Key,
		"container_name":         s.StorageAccount.ContainerName,
	}
}

func (s *AzureStorageAccountProvider) GetProviderRootPath(ctx basecontext.ApiContext) string {
	return "/"
}

func (s *AzureStorageAccountProvider) SetProgressChannel(fileNameChannel chan string, progressChannel chan int) {
	s.ProgressChannel = progressChannel
	s.FileNameChannel = fileNameChannel
}

func (s *AzureStorageAccountProvider) Check(ctx basecontext.ApiContext, connection string) (bool, error) {
	parts := strings.Split(connection, ";")
	provider := ""
	for _, part := range parts {
		if strings.Contains(strings.ToLower(part), common.PROVIDER_VAR_NAME+"=") {
			provider = strings.ReplaceAll(part, common.PROVIDER_VAR_NAME+"=", "")
		}
		if strings.Contains(strings.ToLower(part), "storage_account_name=") {
			s.StorageAccount.Name = strings.ReplaceAll(part, "storage_account_name=", "")
		}
		if strings.Contains(strings.ToLower(part), "storage_account_key=") {
			s.StorageAccount.Key = strings.ReplaceAll(part, "storage_account_key=", "")
		}
		if strings.Contains(strings.ToLower(part), "container_name=") {
			s.StorageAccount.ContainerName = strings.ReplaceAll(part, "container_name=", "")
		}
	}
	if provider == "" || !strings.EqualFold(provider, providerName) {
		ctx.LogDebugf("Provider %s is not %s, skipping", providerName, provider)
		return false, nil
	}

	if s.StorageAccount.Name == "" {
		return false, fmt.Errorf("missing storage account name")
	}
	if s.StorageAccount.ContainerName == "" {
		return false, fmt.Errorf("missing storage account container name")
	}
	if s.StorageAccount.Key == "" {
		return false, fmt.Errorf("missing storage account key")
	}

	return true, nil
}

func (s *AzureStorageAccountProvider) PushFile(ctx basecontext.ApiContext, rootLocalPath string, path string, filename string) error {
	ctx.LogInfof("Pushing file %s", filename)
	localFilePath := filepath.Join(rootLocalPath, filename)
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")

	credential, err := azblob.NewSharedKeyCredential(s.StorageAccount.Name, s.StorageAccount.Key)
	if err != nil {
		return fmt.Errorf("invalid credentials with error: %s", err.Error())
	}
	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.StorageAccount.Name, s.StorageAccount.ContainerName, remoteFilePath))

	blobUrl := azblob.NewBlockBlobURL(*URL, azblob.NewPipeline(credential, azblob.PipelineOptions{}))

	file, err := os.Open(filepath.Clean(localFilePath))
	if err != nil {
		return err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		ctx.LogInfof("ERROR:", err)
		return err
	}

	defer file.Close()

	md5, err := helpers.GetFileMD5Checksum(localFilePath)
	if err != nil {
		return err
	}
	// md5Hash := base64.StdEncoding.EncodeToString([]byte(md5))

	_, err = azblob.UploadFileToBlockBlob(ctx.Context(), file, blobUrl, azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16,
		Progress: func(bytesTransferred int64) {
			if s.ProgressChannel != nil {
				s.ProgressChannel <- int(bytesTransferred * 100 / fileInfo.Size())
			}
		},
	})
	if err != nil {
		return err
	}

	_, err = blobUrl.SetHTTPHeaders(ctx.Context(), azblob.BlobHTTPHeaders{
		ContentType: "application/octet-stream",
		ContentMD5:  []byte(md5),
	}, azblob.BlobAccessConditions{})

	return err
}

func (s *AzureStorageAccountProvider) PullFile(ctx basecontext.ApiContext, path string, filename string, destination string) error {
	ctx.LogInfof("Pulling file %s", filename)
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")
	destinationFilePath := filepath.Join(destination, filename)
	credential, err := azblob.NewSharedKeyCredential(s.StorageAccount.Name, s.StorageAccount.Key)
	if err != nil {
		return fmt.Errorf("invalid credentials with error: %s", err.Error())
	}
	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.StorageAccount.Name, s.StorageAccount.ContainerName, remoteFilePath))

	blobUrl := azblob.NewBlockBlobURL(*URL, azblob.NewPipeline(credential, azblob.PipelineOptions{
		Retry: azblob.RetryOptions{
			MaxTries:   5,
			TryTimeout: 40 * time.Minute,
		},
	}))

	file, err := os.Create(filepath.Clean(destinationFilePath))
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new context with a longer deadline
	downloadContext, cancel := context.WithTimeout(ctx.Context(), 5*time.Hour)
	defer cancel()

	properties, err := blobUrl.GetProperties(downloadContext, azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return err
	}

	if properties.ContentLength() == 0 {
		return nil
	}

	err = azblob.DownloadBlobToFile(downloadContext, blobUrl.BlobURL, 0, azblob.CountToEnd, file, azblob.DownloadFromBlobOptions{
		Progress: func(bytesTransferred int64) {
			if s.ProgressChannel != nil {
				s.ProgressChannel <- int(bytesTransferred * 100 / properties.ContentLength())
			}
		},
	})

	return err
}

func (s *AzureStorageAccountProvider) PullFileAndDecompress(ctx basecontext.ApiContext, path string, filename string, destination string) error {
	ctx.LogInfof("Pulling file %s from Azure Blob Storage", filename)

	// Prepare the remote and local paths
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")

	// Create the Azure credentials
	credential, err := azblob.NewSharedKeyCredential(s.StorageAccount.Name, s.StorageAccount.Key)
	if err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	// Build the blob URL
	blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		s.StorageAccount.Name,
		s.StorageAccount.ContainerName,
		remoteFilePath,
	)

	u, err := url.Parse(blobURL)
	if err != nil {
		return fmt.Errorf("failed to parse blob URL: %w", err)
	}

	// Create the pipeline and blob URL
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{
		Retry: azblob.RetryOptions{
			MaxTries:   5,
			TryTimeout: 40 * time.Minute,
		},
	})
	blob := azblob.NewBlockBlobURL(*u, pipeline)

	// Create a new context with a longer deadline for large downloads
	downloadContext, cancel := context.WithTimeout(ctx.Context(), 5*time.Hour)
	defer cancel()

	// Get the blob properties (for size and other metadata)
	properties, err := blob.GetProperties(downloadContext, azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return fmt.Errorf("failed to get blob properties: %w", err)
	}

	blobSize := properties.ContentLength()
	if blobSize == 0 {
		// Empty blob, nothing to do
		ctx.LogInfof("Blob %s is empty, nothing to decompress.", filename)
		return nil
	}

	// Download the blob to get an io.ReadCloser stream
	resp, err := blob.Download(downloadContext, 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return fmt.Errorf("failed to initiate blob download: %w", err)
	}

	// The response body is a stream of the blob data
	bodyStream := resp.Body(azblob.RetryReaderOptions{})

	// Wrap the blob stream with a progress reader (similar to the S3 approach)
	// Adjust this line depending on your progress reader's constructor
	pr := writers.NewProgressReader(bodyStream, blobSize)
	pr.SetPrefix("Pulling")
	pr.SetFilename(filename)
	cid := pr.CorrelationId()

	// Now decompress from the reader directly to the destination
	// This should read the entire blob, decompressing as it goes.
	if err := compressor.DecompressFromReader(ctx, pr, destination); err != nil {
		return fmt.Errorf("decompression failed: %w", err)
	}

	// After successful extraction, notify completion
	ns := notifications.Get()
	msg := fmt.Sprintf("Pulling %s", filename)
	ns.NotifyProgress(cid, msg, 100)
	ns.NotifyInfo(fmt.Sprintf("Finished pulling and decompressing file %s", filename))

	return nil
}

func (s *AzureStorageAccountProvider) PullFileToMemory(ctx basecontext.ApiContext, path string, filename string) ([]byte, error) {
	ctx.LogInfof("Pulling file %s", filename)
	maxFileSize := 0.5 * 1024 * 1024 // 0.5MB

	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")

	credential, err := azblob.NewSharedKeyCredential(s.StorageAccount.Name, s.StorageAccount.Key)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials with error: %s", err.Error())
	}
	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.StorageAccount.Name, s.StorageAccount.ContainerName, remoteFilePath))

	blobUrl := azblob.NewBlockBlobURL(*URL, azblob.NewPipeline(credential, azblob.PipelineOptions{
		Retry: azblob.RetryOptions{
			MaxTries:   5,
			TryTimeout: 40 * time.Minute,
		},
	}))

	// Create a new context with a longer deadline
	downloadContext, cancel := context.WithTimeout(ctx.Context(), 5*time.Hour)
	defer cancel()

	properties, err := blobUrl.GetProperties(downloadContext, azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return nil, err
	}

	if properties.ContentLength() == 0 {
		return []byte{}, nil
	}

	if properties.ContentLength() > int64(maxFileSize) {
		return nil, fmt.Errorf("file size is too large to pull to memory")
	}

	data := make([]byte, properties.ContentLength())

	err = azblob.DownloadBlobToBuffer(downloadContext, blobUrl.BlobURL, 0, azblob.CountToEnd, data, azblob.DownloadFromBlobOptions{
		Progress: func(bytesTransferred int64) {
			if s.ProgressChannel != nil {
				s.ProgressChannel <- int(bytesTransferred * 100 / properties.ContentLength())
			}
		},
	})

	return data, err
}

func (s *AzureStorageAccountProvider) DeleteFile(ctx basecontext.ApiContext, path string, fileName string) error {
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, fileName), "/")
	credential, err := azblob.NewSharedKeyCredential(s.StorageAccount.Name, s.StorageAccount.Key)
	if err != nil {
		return fmt.Errorf("invalid credentials with error: %s", err.Error())
	}
	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.StorageAccount.Name, s.StorageAccount.ContainerName, remoteFilePath))

	blobUrl := azblob.NewBlockBlobURL(*URL, azblob.NewPipeline(credential, azblob.PipelineOptions{}))

	_, err = blobUrl.Delete(ctx.Context(), azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{})

	return err
}

func (s *AzureStorageAccountProvider) FileChecksum(ctx basecontext.ApiContext, path string, fileName string) (string, error) {
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, fileName), "/")
	credential, err := azblob.NewSharedKeyCredential(s.StorageAccount.Name, s.StorageAccount.Key)
	if err != nil {
		return "", fmt.Errorf("invalid credentials with error: %s", err.Error())
	}
	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.StorageAccount.Name, s.StorageAccount.ContainerName, remoteFilePath))

	blobUrl := azblob.NewBlockBlobURL(*URL, azblob.NewPipeline(credential, azblob.PipelineOptions{}))

	props, err := blobUrl.GetProperties(ctx.Context(), azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})

	fileCheckSum := string(props.ContentMD5())
	return fileCheckSum, err
}

func (s *AzureStorageAccountProvider) FileExists(ctx basecontext.ApiContext, path string, fileName string) (bool, error) {
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, fileName), "/")
	credential, err := azblob.NewSharedKeyCredential(s.StorageAccount.Name, s.StorageAccount.Key)
	if err != nil {
		return false, fmt.Errorf("invalid credentials with error: %s", err.Error())
	}
	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.StorageAccount.Name, s.StorageAccount.ContainerName, remoteFilePath))

	blobUrl := azblob.NewBlockBlobURL(*URL, azblob.NewPipeline(credential, azblob.PipelineOptions{}))

	props, err := blobUrl.GetProperties(ctx.Context(), azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return false, err
	}

	return props.ContentLength() > 0, err
}

func (s *AzureStorageAccountProvider) CreateFolder(ctx basecontext.ApiContext, folderPath string, folderName string) error {
	return nil
}

func (s *AzureStorageAccountProvider) DeleteFolder(ctx basecontext.ApiContext, folderPath string, folderName string) error {
	return nil
}

func (s *AzureStorageAccountProvider) FolderExists(ctx basecontext.ApiContext, folderPath string, folderName string) (bool, error) {
	return true, nil
}

func (s *AzureStorageAccountProvider) FileSize(ctx basecontext.ApiContext, path string, fileName string) (int64, error) {
	ctx.LogInfof("Getting file %s size", fileName)
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, fileName), "/")
	credential, err := azblob.NewSharedKeyCredential(s.StorageAccount.Name, s.StorageAccount.Key)
	if err != nil {
		return -1, fmt.Errorf("invalid credentials with error: %s", err.Error())
	}
	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.StorageAccount.Name, s.StorageAccount.ContainerName, remoteFilePath))

	blobUrl := azblob.NewBlockBlobURL(*URL, azblob.NewPipeline(credential, azblob.PipelineOptions{
		Retry: azblob.RetryOptions{
			MaxTries:   5,
			TryTimeout: 40 * time.Minute,
		},
	}))

	// Create a new context with a longer deadline
	downloadContext, cancel := context.WithTimeout(ctx.Context(), 5*time.Hour)
	defer cancel()

	properties, err := blobUrl.GetProperties(downloadContext, azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return -1, err
	}

	if properties.ContentLength() == 0 {
		return 0, nil
	}

	return properties.ContentLength(), nil
}
