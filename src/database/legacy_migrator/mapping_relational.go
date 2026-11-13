package legacy_migrator

// Relational half of the plan (task 04, plan D-7): the #419 mapping document
// (419-gorm-model-tags/plan.md, table at its "Row-count formula" section) is
// the ONLY source of join names — `user_roles(user_id,role_id)`,
// `user_claims(user_id,claim_id)`, `role_claims` (explicit model,
// role_id+claim_id composite PK) — plus the serialise-degrade decisions:
// CatalogManifest contents, Job steps, PackerTemplate blobs and host
// cache_items are TEXT columns there (rows 7/9/11/14), so those collections
// stay on the scalar identity path. Nothing invented.
//
// Semantics carried from the spec: embedded children contribute EDGES;
// canonical ENTITY rows come from the top-level collections; orphan embedded
// entities are ALSO upserted ("no data loss" wins); edges are
// OnConflict-DoNothing batched — never delete-then-insert. `Role.Users` /
// `Claim.Users` back-references are json:"-" and never exist in data.json:
// edges derive from the forward direction only.

import (
	"sort"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// collectionPlan — the FULL ordered list (replaces task 03's scalar-only
// body; order rule: parents before children per #419's own table list, then
// Data-declaration order). Load-bearing for #425's FK reasoning even though
// sqlite does not enforce the edges (constraint:- everywhere, #419).
func collectionPlan() []CollectionSpec {
	return []CollectionSpec{
		entitySpec("Claims", "id", claimRows, countOf[models.Claim]),
		entitySpec("Roles", "id", roleRows, countOf[models.Role]),
		edgeSpec("RoleClaims", []string{"role_id", "claim_id"}, roleClaimEdges, roleClaimRawTotal, countOf[models.RoleClaims]),
		scalarSpec("Users", "id",
			func(d *data.Data) []models.User { return d.Users },
			countOf[models.User]),
		edgeSpec("UserRoles", []string{"user_id", "role_id"}, userRoleEdges, userRoleRawTotal, countTable("user_roles")),
		edgeSpec("UserClaims", []string{"user_id", "claim_id"}, userClaimEdges, userClaimRawTotal, countTable("user_claims")),
		scalarSpec("ApiKeys", "id",
			func(d *data.Data) []models.ApiKey { return d.ApiKeys },
			countOf[models.ApiKey]),
		pointerSpec("Configuration", "id",
			func(d *data.Data) *models.Configuration { return d.Configuration },
			countOf[models.Configuration]),
		scalarSpec("PackerTemplates", "id",
			func(d *data.Data) []models.PackerTemplate { return d.PackerTemplates },
			countOf[models.PackerTemplate]),
		scalarSpec("ManifestsCatalog", "id", // contents/packs are serialised TEXT (#419 row 7) — scalar path degrades, spec transcribes
			func(d *data.Data) []models.CatalogManifest { return d.ManifestsCatalog },
			countOf[models.CatalogManifest]),
		scalarSpec("OrchestratorHosts", "id",
			func(d *data.Data) []models.OrchestratorHost { return d.OrchestratorHosts },
			countOf[models.OrchestratorHost]),
		scalarSpec("VirtualMachines", "id", vmRows, countOf[models.VirtualMachine]),
		scalarSpec("VMSnapshots", "id", snapshotRows, countOf[models.VMSnapshot]),
		pointerSpec("ReverseProxy", "id",
			func(d *data.Data) *models.ReverseProxy { return d.ReverseProxy },
			countOf[models.ReverseProxy]),
		scalarSpec("ReverseProxyHosts", "id",
			func(d *data.Data) []models.ReverseProxyHost { return d.ReverseProxyHosts },
			countOf[models.ReverseProxyHost]),
		scalarSpec("CatalogManagers", "id",
			func(d *data.Data) []models.CatalogManager { return d.CatalogManagers },
			countOf[models.CatalogManager]),
		scalarSpec("Jobs", "id",
			func(d *data.Data) []models.Job { return d.Jobs },
			countOf[models.Job]),
		scalarSpec("EnrollmentTokens", "id",
			func(d *data.Data) []models.OrchestratorEnrollmentToken { return d.EnrollmentTokens },
			countOf[models.OrchestratorEnrollmentToken]),
		scalarSpec("UserConfigs", "id",
			func(d *data.Data) []models.UserConfig { return d.UserConfigs },
			countOf[models.UserConfig]),
	}
}

// entitySpec: scalar-shaped writer whose ROWS are a dedupe (top-level
// collection canonical, embedded copies preserved as orphans).
func entitySpec[T any](name, pk string,
	rows func(d *data.Data) []T,
	count func(db *gorm.DB) (int64, error),
) CollectionSpec {
	return scalarSpec(name, pk, rows, count) // same mechanics; rows() carries the dedupe
}

// edgeUpsert rows: typed structs bound to the #419 join table names.
type userRoleEdge struct {
	UserID string `gorm:"column:user_id"`
	RoleID string `gorm:"column:role_id"`
}

func (userRoleEdge) TableName() string { return "user_roles" }

type userClaimEdge struct {
	UserID  string `gorm:"column:user_id"`
	ClaimID string `gorm:"column:claim_id"`
}

func (userClaimEdge) TableName() string { return "user_claims" }

func edgeCols(columns []string) []clause.Column {
	c := make([]clause.Column, 0, len(columns))
	for _, col := range columns {
		c = append(c, clause.Column{Name: col})
	}
	return c
}

// edgeSpec: OnConflict DoNothing (idempotent, never delete-then-insert).
// LegacyCount keeps the #419 Σ formula (raw edge mentions); WrittenCount is
// the deduped batch insert result.
func edgeSpec[T any](name string, columns []string,
	rows func(d *data.Data) []T,
	rawTotal func(d *data.Data) int,
	count func(db *gorm.DB) (int64, error),
) CollectionSpec {
	return CollectionSpec{
		Name:        name,
		LegacyCount: func(d *data.Data) int { return rawTotal(d) },
		Migrate: func(ctx basecontext.ApiContext, tx *gorm.DB, d *data.Data, chunk int) (int64, error) {
			r := rows(d)
			if len(r) == 0 {
				return 0, nil
			}
			res := tx.Clauses(clause.OnConflict{Columns: edgeCols(columns), DoNothing: true}).
				CreateInBatches(&r, chunk)
			return res.RowsAffected, res.Error
		},
		Validate: func(db *gorm.DB, d *data.Data) (int64, int64, error) {
			a, err := count(db)
			return int64(len(rows(d))), a, err // expected = deduped entity-edge count
		},
	}
}

func countTable(table string) func(db *gorm.DB) (int64, error) {
	return func(db *gorm.DB) (int64, error) {
		var n int64
		err := db.Table(table).Count(&n).Error
		return n, err
	}
}

// --- entity dedupe (canonical first; first-seen order = deterministic file order) ---

func claimRows(d *data.Data) []models.Claim {
	seen := map[string]int{}
	var out []models.Claim
	add := func(c models.Claim) {
		if c.ID == "" {
			return
		}
		if i, ok := seen[c.ID]; ok {
			_ = i
			return // canonical (top-level) first-seen wins; embedded copies never overwrite
		}
		seen[c.ID] = len(out)
		out = append(out, c)
	}
	for _, c := range d.Claims {
		add(c) // canonical first
	}
	for _, r := range roleRows(d) { // top-level AND orphan roles (embedded copies carry their own claims)
		for _, c := range r.Claims {
			add(c)
		}
	}
	for _, u := range d.Users {
		for _, c := range u.Claims {
			add(c)
		}
	}
	return out
}

func roleRows(d *data.Data) []models.Role {
	seen := map[string]int{}
	var out []models.Role
	add := func(r models.Role) {
		if r.ID == "" {
			return
		}
		if _, ok := seen[r.ID]; ok {
			return // canonical (top-level) first-seen wins
		}
		seen[r.ID] = len(out)
		out = append(out, r)
	}
	for _, r := range d.Roles {
		add(r)
	}
	for _, u := range d.Users {
		for _, r := range u.Roles {
			add(r)
		}
	}
	return out
}

// --- edges (deduped + sorted for byte-deterministic batches) ---

func roleClaimEdges(d *data.Data) []models.RoleClaims {
	seen := map[string]bool{}
	var out []models.RoleClaims
	for _, r := range roleRows(d) {
		for _, c := range r.Claims {
			if c.ID == "" {
				continue
			}
			k := r.ID + "|" + c.ID
			if seen[k] {
				continue
			}
			seen[k] = true
			out = append(out, models.RoleClaims{RoleID: r.ID, ClaimID: c.ID})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].RoleID != out[j].RoleID {
			return out[i].RoleID < out[j].RoleID
		}
		return out[i].ClaimID < out[j].ClaimID
	})
	return out
}

