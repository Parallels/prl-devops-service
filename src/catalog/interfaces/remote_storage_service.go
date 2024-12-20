package interfaces

import (
	"github.com/Parallels/prl-devops-service/basecontext"
)

type RemoteStorageService interface {
	Name() string
	Check(ctx basecontext.ApiContext, connection string) (bool, error)
	CanStream() bool
	SetProgressChannel(fileNameChannel chan string, progressChannel chan int)
	GetProviderRootPath(ctx basecontext.ApiContext) string
	FileChecksum(ctx basecontext.ApiContext, path string, fileName string) (string, error)
	FileSize(ctx basecontext.ApiContext, path string, fileName string) (int64, error)
	GetProviderMeta(ctx basecontext.ApiContext) map[string]string
	FileExists(ctx basecontext.ApiContext, path string, fileName string) (bool, error)
	PushFile(ctx basecontext.ApiContext, rootLocalPath string, path string, filename string) error
	PullFile(ctx basecontext.ApiContext, path string, filename string, rootDestination string) error
	PullFileAndDecompress(ctx basecontext.ApiContext, path string, filename string, destination string) error
	PullFileToMemory(ctx basecontext.ApiContext, path string, filename string) ([]byte, error)
	DeleteFile(ctx basecontext.ApiContext, path string, fileName string) error
	CreateFolder(ctx basecontext.ApiContext, path string, folderName string) error
	DeleteFolder(ctx basecontext.ApiContext, path string, folderName string) error
	FolderExists(ctx basecontext.ApiContext, path string, folderName string) (bool, error)
}
