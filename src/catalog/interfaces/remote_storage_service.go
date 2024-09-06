package interfaces

import (
	"github.com/Parallels/prl-devops-service/basecontext"
)

type RemoteStorageService interface {
	Name() string
	Check(ctx basecontext.ApiContext, connection string) (bool, error)
	SetProgressChannel(fileNameChannel chan string, progressChannel chan int)
	GetProviderRootPath(ctx basecontext.ApiContext) string
	FileChecksum(ctx basecontext.ApiContext, path string, fileName string) (string, error)
	GetProviderMeta(ctx basecontext.ApiContext) map[string]string
	FileExists(ctx basecontext.ApiContext, path string, fileName string) (bool, error)
	PushFile(ctx basecontext.ApiContext, rootLocalPath string, path string, filename string) error
	PullFile(ctx basecontext.ApiContext, path string, filename string, rootDestination string) error
	DeleteFile(ctx basecontext.ApiContext, path string, fileName string) error
	CreateFolder(ctx basecontext.ApiContext, path string, folderName string) error
	DeleteFolder(ctx basecontext.ApiContext, path string, folderName string) error
	FolderExists(ctx basecontext.ApiContext, path string, folderName string) (bool, error)
}
