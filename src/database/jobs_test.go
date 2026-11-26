package database

// Jobs on GORM (PBI #421 task 15): CRUD quirks, enrichment authz, bulk
// delete, recovery, clock-injected stale-window edges, watcher body.

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
)

func TestSQLJobs_CRUD(t *testing.T) {
	g := newSQLDB(t)
	root := tplRootCtx()
	bare := basecontext.NewBaseContext()

	// Cold guard.
	cold := NewGormDatabase(bare, nil)
	if _, err := cold.GetJobs(root); !errors.Is(err, interfaces.ErrDatabaseNotConnected) {
		t.Fatalf("cold GetJobs: %v", err)
	}

	// Zero rows: GetJobs NON-NIL, GetJobsByOwner nil (mirror shapes).
	all, err := g.GetJobs(root)
	if err != nil || all == nil || len(all) != 0 {
		t.Fatalf("empty GetJobs must be non-nil: %v %#v", err, all)
	}
	by, err := g.GetJobsByOwner(root, "alice")
	if err != nil || by != nil {
		t.Fatalf("empty GetJobsByOwner must be nil: %v %#v", err, by)
	}

	// Owner for the enrichment path (default roles must exist for CreateUser).
	seedDefaults(t, g, root)
	if _, err := g.CreateUser(root, models.User{ID: "u-1", Username: "alice", Name: "Alice A", Email: "alice@x.local", Password: fixtureSecret("alice")}); err != nil {
		t.Fatal(err)
	}

	j1, err := g.CreateJob(root, models.Job{ID: "j-1", Owner: "U-1", JobType: "vm.create", State: "bogus-state"})
	if err != nil {
		t.Fatal(err)
	}
	if j1.State != constants.JobStatePending || j1.CreatedAt == "" || j1.UpdatedAt == "" {
		t.Fatalf("create stamps/fix: %+v", j1)
	}
	if j1.OwnerName != "Alice A" { // returned view enriched (Fold ID)
		t.Fatalf("return enrichment: %+v", j1)
	}
	if _, err := g.CreateJob(root, models.Job{Owner: "u-1", JobType: "vm.delete", State: constants.JobStateRunning}); err != nil {
		t.Fatal(err) // minted id
	}
	if _, err := g.CreateJob(root, models.Job{ID: "j-3", Owner: "ghost-owner", JobType: "cfg", State: constants.JobStateCompleted}); err != nil {
		t.Fatal(err)
	}

	// Stored row keeps the caller's (empty) owner name — enrichment is
	// read-side only.
	row, err := g.GetJob(root, "J-1") // Fold probe
	if err != nil || row.OwnerName != "Alice A" {
		t.Fatalf("read enrichment: %v %+v", err, row)
	}
	// Enrichment is NOT ctx-gated: legacy GetUser wraps GetUsers, which has
	// no authorisation at the data layer (users.go:30-43 is a plain
	// property filter) — even a bare ctx enriches. Pinned deliberately.
	bareRow, err := g.GetJob(bare, "j-1")
	if err != nil || bareRow.OwnerName != "Alice A" {
		t.Fatalf("data-layer enrichment must not be ctx-gated: %v %+v", err, bareRow)
	}
	if _, err := g.GetJob(root, "nope"); !errors.Is(err, interfaces.ErrJobNotFound) {
		t.Fatalf("get miss: %v", err)
	}
	by, err = g.GetJobsByOwner(root, "ALICE") // Fold on owner (:44)
	if err != nil || len(by) != 0 {           // owner is "U-1"/"u-1", not the username
		t.Fatalf("owner filter: %v %+v", err, by)
	}
	if by, err := g.GetJobsByOwner(root, "u-1"); err != nil || len(by) != 2 {
		t.Fatalf("owner filter id: %v %+v", err, by)
	}

	// UpdateJob: EXACT id only — a cased probe misses (:110).
	if err := g.UpdateJob(root, models.Job{ID: "J-1"}); !errors.Is(err, interfaces.ErrJobNotFound) {
		t.Fatalf("update must be EXACT id: %v", err)
	}
	steps := []models.JobStep{{Name: "s1", Weight: 1}, {Name: "s2", Weight: 2.5}}
	if err := g.UpdateJob(root, models.Job{ID: "j-1", State: constants.JobStateRunning, Message: "m",
		Progress: 42, Result: "r", ResultRecordId: "rr", ResultRecordName: "rrn",
		ResultRecordLinkId: "rrl", ResultRecordType: "rrt", Error: "e", Steps: steps,
		IsOrchestratorJob: true}); err != nil {
		t.Fatal(err)
	}
	up, _ := g.GetJob(root, "j-1")
	if up.Progress != 42 || up.Result != "r" || up.ResultRecordLinkId != "rrl" || !up.IsOrchestratorJob ||
		!reflect.DeepEqual(up.Steps, steps) || up.JobType != "vm.create" { // untouched fields survive
		t.Fatalf("update set/steps serializer: %+v", up)
	}

	// DeleteJob: "" silently nil; Fold delete; miss errors.
	if err := g.DeleteJob(root, ""); err != nil {
		t.Fatalf("empty delete must be nil: %v", err)
	}
	if err := g.DeleteJob(root, "J-3"); err != nil { // Fold on delete (:142)
		t.Fatal(err)
	}
	if err := g.DeleteJob(root, "j-3"); !errors.Is(err, interfaces.ErrJobNotFound) {
		t.Fatalf("double delete: %v", err)
	}
	if all, _ := g.GetJobs(root); len(all) != 2 {
		t.Fatalf("post-delete count: %+v", all)
	}
}

