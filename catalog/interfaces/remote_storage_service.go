package interfaces

type RemoteStorageService interface {
	Name() string
	Check(connection string) (bool, error)
	GetProviderRootPath() string
	FileChecksum(path string, fileName string) (string, error)
	GetProviderMeta() map[string]string
	FileExists(path string, fileName string) (bool, error)
	PushFile(rootLocalPath string, path string, filename string) error
	PullFile(path string, filename string, rootDestination string) error
	DeleteFile(path string, fileName string) error
	CreateFolder(path string, folderName string) error
	DeleteFolder(path string, folderName string) error
	FolderExists(path string, folderName string) (bool, error)
}
