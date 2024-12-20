package models

type CacheResponse struct {
	IsCached         bool
	MetadataFilePath string
	PackFilePath     string
	Checksum         string
	Type             CatalogCacheType
}
