package models

import (
	"sort"
	"time"
)

type CachedManifests struct {
	TotalSize int64                           `json:"total_size"`
	Manifests []VirtualMachineCatalogManifest `json:"manifests"`
}

func (c *CachedManifests) SortManifestsByRanking() {
	sort.SliceStable(c.Manifests, func(i, j int) bool {
		thisDate, err := time.Parse(time.RFC3339, c.Manifests[i].CacheLastUsed)
		if err != nil {
			return false
		}
		thisDate = time.Date(thisDate.Year(), thisDate.Month(), thisDate.Day(), 0, 0, 0, 0, thisDate.Location())
		thatDate, err := time.Parse(time.RFC3339, c.Manifests[j].CacheLastUsed)
		if err != nil {
			return false
		}
		thatDate = time.Date(thatDate.Year(), thatDate.Month(), thatDate.Day(), 0, 0, 0, 0, thatDate.Location())

		if thisDate.Equal(thatDate) {
			return c.Manifests[i].CacheUsedCount < c.Manifests[j].CacheUsedCount
		}

		return thisDate.Before(thatDate)
	})
}

func (c *CachedManifests) SortManifestsByCachedDate() {
	sort.SliceStable(c.Manifests, func(i, j int) bool {
		thisDate, err := time.Parse(time.RFC3339, c.Manifests[i].CachedDate)
		if err != nil {
			return false
		}
		thisDate = time.Date(thisDate.Year(), thisDate.Month(), thisDate.Day(), 0, 0, 0, 0, thisDate.Location())
		thatDate, err := time.Parse(time.RFC3339, c.Manifests[j].CachedDate)
		if err != nil {
			return false
		}
		thatDate = time.Date(thatDate.Year(), thatDate.Month(), thatDate.Day(), 0, 0, 0, 0, thatDate.Location())

		return thisDate.Before(thatDate)
	})
}
