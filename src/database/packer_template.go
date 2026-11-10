package database

// SQL implementation of the packer-templates domain (PBI #421 task 10) —
// the D-4 showcase: the list read is the full ported pipeline (load →
// query.ParseFilter → query.FilterByProperty → query.GetAuthorizedRecords,
// insertion/rowid order, NO order layer) with zero SQL filter push-down.
// Legacy quirks mirrored:
//   - GetPackerTemplate goes THROUGH GetPackerTemplates, so authz applies to
//     single lookups too (an authz-blind caller sees NotFound, not the row).
//   - AddPackerTemplate's duplicate probe rides that authz-filtered lookup —
//     same-id creates that are authz-blind pass legacy's check and append a
//     second row; SQL cannot double-store a primary key, so the collision
//     surfaces as the duplicate sentinel (deviation: SQL is STRICTER here;
//     pinned by unit test).
//   - timestamps are raw time.Now().UTC().Format(RFC3339Nano) at legacy
//     :67-68 — NOT helpers.GetUtcCurrentDateTime (spec claim corrected).
//   - required claims/roles must EXIST: errors propagate verbatim (cross-
//     domain reads against the ported roles/claims domains, mirroring :71-82
//     — spec's "no cross-domain reads" corrected at task start).
//   - Delete: empty nameOrId answers nil (no sentinel — :94-96); scans the
//     RAW list (authz does NOT gate deletes); first EqualFold name||id hit;
//     Internal rows answer ErrRemovingInternalPackerTemplate unless
//     query.IsRootUser.
//   - Update: NO authz gate; first EqualFold-ID hit; Internal rows are
//     unupdatable for EVERYONE (:120-122 — root included); full-replacement
//     semantics (legacy assigns the whole struct :123 — unlike user-configs
//     this is NOT partial: payload CreatedAt/Name/everything wins, id case
//     included → raw column map with json.Marshal encodings, byte-identical
//     to GORM's serializer:json since both encode via encoding/json).
// 0 tx sites — D-9 lists no packer-template flow.

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/query"
)

func (g *GormDatabase) loadPackerTemplates() ([]models.PackerTemplate, error) {
	var rows []models.PackerTemplate
	err := g.gormDB.Table("packer_templates").Order("rowid").Find(&rows).Error
	if err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return rows, nil
}

func (g *GormDatabase) GetPackerTemplates(ctx basecontext.ApiContext, filter string) ([]models.PackerTemplate, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	dbFilter, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	rows, err := g.loadPackerTemplates()
	if err != nil {
		return nil, err
	}
	filtered, err := query.FilterByProperty(rows, dbFilter)
	if err != nil {
		return nil, err
	}
	return query.GetAuthorizedRecords(ctx, filtered...), nil
}

func (g *GormDatabase) GetPackerTemplate(ctx basecontext.ApiContext, nameOrId string) (*models.PackerTemplate, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	// Legacy routes the lookup through the AUTHZ-WRAPPED list (:44) — an
	// unauthorized caller cannot probe rows it cannot list. No empty-input
	// sentinel exists (:39-57 — spec claim #6; EqualFold("") only matches
	// empty name/id rows).
	templates, err := g.GetPackerTemplates(ctx, "")
	if err != nil {
		return nil, err
	}
	if found := scanPackerTemplate(templates, nameOrId); found != nil {
		return found, nil
	}
	return nil, interfaces.ErrPackerTemplateNotFound
}

// scanPackerTemplate is the shared EqualFold name||id matcher (first hit)
// over AUTHZ-FILTERED lists for reads, RAW lists for mutations — matching
// which list each legacy method scans.
func scanPackerTemplate(templates []models.PackerTemplate, nameOrId string) *models.PackerTemplate {
	for i := range templates {
		if equalFoldAny(nameOrId, templates[i].Name, templates[i].ID) {
			return &templates[i]
		}
	}
	return nil
}

func equalFoldAny(needle string, candidates ...string) bool {
	for _, c := range candidates {
		if strings.EqualFold(c, needle) {
			return true
		}
	}
	return false
}

