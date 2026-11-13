package database

// The eight Tags/RequiredClaims/RequiredRoles mutators of the catalog
// domain (PBI #421 task 11 split, ER-2 file-size convention). Shared shape:
// RAW scan (no authz), first FoldEqual-ID hit, read-modify-write the slice
// in Go, one targeted serializer-column update. Legacy quirks mirrored:
//   - ADD variants declare `found` OUTSIDE the per-value loop (:383, :411,
//     :465, :521): once any value matches the stored array, every LATER
//     value in the same call is skipped even when new. Pinned, not fixed.
//   - REMOVE variants iterate the ORIGINAL (pre-splice) slice for the index
//     but splice the live one (:441-449): removing several values can thus
//     miss later ones or address a shifted index — the same append
//     expressions reproduce the same outcomes (including legacy's out-of-
//     range panic shape) on identical inputs.
//   - UpdateCatalogManifestRequiredRoles is an ADD-with-dedupe despite its
//     name (spec's "replace array" line stale — :392 appends).
// UpdateCatalogManifestTags (ID OR CatalogId+Version matcher) lives in
// catalog.go next to the reads it serves.

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
)

// catalogMutateArray runs the shared preamble + a Go-side slice mutation,
// then writes exactly one serialised column (never Save(struct) — that
// would clobber concurrent UpdatedAt/counter writes, task-11 spec).
func (g *GormDatabase) catalogMutateArray(recordId, col string, mutate func(original []string) []string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	rows, err := g.loadCatalogManifests()
	if err != nil {
		return err
	}
	for i := range rows {
		if !strings.EqualFold(rows[i].ID, recordId) {
			continue
		}
		var original []string
		switch col {
		case "tags":
			original = rows[i].Tags
		case "required_claims":
			original = rows[i].RequiredClaims
		case "required_roles":
			original = rows[i].RequiredRoles
		}
		set, err := catalogJSON(col, mutate(original))
		if err != nil {
			return err
		}
		return g.catalogSetColumn(rows[i].ID, set)
	}
	return interfaces.ErrCatalogManifestNotFound
}

// catalogAppender is the ADD body with the found-outside-loop quirk.
func catalogAppender(original []string, vals ...string) []string {
	cur := append([]string(nil), original...) // legacy splices the stored row while iterating its copy
	found := false
	for _, val := range vals {
		for _, r := range original {
			if strings.EqualFold(r, val) {
				found = true
				break
			}
		}
		if !found {
			cur = append(cur, val)
		}
	}
	return cur
}

// catalogRemover is the REMOVE body: index found on ORIGINAL, splice on
// the live copy, verbatim :448-450 arithmetic.
func catalogRemover(original []string, vals ...string) []string {
	cur := append([]string(nil), original...)
	for _, val := range vals {
		foundAt := -1
		for index, r := range original {
			if strings.EqualFold(r, val) {
				foundAt = index
				break
			}
		}
		if foundAt != -1 {
			cur = append(cur[:foundAt], cur[foundAt+1:]...)
		}
	}
	return cur
}

func (g *GormDatabase) UpdateCatalogManifestRequiredRoles(ctx basecontext.ApiContext, recordId string, roles ...string) error {
	return g.catalogMutateArray(recordId, "required_roles", func(o []string) []string { return catalogAppender(o, roles...) })
}

func (g *GormDatabase) AddCatalogManifestRequiredClaims(ctx basecontext.ApiContext, recordId string, claims ...string) error {
	return g.catalogMutateArray(recordId, "required_claims", func(o []string) []string { return catalogAppender(o, claims...) })
}

func (g *GormDatabase) RemoveCatalogManifestRequiredClaims(ctx basecontext.ApiContext, recordId string, claims ...string) error {
	return g.catalogMutateArray(recordId, "required_claims", func(o []string) []string { return catalogRemover(o, claims...) })
}

func (g *GormDatabase) AddCatalogManifestRequiredRoles(ctx basecontext.ApiContext, recordId string, roles ...string) error {
	return g.catalogMutateArray(recordId, "required_roles", func(o []string) []string { return catalogAppender(o, roles...) })
}

func (g *GormDatabase) RemoveCatalogManifestRequiredRoles(ctx basecontext.ApiContext, recordId string, roles ...string) error {
	return g.catalogMutateArray(recordId, "required_roles", func(o []string) []string { return catalogRemover(o, roles...) })
}

func (g *GormDatabase) AddCatalogManifestTags(ctx basecontext.ApiContext, recordId string, tags ...string) error {
	return g.catalogMutateArray(recordId, "tags", func(o []string) []string { return catalogAppender(o, tags...) })
}

func (g *GormDatabase) RemoveCatalogManifestTags(ctx basecontext.ApiContext, recordId string, tags ...string) error {
	return g.catalogMutateArray(recordId, "tags", func(o []string) []string { return catalogRemover(o, tags...) })
}
