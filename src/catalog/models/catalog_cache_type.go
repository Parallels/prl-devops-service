package models

type CatalogCacheType int

const (
	CatalogCacheTypeNone CatalogCacheType = iota
	CatalogCacheTypeFile
	CatalogCacheTypeFolder
)

func (c CatalogCacheType) String() string {
	switch c {
	case CatalogCacheTypeNone:
		return "none"
	case CatalogCacheTypeFile:
		return "file"
	case CatalogCacheTypeFolder:
		return "folder"
	default:
		return "unknown"
	}
}
