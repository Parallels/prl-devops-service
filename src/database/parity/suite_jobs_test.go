package parity

// DomainSuite "jobs" (PBI #421 task 15): CRUD shapes (GetJobs NON-NIL at
// zero vs GetJobsByOwner nil), owner enrichment through the data-layer
// (ctx-free) GetUser lookup, UpdateJob EXACT-id quirk (a cased probe
// 404s while GetJob finds the row), DeleteJob("") silent nil + Fold delete,
// variadic DeleteJobsByState incl. the no-states no-op, RecoverOngoingJobs
// state set + verbatim error string, and the DetectStaleJobs no-op path
// (stale-window edges live in the clock-injected SQL units; the legacy
// engine has no clock hook to travel — ER-5 keeps it untouched).

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
)

func normaliseJobs(in []models.Job) []models.Job {
	if len(in) == 0 {
		return nil
	}
	out := append([]models.Job(nil), in...)
	for i := range out {
		if out[i].CreatedAt != "" {
			out[i].CreatedAt = "SET"
		}
		if out[i].UpdatedAt != "" {
			out[i].UpdatedAt = "SET"
		}
	}
	return out
}

func seedJobs(t testing.TB, db datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	for _, r := range constants.DefaultRoles {
		if _, err := db.CreateRole(ctx, models.Role{Name: r}); err != nil {
			t.Fatalf("seed default role %s: %v", r, err)
		}
	}
	for _, c := range constants.DefaultClaims {
		if _, err := db.CreateClaim(ctx, models.Claim{Name: c}); err != nil {
			t.Fatalf("seed default claim %s: %v", c, err)
		}
	}
	if _, err := db.CreateUser(ctx, models.User{ID: "u-1", Username: "alice", Email: "alice@x.local", Name: "Alice A", Password: fixtureSecret("alice")}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	for _, j := range []models.Job{
		{ID: "j-1", Owner: "u-1", JobType: "vm.create", State: constants.JobStatePending, Progress: 10},
		{ID: "j-2", Owner: "U-1", JobType: "vm.delete", State: constants.JobStateRunning}, // cased owner: Fold enrichment probe
		{ID: "j-3", Owner: "ghost", JobType: "cfg", State: constants.JobStateCompleted, Result: "done"},
		{ID: "j-4", JobType: "misc", State: constants.JobStatePending}, // ownerless
	} {
		if _, err := db.CreateJob(ctx, j); err != nil {
			t.Fatalf("seed job %s: %v", j.ID, err)
		}
	}
}

func TestParityDomain_Jobs(t *testing.T) {
	RunMatrix(t, DomainSuite{
		Name: "jobs",
		Seed: seedJobs,
		Checks: []func(t testing.TB, legacy, sql datainterfaces.DatabaseService){
			checkJobsLists,
			checkJobsMutations,
			checkJobsRecovery,
		},
	})
}

func checkJobsLists(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()

	all, lerr := legacy.GetJobs(ctx)
	allS, serr := sql.GetJobs(ctx)
	assertSameError(t, "GetJobs", lerr, serr)
	AssertSameList(t, "GetJobs", normaliseJobs(all), normaliseJobs(allS))
	if (all == nil) != (allS == nil) { // nil-vs-empty shape parity (:25 non-nil)
		t.Fatalf("GetJobs nil-shape drift: legacy=%v sql=%v", all == nil, allS == nil)
	}

	for _, owner := range []string{"u-1", "U-1", "ALICE", "ghost", "", "Nope"} {
		l, lerr := legacy.GetJobsByOwner(ctx, owner)
		s, serr := sql.GetJobsByOwner(ctx, owner)
		assertSameError(t, "GetJobsByOwner["+owner+"]", lerr, serr)
		AssertSameList(t, "GetJobsByOwner["+owner+"]", normaliseJobs(l), normaliseJobs(s))
		if (l == nil) != (s == nil) { // nil-shape (:42 var-scan)
			t.Fatalf("GetJobsByOwner[%s] nil-shape drift: %v vs %v", owner, l == nil, s == nil)
		}
	}

	for _, probe := range []string{"j-1", "J-1", "j-none", ""} {
		l, lerr := legacy.GetJob(ctx, probe)
		s, serr := sql.GetJob(ctx, probe)
		assertSameError(t, "GetJob["+probe+"]", lerr, serr)
		var ln, sn []models.Job
		if l != nil {
			ln = normaliseJobs([]models.Job{*l})
		}
		if s != nil {
			sn = normaliseJobs([]models.Job{*s})
		}
		AssertSameList(t, "GetJob["+probe+"]", ln, sn)
	}
}

func checkJobsMutations(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	runPair := func(name string, fn func(db datainterfaces.DatabaseService) error) {
		le, se := fn(legacy), fn(sql)
		assertSameError(t, name, le, se)
	}

	runPair("UpdateJob-set", func(db datainterfaces.DatabaseService) error {
		return db.UpdateJob(ctx, models.Job{ID: "j-2", State: constants.JobStateRunning, Message: "step",
			Progress: 55, Result: "r2", ResultRecordId: "rr", ResultRecordName: "rrn",
			ResultRecordLinkId: "rrl", ResultRecordType: "rrt", Error: "e2",
			Steps:             []models.JobStep{{Name: "s1", Weight: 1}, {Name: "s2", Weight: 2.5}},
			IsOrchestratorJob: true})
	})
	// EXACT-id quirk: the cased probe 404s on BOTH engines.
	runPair("UpdateJob-exact", func(db datainterfaces.DatabaseService) error {
		return db.UpdateJob(ctx, models.Job{ID: "J-1", State: constants.JobStateFailed})
	})
	for _, probe := range []string{"j-2", "j-1"} {
		l, lerr := legacy.GetJob(ctx, probe)
		s, serr := sql.GetJob(ctx, probe)
		assertSameError(t, "GetJob-after-update["+probe+"]", lerr, serr)
		var ln, sn []models.Job
		if l != nil {
			ln = normaliseJobs([]models.Job{*l})
		}
		if s != nil {
			sn = normaliseJobs([]models.Job{*s})
		}
		AssertSameList(t, "GetJob-after-update["+probe+"]", ln, sn)
	}

	// Minted ids differ: compare everything but identity, then remove.
	lc, lerr := legacy.CreateJob(ctx, models.Job{Owner: "u-1", JobType: "minted", State: constants.JobStatePending})
	sc, serr := sql.CreateJob(ctx, models.Job{Owner: "u-1", JobType: "minted", State: constants.JobStatePending})
	assertSameError(t, "CreateJob", lerr, serr)
	if lc == nil || sc == nil || lc.ID == "" || sc.ID == "" {
		t.Fatalf("minted create: %+v %+v", lc, sc)
	}
	lmint, smint := lc.ID, sc.ID // real minted ids, captured before blanking
	lc.ID, sc.ID = "MINT", "MINT"
	AssertSameList(t, "CreateJob-view", normaliseJobs([]models.Job{*lc}), normaliseJobs([]models.Job{*sc}))
	runPair("DeleteJob-mint", func(db datainterfaces.DatabaseService) error {
		if db == legacy {
			return db.DeleteJob(ctx, lmint)
		}
		return db.DeleteJob(ctx, smint)
	})

	runPair("DeleteJob-empty", func(db datainterfaces.DatabaseService) error { return db.DeleteJob(ctx, "") })
	runPair("DeleteJob-fold", func(db datainterfaces.DatabaseService) error { return db.DeleteJob(ctx, "J-3") })
	runPair("DeleteJob-miss", func(db datainterfaces.DatabaseService) error { return db.DeleteJob(ctx, "j-3") })
	runPair("DeleteJobsByState-pending", func(db datainterfaces.DatabaseService) error {
		return db.DeleteJobsByState(ctx, constants.JobStatePending)
	})
	runPair("DeleteJobsByState-none-args", func(db datainterfaces.DatabaseService) error {
		return db.DeleteJobsByState(ctx)
	})
	l, lerr := legacy.GetJobs(ctx)
	s, serr := sql.GetJobs(ctx)
	assertSameError(t, "GetJobs-after-bulk", lerr, serr)
	AssertSameList(t, "GetJobs-after-bulk", normaliseJobs(l), normaliseJobs(s))
}

func checkJobsRecovery(t testing.TB, legacy, sql datainterfaces.DatabaseService) {
	ctx := basecontext.NewBaseContext()
	for _, j := range []models.Job{
		{ID: "r-1", Owner: "u-1", State: constants.JobStateRunning},
		{ID: "r-2", State: constants.JobStatePending},
		{ID: "r-3", State: constants.JobStateCompleted, Result: "keep"},
	} {
		if _, err := legacy.CreateJob(ctx, j); err != nil {
			t.Fatalf("legacy seed %s: %v", j.ID, err)
		}
		if _, err := sql.CreateJob(ctx, j); err != nil {
			t.Fatalf("sql seed %s: %v", j.ID, err)
		}
	}
	legacy.RecoverOngoingJobs(ctx)
	sql.RecoverOngoingJobs(ctx)
	l, lerr := legacy.GetJobs(ctx)
	s, serr := sql.GetJobs(ctx)
	assertSameError(t, "GetJobs-after-recover", lerr, serr)
	AssertSameList(t, "GetJobs-after-recover", normaliseJobs(l), normaliseJobs(s))

	// Fresh stamps everywhere (just created): the stale pass must mark
	// NOTHING and must not fire the registered handler on either engine.
	lfired, sfired := 0, 0
	legacy.SetJobTimeoutHandler(func(models.Job) { lfired++ })
	sql.SetJobTimeoutHandler(func(models.Job) { sfired++ })
	legacy.DetectStaleJobs(ctx)
	sql.DetectStaleJobs(ctx)
	if lfired != 0 || sfired != 0 {
		t.Fatalf("handler fired on fresh jobs: legacy=%d sql=%d", lfired, sfired)
	}
	l2, lerr2 := legacy.GetJobs(ctx)
	s2, serr2 := sql.GetJobs(ctx)
	assertSameError(t, "GetJobs-after-stale", lerr2, serr2)
	AssertSameList(t, "GetJobs-after-stale", normaliseJobs(l2), normaliseJobs(s2))
}
