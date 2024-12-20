package cacheservice

type CleanupRequirements struct {
	NeedsCleaning    bool
	SpaceNeeded      int64
	IsFatal          bool
	Reason           string
	FreeDiskSpace    int64
	CatalogTotalSize int64
}
