package query

import (
	"testing"

	"github.com/Parallels/prl-devops-service/data"
)

// TestGetRecordIndex_DeadBranches pins index_layer.go:16-31: string is the
// only live comparison (EqualFold, first match wins); bool and numeric
// properties return -1 even on textual match — legacy dead code, preserved.
func TestGetRecordIndex_DeadBranches(t *testing.T) {
	objs := loadFixtures(t).Objects

	idx, err := GetRecordIndex(objs, "Name", "BOB") // EqualFold
	if err != nil || idx != 1 || objs[idx].Name != "Bob" {
		t.Errorf("string EqualFold first-match: want 1, got %d (%v)", idx, err)
	}

	idx, _ = GetRecordIndex(objs, "FailedLogins", "3") // int vs string — never equal
	if idx != -1 {
		t.Errorf("numeric dead branch: want -1, got %d", idx)
	}

	idx, _ = GetRecordIndex(objs, "Active", "true") // bool branch resets to -1
	if idx != -1 {
		t.Errorf("bool dead branch: want -1, got %d", idx)
	}

	idx, _ = GetRecordIndex(objs, "Name", "nobody")
	if idx != -1 {
		t.Errorf("no match: want -1, got %d", idx)
	}
}

// TestParityVsLegacy_Index: corpus through both engines.
func TestParityVsLegacy_Index(t *testing.T) {
	doc := loadFixtures(t)
	for _, ix := range doc.Indexes {
		ix := ix
		t.Run(ix.Property+"="+ix.Value, func(t *testing.T) {
			a, b := copies(t, doc.Objects)
			lidx, lerr := data.GetRecordIndex(a, ix.Property, ix.Value)
			qidx, qerr := GetRecordIndex(b, ix.Property, ix.Value)
			if (lerr == nil) != (qerr == nil) || lerr != nil && lerr.Error() != qerr.Error() {
				t.Fatalf("error mismatch: %v vs %v", lerr, qerr)
			}
			if lidx != qidx {
				t.Fatalf("index differs: %d vs %d", lidx, qidx)
			}
		})
	}
}
