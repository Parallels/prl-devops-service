package parity

// DomainSuite "config_schema" (PBI #421 task 08): the cold-engine first-boot
// matrix — SetId-as-first-write, seed-on-read singletons, idempotent SeedId,
// the single-write guard, and the schema-version round-trip including the
// empty-string and empty-table paths #424 and the seeds depend on.
// Seed is a deliberate no-op: every check IS the first-boot sequence.

import (
	"testing"

	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
)

func TestParityDomain_ConfigSchema(t *testing.T) {
	RunMatrix(t, DomainSuite{
		Name: "config_schema",
		Seed: func(t testing.TB, db datainterfaces.DatabaseService) {}, // cold engines on purpose
		Checks: []func(t testing.TB, legacy, sql datainterfaces.DatabaseService){
			checkBootSetId,
			checkSeedIdStability,
			checkSecondSetIdGuard,
			checkSchemaVersionRoundTrip,
			checkIdStableAfterAll,
		},
	})
}

func checkBootSetId(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := newRoleClaimCtx()
	lg, lerr := legacy.SetId(ctx, "fixed-boot-id")
	sg, serr := sql.SetId(ctx, "fixed-boot-id")
	assertSameError(t, "boot SetId", lerr, serr)
	if lg != sg || lg != "fixed-boot-id" {
		t.Errorf("boot SetId returns: %q vs %q", lg, sg)
	}
	lid, lerr := legacy.GetId(ctx)
	sid, serr := sql.GetId(ctx)
	assertSameError(t, "GetId after boot", lerr, serr)
	if lid != sid || lid != "fixed-boot-id" {
		t.Errorf("GetId after boot: %q vs %q", lid, sid)
	}
	lc, lcerr := legacy.GetConfiguration(ctx)
	sc, scerr := sql.GetConfiguration(ctx)
	assertSameError(t, "GetConfiguration after boot", lcerr, scerr)
	if lc == nil || sc == nil || lc.ID != sc.ID {
		t.Errorf("GetConfiguration drift: %+v vs %+v", lc, sc)
	}
}

func checkSeedIdStability(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := newRoleClaimCtx()
	for i := 0; i < 3; i++ {
		lid, lerr := legacy.SeedId(ctx)
		sid, serr := sql.SeedId(ctx)
		assertSameError(t, "SeedId", lerr, serr)
		if lid != sid || lid != "fixed-boot-id" {
			t.Fatalf("SeedId round %d: %q vs %q", i, lid, sid)
		}
	}
}

func checkSecondSetIdGuard(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := newRoleClaimCtx()
	lg, lerr := legacy.SetId(ctx, "late-write")
	sg, serr := sql.SetId(ctx, "late-write")
	assertSameError(t, "second SetId", lerr, serr)
	if lg != sg { // both "" on the error path (legacy returns "" with the error)
		t.Errorf("second SetId returns: %q vs %q", lg, sg)
	}
}

func checkSchemaVersionRoundTrip(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := newRoleClaimCtx()
	for _, v := range []string{"", "0", "1", "2024-01-01", "anything—goes", ""} {
		assertSameError(t, "UpdateSchemaVersion",
			legacy.UpdateSchemaVersion(ctx, v),
			sql.UpdateSchemaVersion(ctx, v))
		lv, lerr := legacy.GetSchemaVersion(ctx)
		sv, serr := sql.GetSchemaVersion(ctx)
		assertSameError(t, "GetSchemaVersion", lerr, serr)
		if lv != sv || lv != v {
			t.Fatalf("version round-trip: %q vs %q (want %q)", lv, sv, v)
		}
	}
}

func checkIdStableAfterAll(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := newRoleClaimCtx()
	lid, _ := legacy.GetId(ctx)
	sid, _ := sql.GetId(ctx)
	if lid != sid || lid != "fixed-boot-id" {
		t.Errorf("installation id drifted: %q vs %q", lid, sid)
	}
}