func roleClaimRawTotal(d *data.Data) int {
	n := 0
	for _, r := range roleRows(d) {
		n += len(r.Claims)
	}
	return n
}

func userRoleEdges(d *data.Data) []userRoleEdge {
	seen := map[string]bool{}
	var out []userRoleEdge
	for _, u := range d.Users {
		for _, r := range u.Roles {
			if r.ID == "" {
				continue
			}
			k := u.ID + "|" + r.ID
			if seen[k] {
				continue
			}
			seen[k] = true
			out = append(out, userRoleEdge{UserID: u.ID, RoleID: r.ID})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UserID != out[j].UserID {
			return out[i].UserID < out[j].UserID
		}
		return out[i].RoleID < out[j].RoleID
	})
	return out
}

func userRoleRawTotal(d *data.Data) int {
	n := 0
	for _, u := range d.Users {
		n += len(u.Roles)
	}
	return n
}

func userClaimEdges(d *data.Data) []userClaimEdge {
	seen := map[string]bool{}
	var out []userClaimEdge
	for _, u := range d.Users {
		for _, c := range u.Claims {
			if c.ID == "" {
				continue
			}
			k := u.ID + "|" + c.ID
			if seen[k] {
				continue
			}
			seen[k] = true
			out = append(out, userClaimEdge{UserID: u.ID, ClaimID: c.ID})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UserID != out[j].UserID {
			return out[i].UserID < out[j].UserID
		}
		return out[i].ClaimID < out[j].ClaimID
	})
	return out
}