func (g *GormDatabase) AddPackerTemplate(ctx basecontext.ApiContext, template *models.PackerTemplate) (*models.PackerTemplate, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	if probe, _ := g.GetPackerTemplate(ctx, template.ID); probe != nil {
		return nil, interfaces.ErrPackerTemplateAlreadyExists // authz-blind quirk, :63
	}
	template.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	template.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	for _, claim := range template.RequiredClaims {
		if _, err := g.GetClaim(ctx, claim); err != nil { // ported domain — verbatim propagation
			return nil, err
		}
	}
	for _, role := range template.RequiredRoles {
		if _, err := g.GetRole(ctx, role); err != nil {
			return nil, err
		}
	}
	dup := interfaces.ErrPackerTemplateAlreadyExists
	notFound := interfaces.ErrPackerTemplateNotFound
	if err := mapGormError(g.gormDB.Table("packer_templates").Create(template).Error, notFound, dup); err != nil {
		return nil, err
	}
	return template, nil
}

func (g *GormDatabase) DeletePackerTemplate(ctx basecontext.ApiContext, nameOrId string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if nameOrId == "" {
		return nil // legacy :94-96 — answers nil, NOT a sentinel
	}
	rows, err := g.loadPackerTemplates() // RAW list: deletes are not authz-gated
	if err != nil {
		return err
	}
	found := scanPackerTemplate(rows, nameOrId)
	if found == nil {
		return interfaces.ErrPackerTemplateNotFound
	}
	if found.Internal && !query.IsRootUser(ctx) {
		return interfaces.ErrRemovingInternalPackerTemplate
	}
	return mapGormError(g.gormDB.Table("packer_templates").Where("id = ?", found.ID).Delete(&models.PackerTemplate{}).Error,
		interfaces.ErrPackerTemplateNotFound, nil)
}

func (g *GormDatabase) UpdatePackerTemplate(ctx basecontext.ApiContext, template *models.PackerTemplate) (*models.PackerTemplate, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rows, err := g.loadPackerTemplates() // RAW list, EqualFold on ID only
	if err != nil {
		return nil, err
	}
	var found *models.PackerTemplate
	for i := range rows {
		if strings.EqualFold(rows[i].ID, template.ID) {
			found = &rows[i]
			break
		}
	}
	if found == nil {
		return nil, interfaces.ErrPackerTemplateNotFound
	}
	if found.Internal {
		return nil, interfaces.ErrUpdatingInternalPackerTemplate // everyone, root included
	}
	updates, err := packerTemplateColumnMap(template)
	if err != nil {
		return nil, err
	}
	if err := g.gormDB.Table("packer_templates").Where("id = ?", found.ID).Updates(updates).Error; err != nil {
		return nil, mapGormError(err, interfaces.ErrPackerTemplateNotFound, interfaces.ErrPackerTemplateAlreadyExists)
	}
	return template, nil
}

// packerTemplateColumnMap mirrors the whole-struct assignment at :123 —
// every column including the primary key (payload id case survives, as in
// legacy) takes the payload value. The serializer:json columns are encoded
// with encoding/json directly — the same encoder and thus the same bytes
// (sorted map keys) GORM's serializer:json writes on Create/Find.
func packerTemplateColumnMap(t *models.PackerTemplate) (map[string]interface{}, error) {
	m := map[string]interface{}{
		"id": t.ID, "name": t.Name, "owner": t.Owner, "hostname": t.Hostname,
		"description": t.Description, "packer_folder": t.PackerFolder,
		"internal": t.Internal, "updated_at": t.UpdatedAt, "created_at": t.CreatedAt,
	}
	for col, v := range map[string]interface{}{
		"variables": t.Variables, "specs": t.Specs, "defaults": t.Defaults,
		"addons": t.Addons, "required_roles": t.RequiredRoles, "required_claims": t.RequiredClaims,
	} {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, mapGormError(err, nil, nil)
		}
		m[col] = string(b)
	}
	return m, nil
}
