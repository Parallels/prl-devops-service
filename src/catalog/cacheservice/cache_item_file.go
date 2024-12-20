package cacheservice

type CacheItemFile struct {
	BaseName         string
	CacheFileName    string
	IsCompressed     bool
	IsCachedFolder   bool
	NeedsRenaming    bool
	MetadataFileName string
	InvalidFiles     []string
}

func NewCacheItemFile(baseName string) *CacheItemFile {
	return &CacheItemFile{
		BaseName: baseName,
	}
}

func (c *CacheItemFile) IsValid() bool {
	return c.CacheFileName != "" && c.MetadataFileName != ""
}

func (c *CacheItemFile) NeedsCleaning() bool {
	return len(c.InvalidFiles) > 0
}
