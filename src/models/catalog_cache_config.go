package models

type CatalogCacheConfig struct {
	Enabled                 bool   `json:"enabled"`
	Folder                  string `json:"folder,omitempty"`
	KeepFreeDiskSpace       int64  `json:"keep_free_disk_space,omitempty"`
	MaxSize                 int64  `json:"max_size,omitempty"`
	AllowAboveFreeDiskSpace bool   `json:"allow_above_free_disk_space,omitempty"`
}