func TestSQLJobs_BulkDelete(t *testing.T) {
	g := newSQLDB(t)
	root := tplRootCtx()
	seed := []models.Job{
		{ID: "b1", State: constants.JobStatePending},
		{ID: "b2", State: constants.JobStateRunning},
		{ID: "b3", State: constants.JobStateCompleted},
		{ID: "b4", State: constants.JobStateFailed},
	}
	for _, j := range seed {
		if _, err := g.CreateJob(root, j); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.DeleteJobsByState(root, constants.JobStatePending, constants.JobStateRunning); err != nil {
		t.Fatal(err)
	}
	left, _ := g.GetJobs(root)
	if len(left) != 2 || left[0].ID != "b3" || left[1].ID != "b4" {
		t.Fatalf("bulk delete: %+v", left)
	}
	if err := g.DeleteJobsByState(root); err != nil { // no states = keep all
		t.Fatal(err)
	}
	if left, _ := g.GetJobs(root); len(left) != 2 {
		t.Fatalf("empty-arg wipe: %+v", left)
	}
}

func TestSQLJobs_StaleDetect(t *testing.T) {
	g := newSQLDB(t)
	root := tplRootCtx()
	t0 := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	g.nowFn = func() time.Time { return t0 } // default timeout: 5 minutes

	stampAt := func(d time.Duration) string { return t0.Add(d).UTC().Format(time.RFC3339Nano) }
	seed := []models.Job{
		{ID: "s-stale", State: constants.JobStateRunning},
		{ID: "s-boundary", State: constants.JobStatePending}, // exactly at the timeout: NOT stale
		{ID: "s-future", State: constants.JobStateRunning},
		{ID: "s-parsefail", State: constants.JobStatePending},
		{ID: "s-done", State: constants.JobStateCompleted},
	}
	for _, j := range seed {
		if _, err := g.CreateJob(root, j); err != nil {
			t.Fatal(err)
		}
	}
	raw := func(id, stamp string) {
		if err := g.gormDB.Table("jobs").Where("id = ?", id).Update("updated_at", stamp).Error; err != nil {
			t.Fatal(err)
		}
	}
	raw("s-stale", stampAt(-6*time.Minute))
	raw("s-boundary", stampAt(-5*time.Minute))
	raw("s-future", stampAt(time.Hour))
	raw("s-parsefail", "not-a-timestamp")
	raw("s-done", stampAt(-90*time.Minute))

	var fired []models.Job
	g.SetJobTimeoutHandler(func(j models.Job) { fired = append(fired, j) })
	g.DetectStaleJobs(root)

	jobs, _ := g.GetJobs(root)
	byID := map[string]models.Job{}
	for _, j := range jobs {
		byID[j.ID] = j
	}
	stale := byID["s-stale"]
	if stale.State != constants.JobStateFailed ||
		!strings.HasPrefix(stale.Error, "Job timed out after 5 minutes without a progress update. Last update: ") ||
		stale.Message != stale.Error || stale.UpdatedAt != t0.UTC().Format(time.RFC3339Nano) {
		t.Fatalf("stale mark: %+v", stale)
	}
	for _, id := range []string{"s-boundary", "s-future", "s-parsefail", "s-done"} {
		if byID[id].State == constants.JobStateFailed {
			t.Fatalf("%s must survive: %+v", id, byID[id])
		}
	}
	if len(fired) != 1 || fired[0].ID != "s-stale" || fired[0].State != constants.JobStateFailed {
		t.Fatalf("handler: %+v", fired)
	}
	// Progress/steps survive the timeout mark (legacy keeps them).
	if err := g.UpdateJob(root, models.Job{ID: "s-stale", State: constants.JobStateFailed, Progress: 7}); err != nil {
		t.Fatal(err)
	}
	g.DetectStaleJobs(root) // second pass: nothing newly stale
	if len(fired) != 1 {
		t.Fatalf("handler refire: %+v", fired)
	}
}

func TestSQLJobs_Recover(t *testing.T) {
	g := newSQLDB(t)
	root := tplRootCtx()
	for _, j := range []models.Job{
		{ID: "rc1", State: constants.JobStateRunning},
		{ID: "rc2", State: constants.JobStatePending},
		{ID: "rc3", State: constants.JobStateCompleted, Result: "ok"},
		{ID: "rc4", State: constants.JobStateFailed, Error: "old"},
	} {
		if _, err := g.CreateJob(root, j); err != nil {
			t.Fatal(err)
		}
	}
	g.RecoverOngoingJobs(root)
	jobs, _ := g.GetJobs(root)
	for _, j := range jobs {
		switch j.ID {
		case "rc1", "rc2":
			if j.State != constants.JobStateFailed || j.Error != "Service crashed or was restarted" {
				t.Fatalf("recover %s: %+v", j.ID, j)
			}
		default:
			if j.Error == "Service crashed or was restarted" {
				t.Fatalf("recover touched %s: %+v", j.ID, j)
			}
		}
	}
}

func TestSQLJobs_WatcherBody(t *testing.T) {
	// Zero goroutines are started by the port; the body is a plain
	// cancellable function — exercised directly, no sleeps.
	g := newSQLDB(t)
	root := tplRootCtx()
	done := make(chan bool, 1)
	g.staleCancel = make(chan bool) // pre-assigned: Stop/Start rendezvous without nil race
	go func() { done <- g.runStaleJobWatcher(root) }()
	g.StopStaleJobWatcher() // rendezvous with the select — deterministic
	if !<-done {
		t.Fatal("watcher must report cancelled")
	}
	// Cancelled channel up front returns immediately.
	g2 := newSQLDB(t)
	g2.staleCancel = make(chan bool)
	close(g2.staleCancel)
	if !g2.runStaleJobWatcher(root) {
		t.Fatal("closed cancel")
	}
}

func TestSQLJobs_RecoverByNode(t *testing.T) {
	g := newSQLDB(t)
	root := tplRootCtx()
	for _, j := range []models.Job{
		{ID: "nd-a1", OwnerNode: "node-a", State: constants.JobStateRunning},
		{ID: "nd-a2", OwnerNode: "node-a", State: constants.JobStatePending},
		{ID: "nd-a-old", OwnerNode: "node-a-former-name", State: constants.JobStateRunning},
		{ID: "nd-b1", OwnerNode: "node-b", State: constants.JobStateRunning},
		{ID: "nd-legacy", State: constants.JobStateRunning},
		{ID: "nd-done", OwnerNode: "node-a", State: constants.JobStateCompleted, Result: "ok"},
	} {
		if _, err := g.CreateJob(root, j); err != nil {
			t.Fatal(err)
		}
	}

	// Recovering must not touch another replica's work: that job is being run
	// over there right now, and the create it describes is holding capacity for
	// a VM that exists. The list also carries an empty, a blank and a repeat,
	// which must not widen the scope or the statement.
	g.RecoverOngoingJobsForNode(root, []string{"node-a", "", "   ", "node-a"})

	want := map[string]constants.JobState{
		"nd-a1":     constants.JobStateFailed, // mine, and genuinely abandoned
		"nd-a2":     constants.JobStateFailed,
		"nd-legacy": constants.JobStateFailed, // predates the field: nobody's is everybody's
		"nd-a-old":  constants.JobStateRunning,
		"nd-b1":     constants.JobStateRunning,
		"nd-done":   constants.JobStateCompleted,
	}
	jobs, _ := g.GetJobs(root)
	for _, j := range jobs {
		if state, ok := want[j.ID]; ok && j.State != state {
			t.Fatalf("recover by node %s: state %q want %q", j.ID, j.State, state)
		}
	}

	// An identity this replica used before is still its own work. Without the
	// alias, a replica that had stamped rows under its hostname before identity
	// moved to the machine id would read them as someone else's and leave them
	// running with nobody to finish them.
	g.RecoverOngoingJobsForNode(root, []string{"node-a", "node-a-former-name"})
	jobs, _ = g.GetJobs(root)
	for _, j := range jobs {
		if j.ID == "nd-a-old" && j.State != constants.JobStateFailed {
			t.Fatalf("former identity not treated as own work: %+v", j)
		}
	}

	// An unidentifiable replica recovers nothing at all. Sweeping the cluster
	// because the node id failed to resolve is the destructive reading of this
	// method, and the stale-job watcher will still reap what is left.
	g.RecoverOngoingJobsForNode(root, []string{})
	g.RecoverOngoingJobsForNode(root, []string{"", "  "})
	jobs, _ = g.GetJobs(root)
	for _, j := range jobs {
		if j.ID == "nd-b1" && j.State != constants.JobStateRunning {
			t.Fatalf("empty node id swept a foreign job: %+v", j)
		}
	}
}

// TestSQLJobs_RenameOwner covers the bookkeeping behind a node rename through the
// settings service. A rename that leaves rows under the retired name is not
// cosmetic: recovery matches on owner_node, so those jobs belong to a node that no
// longer exists and nobody will ever finish them.
func TestSQLJobs_RenameOwner(t *testing.T) {
	g := newSQLDB(t)
	root := tplRootCtx()
	for _, j := range []models.Job{
		{ID: "rn-current", OwnerNode: "node-a", State: constants.JobStateRunning},
		{ID: "rn-former", OwnerNode: "node-a-old", State: constants.JobStateRunning},
		{ID: "rn-foreign", OwnerNode: "node-b", State: constants.JobStateRunning},
		{ID: "rn-ownerless", State: constants.JobStateRunning},
	} {
		if _, err := g.CreateJob(root, j); err != nil {
			t.Fatal(err)
		}
	}

	owner := func(id string) string {
		t.Helper()
		job, err := g.GetJob(root, id)
		if err != nil {
			t.Fatalf("get %s: %v", id, err)
		}

		return job.OwnerNode
	}

	if err := g.RenameJobOwner(root, "node-a-old", "node-a"); err != nil {
		t.Fatal(err)
	}
	if got := owner("rn-former"); got != "node-a" {
		t.Fatalf("renamed row: owner %q", got)
	}

	// A rename moves this replica's rows and nothing else. Reaching into another
	// replica's work would attribute live creates to the wrong owner, and the
	// ownerless rows are the ones recovery deliberately adopts — an empty `from`
	// would have matched both.
	if got := owner("rn-foreign"); got != "node-b" {
		t.Fatalf("foreign row touched: owner %q", got)
	}
	if got := owner("rn-ownerless"); got != "" {
		t.Fatalf("ownerless row re-attributed: owner %q", got)
	}

	// The refusals. Blank-looking arguments are the dangerous ones, not the
	// harmless no-ops they appear to be.
	for _, refusal := range [][2]string{{"", "node-a"}, {"   ", "node-a"}, {"node-a", ""}, {"node-a", "   "}, {"node-a", "node-a"}} {
		if err := g.RenameJobOwner(root, refusal[0], refusal[1]); !errors.Is(err, interfaces.ErrNodeIdentityUnusable) {
			t.Fatalf("rename %q -> %q: want refusal, got %v", refusal[0], refusal[1], err)
		}
	}

	// Idempotent, which is what makes a rename interrupted between the two tables
	// safe to simply run again.
	if err := g.RenameJobOwner(root, "node-a-old", "node-a"); err != nil {
		t.Fatalf("re-run of a finished rename: %v", err)
	}
	if err := g.RenameJobOwner(root, "node-b", "node-c"); err != nil {
		t.Fatal(err)
	}
	if got := owner("rn-foreign"); got != "node-c" {
		t.Fatalf("second rename: owner %q", got)
	}
}
