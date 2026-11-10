package query

// Shared differential-test fixtures (PBI #421 task 02). Package-internal so
// the goldens can also exercise getProperty directly. The parity tests run
// the legacy data.* functions against INDEPENDENT copies of these fixtures.

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Parallels/prl-devops-service/data"
)

type fixtureAuth struct {
	Username string `json:"username"`
	Kind     string `json:"kind"`
}

type fixtureNested struct {
	Deep string `json:"deep"`
}

type fixture struct {
	Name         string            `json:"name"`
	Email        string            `json:"email"`
	Active       bool              `json:"active"`
	FailedLogins int               `json:"failed_logins"`
	Tags         []string          `json:"tags"`
	Secret       *string           `json:"secret"`
	Auth         *fixtureAuth      `json:"authentication"`
	Meta         map[string]string `json:"meta"`
	Counts       map[string]int    `json:"counts"`
	Nested       fixtureNested     `json:"nested"`
}

type fixturesDoc struct {
	Objects []fixture           `json:"objects"`
	Filters []string            `json:"filters"`
	Orders  []fixturesOrder     `json:"orders"`
	Indexes []fixturesIndexPair `json:"indexes"`
}

type fixturesOrder struct {
	Property  string `json:"property"`
	Direction int    `json:"direction"`
}

type fixturesIndexPair struct {
	Property string `json:"property"`
	Value    string `json:"value"`
}

func loadFixtures(t *testing.T) fixturesDoc {
	t.Helper()
	raw, err := os.ReadFile("testdata/fixtures.json")
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	var doc fixturesDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse fixtures: %v", err)
	}
	return doc
}

// copies returns two independent deep copies of the fixture objects (JSON
// round-trip), so each engine sorts over a pristine slice.
func copies(t *testing.T, objs []fixture) ([]fixture, []fixture) {
	t.Helper()
	raw, err := json.Marshal(objs)
	if err != nil {
		t.Fatalf("marshal fixtures: %v", err)
	}
	a := make([]fixture, 0, len(objs))
	b := make([]fixture, 0, len(objs))
	if err := json.Unmarshal(raw, &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &b); err != nil {
		t.Fatal(err)
	}
	return a, b
}

// legacyFilter is a compile-named alias keeping the differential intent visible.
var legacyFilter = data.ParseFilter
