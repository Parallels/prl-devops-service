package database

// VM-snapshot stores on GORM (PBI #421 task 14). Both legacy stores share
// the normalised vm_snapshots table, separated by the host_id carrier.

import (
	"errors"
	"reflect"
	"testing"

	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
)

func TestSQLSnapshots_HostRecords(t *testing.T) {
	g := newSQLDB(t)
	ctx := tplRootCtx()

	if _, err := g.GetHostVMSnapshots(ctx, "h-A"); !errors.Is(err, interfaces.ErrOrchestratorHostsVMSnapshotsNotFound) {
		t.Fatalf("cold record: %v", err)
	}
	set := func(hostID string, snap models.VMSnapshots) {
		t.Helper()
		if err := g.SetHostVMSnapshots(ctx, hostID, snap); err != nil {
			t.Fatal(err)
		}
	}
	set("h-A", models.VMSnapshots{VMId: "vm-a", VMSnapshot: []models.VMSnapshot{{ID: "s1", Name: "one"}, {ID: "s2", Name: "two"}}})
	set("h-A", models.VMSnapshots{VMId: "vm-b", VMSnapshot: []models.VMSnapshot{{ID: "s3", Name: "three"}}})

	rec, err := g.GetHostVMSnapshots(ctx, "h-A")
	if err != nil || rec.HostId != "h-A" || len(rec.VMSnapshots) != 2 {
		t.Fatalf("record: %v %+v", err, rec)
	}
	if names := []string{rec.VMSnapshots["vm-a"][0].Name, rec.VMSnapshots["vm-a"][1].Name}; !reflect.DeepEqual(names, []string{"one", "two"}) {
		t.Fatalf("vm-a order: %v", names)
	}
	// Upsert replaces the touched VMId key only (:925 map[key]=list).
	set("h-A", models.VMSnapshots{VMId: "vm-a", VMSnapshot: []models.VMSnapshot{{ID: "s1b", Name: "one-prime"}}})
	rec, _ = g.GetHostVMSnapshots(ctx, "h-A")
	if len(rec.VMSnapshots["vm-a"]) != 1 || rec.VMSnapshots["vm-a"][0].ID != "s1b" || len(rec.VMSnapshots["vm-b"]) != 1 {
		t.Fatalf("key replace: %+v", rec.VMSnapshots)
	}
	// Empty ID gets minted (SQL primary key deviation, documented).
	set("h-A", models.VMSnapshots{VMId: "vm-c", VMSnapshot: []models.VMSnapshot{{Name: "minted"}}})
	rec, _ = g.GetHostVMSnapshots(ctx, "h-A")
	if rec.VMSnapshots["vm-c"][0].ID == "" {
		t.Fatal("empty id not minted")
	}
	// EXACT hostId match: Fold must NOT apply (:896 uses ==).
	if _, err := g.GetHostVMSnapshots(ctx, "h-a"); !errors.Is(err, interfaces.ErrOrchestratorHostsVMSnapshotsNotFound) {
		t.Fatalf("case-fold leak: %v", err)
	}
	// Empty list leaves the key gone and, if it was the last one, the record:
	// documented carrier deviation (legacy persists a nil-valued key).
	set("h-A", models.VMSnapshots{VMId: "vm-c", VMSnapshot: nil})
	rec, err = g.GetHostVMSnapshots(ctx, "h-A")
	_, gone := rec.VMSnapshots["vm-c"]
	if err != nil || gone {
		t.Fatalf("empty-list key: %v %+v", err, rec.VMSnapshots)
	}
}

func TestSQLSnapshots_GlobalList(t *testing.T) {
	g := newSQLDB(t)
	ctx := tplRootCtx()

	rows, err := g.GetListVMSnapshotsByVMId("vm-none")
	if err != nil || rows == nil || len(rows) != 0 {
		t.Fatalf("miss must be empty non-nil: %v %#v", err, rows)
	}
	set := func(vmID string, snaps ...models.VMSnapshot) {
		t.Helper()
		if err := g.SetListVMSnapshotsByVMId(vmID, models.VMSnapshots{VMId: vmID, VMSnapshot: snaps}); err != nil {
			t.Fatal(err)
		}
	}
	set("vm-a", models.VMSnapshot{ID: "g1", Name: "first"})
	set("vm-a", models.VMSnapshot{ID: "g2", Name: "second"}, models.VMSnapshot{ID: "g3", Name: "third"})
	rows, err = g.GetListVMSnapshotsByVMId("vm-a")
	if err != nil || len(rows) != 2 || rows[0].ID != "g2" || rows[1].ID != "g3" {
		t.Fatalf("whole-record replace: %v %+v", err, rows)
	}

	// Namespace isolation: host-record rows for the SAME vm id stay invisible
	// to the global store and vice versa.
	if err := g.SetHostVMSnapshots(ctx, "h-A", models.VMSnapshots{VMId: "vm-a",
		VMSnapshot: []models.VMSnapshot{{ID: "hr1", Name: "host-row"}}}); err != nil {
		t.Fatal(err)
	}
	if rows, _ := g.GetListVMSnapshotsByVMId("vm-a"); len(rows) != 2 || rows[0].ID != "g2" {
		t.Fatalf("global view polluted by host rows: %+v", rows)
	}
	set("vm-a") // empty replace wipes the record entirely
	if rows, err := g.GetListVMSnapshotsByVMId("vm-a"); err != nil || len(rows) != 0 {
		t.Fatalf("empty replace: %v %+v", err, rows)
	}
	// Host record survives untouched.
	rec, err := g.GetHostVMSnapshots(ctx, "h-A")
	if err != nil || rec.VMSnapshots["vm-a"][0].ID != "hr1" {
		t.Fatalf("host record: %v %+v", err, rec)
	}
}
