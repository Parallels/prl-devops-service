package models

type PullProgressType string

const (
	PullProgressTypeDownloading      PullProgressType = "downloading"
	PullProgressTypeDecompressing    PullProgressType = "decompressing"
	PullProgressTypeCopyingFromCache PullProgressType = "copying_from_cache"
)

type PullProgressNotification struct {
	Type        PullProgressType
	Message     string
	Percentage  int
	TotalSize   int64
	CurrentSize int64
}