func userClaimRawTotal(d *data.Data) int {
	n := 0
	for _, u := range d.Users {
		n += len(u.Claims)
	}
	return n
}

// --- embedded-child flattenings ---

func vmRows(d *data.Data) []models.VirtualMachine {
	var out []models.VirtualMachine
	for _, h := range d.OrchestratorHosts {
		for _, vm := range h.VirtualMachines {
			if vm.HostId == "" { // host key preserved onto the child (#419 row 22)
				vm.HostId = h.ID
			}
			out = append(out, vm)
		}
	}
	return out
}

// snapshotRows flattens BOTH wrapper collections into vm_snapshots (row 20/21):
// HostsVMSnapshotsRecord.map keys are data (VM uuids) → vm_id; record.HostId →
// host_id. Sorted keys + a final total sort ⇒ byte-deterministic batches.
// The Data.VMSnapshots wrapper carries no host — its rows keep host_id "".
func snapshotRows(d *data.Data) []models.VMSnapshot {
	var out []models.VMSnapshot
	for _, rec := range d.HostsVMSnapshots {
		keys := make([]string, 0, len(rec.VMSnapshots))
		for k := range rec.VMSnapshots {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, vmID := range keys {
			for _, s := range rec.VMSnapshots[vmID] {
				s.HostID, s.VMID = rec.HostId, vmID
				out = append(out, s)
			}
		}
	}
	for _, w := range d.VMSnapshots {
		for _, s := range w.VMSnapshot {
			s.HostID, s.VMID = "", w.VMId
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].HostID != out[j].HostID {
			return out[i].HostID < out[j].HostID
		}
		if out[i].VMID != out[j].VMID {
			return out[i].VMID < out[j].VMID
		}
		return out[i].ID < out[j].ID
	})
	return out
}
