package query

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/data"
)

func mustFixture(t *testing.T, name string) fixture {
	t.Helper()
	for _, f := range loadFixtures(t).Objects {
		if f.Name == name {
			return f
		}
	}
	t.Fatalf("fixture %q missing", name)
	return fixture{}
}

// TestParseFilter_Quirks pins every parse quirk of filter_layer.go:23-57.
func TestParseFilter_Quirks(t *testing.T) {
	cases := []struct {
		in      string
		want    *Filter
		wantErr string
	}{
		{in: "", want: nil},
		{in: "a=b", want: &Filter{Property: "a", Value: "b", Options: []FilterOptions{}}},
		{in: "a=b,i", want: &Filter{Property: "a", Value: "b", Options: []FilterOptions{FilterOptionsCaseInsensitive}}},
		{in: "a=b,I", want: &Filter{Property: "a", Value: "b", Options: []FilterOptions{FilterOptionsCaseInsensitive}}}, // EqualFold :41
		{in: "a=b,x", want: &Filter{Property: "a", Value: "b", Options: []FilterOptions{}}},                             // unknown option ignored
		{in: "a=b,c", want: &Filter{Property: "a", Value: "b", Options: []FilterOptions{}}},
		{in: "a=b,i,i", want: &Filter{Property: "a", Value: "b", Options: []FilterOptions{FilterOptionsCaseInsensitive, FilterOptionsCaseInsensitive}}}, // double (?i) prefix below
		{in: "a=b=c", wantErr: "invalid filter: a=b=c"},
		{in: "a", wantErr: "invalid filter: a"},
	}
	for _, c := range cases {
		got, err := ParseFilter(c.in)
		if c.wantErr != "" {
			if err == nil || err.Error() != c.wantErr {
				t.Errorf("ParseFilter(%q): want error %q, got %v", c.in, c.wantErr, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseFilter(%q): unexpected error %v", c.in, err)
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("ParseFilter(%q): want %+v, got %+v", c.in, c.want, got)
		}
	}
}

// TestFilterByProperty_RegexpNotEquality pins filter_layer.go:81-86: the
// value is an unanchored Go regexp, and compile errors surface verbatim.
func TestFilterByProperty_RegexpNotEquality(t *testing.T) {
	objs := loadFixtures(t).Objects // alice, Bob, carlos9, root, e

	// "e" matches every name containing e — NOT equality
	got, err := FilterByProperty(objs, &Filter{Property: "name", Value: "e", Options: []FilterOptions{}})
	if err != nil {
		t.Fatal(err)
	}
	if names(got)["alice"] == false || names(got)["e"] == false || names(got)["root"] == true {
		t.Errorf("unanchored regexp semantics broken: %v", names(got))
	}

	got, _ = FilterByProperty(objs, &Filter{Property: "email", Value: "^ro", Options: []FilterOptions{}})
	if len(got) != 1 || got[0].Name != "root" {
		t.Errorf("^ro anchor: want [root], got %v", names(got))
	}

	got, _ = FilterByProperty(objs, &Filter{Property: "name", Value: "[0-9]$", Options: []FilterOptions{}})
	if len(got) != 1 || got[0].Name != "carlos9" {
		t.Errorf("[0-9]$ anchor: want [carlos9], got %v", names(got))
	}

	_, err = FilterByProperty(objs, &Filter{Property: "name", Value: "(", Options: []FilterOptions{}})
	if err == nil || !strings.Contains(err.Error(), "error parsing regexp") {
		t.Errorf("invalid regexp must surface verbatim, got %v", err)
	}
}

// TestFilterCaseInsensitiveOption pins the (?i)-prefix semantics (NOT LOWER())
// and the repeated-option double-prefix from the :74-79 loop — including its
// cross-object accumulation, which stays idempotent in match results.
func TestFilterCaseInsensitiveOption(t *testing.T) {
	objs := loadFixtures(t).Objects
	f := &Filter{Property: "name", Value: "bob", Options: []FilterOptions{FilterOptionsCaseInsensitive}}
	got, err := FilterByProperty(objs, f)
	if err != nil || len(got) != 1 || got[0].Name != "Bob" {
		t.Fatalf("case-insensitive (?i): want [Bob], got %v err %v", names(got), err)
	}

	// "i,i" → (?i)(?i) applied once per object with a valid property: legal,
	// results must match the single-option outcome regardless of object count.
	double := &Filter{Property: "name", Value: "bob", Options: []FilterOptions{FilterOptionsCaseInsensitive, FilterOptionsCaseInsensitive}}
	got2, err := FilterByProperty(objs, double)
	if err != nil || !reflect.DeepEqual(got, got2) {
		t.Errorf("double (?i) prefix drift: %+v vs %+v (err %v)", got, got2, err)
	}
}

// TestFilterNestedJsonTag pins filter_layer.go:136-140 and 148-155: dotted
// paths resolve through field names (exact-case FieldByName at :149 — a real
// legacy quirk: a differently-cased first segment takes the EqualFold field
// branch but then hits the case-SENSITIVE FieldByName and returns nothing)
// AND through json tags (:151-155), case-insensitively.
func TestFilterNestedJsonTag(t *testing.T) {
	objs := loadFixtures(t).Objects
	for _, prop := range []string{"authentication.username", "AUTHENTICATION.USERNAME", "Auth.username", "Auth.USERNAME"} {
		f := &Filter{Property: prop, Value: "u-root", Options: []FilterOptions{}}
		got, err := FilterByProperty(objs, f)
		if err != nil || len(got) != 1 || got[0].Name != "root" {
			t.Errorf("%s: want [root], got %v (err %v)", prop, names(got), err)
		}
	}
	// the first-segment casing quirk itself (both engines identical — parity-verified):
	f := &Filter{Property: "auth.username", Value: "u-root", Options: []FilterOptions{}}
	got, _ := FilterByProperty(objs, f)
	if len(got) != 0 {
		t.Errorf("field-branch quirk: auth.* (not a tag, near-miss field name) must resolve to nothing, got %v", names(got))
	}
}

// TestFilterMapPaths pins filter_layer.go:115-127: map[string]string is a
// direct (case-sensitive) index; other maps scan keys with EqualFold.
func TestFilterMapPaths(t *testing.T) {
	objs := loadFixtures(t).Objects

	got, _ := FilterByProperty(objs, &Filter{Property: "Meta.k1", Value: "v1", Options: []FilterOptions{}})
	if len(got) != 1 || got[0].Name != "alice" {
		t.Errorf("direct string-map index: want [alice], got %v", names(got))
	}
	got, _ = FilterByProperty(objs, &Filter{Property: "Meta.K1", Value: "v1", Options: []FilterOptions{}})
	if len(got) != 0 {
		t.Errorf("string-map direct index must be case-SENSITIVE, got %v", names(got))
	}

	// map[string]int scans keys with EqualFold (:120-126) → both spellings hit alice
	a, _ := FilterByProperty(objs, &Filter{Property: "Counts.K1", Value: "^1$", Options: []FilterOptions{}})
	b, _ := FilterByProperty(objs, &Filter{Property: "Counts.k1", Value: "^1$", Options: []FilterOptions{}})
	if !reflect.DeepEqual(a, b) || len(a) != 1 || a[0].Name != "alice" {
		t.Errorf("EqualFold int-map scan: want [alice] both ways, got %v / %v", names(a), names(b))
	}
}

// TestFilterNonStringKinds pins :86 (%v rendering) and :165 (struct-kind
// never matches): int "3" matches 3; []string renders "[linux dev]";
// nil *string renders "<nil>"; the nested struct never matches.
func TestFilterNonStringKinds(t *testing.T) {
	objs := loadFixtures(t).Objects

	got, _ := FilterByProperty(objs, &Filter{Property: "failed_logins", Value: "^3$", Options: []FilterOptions{}})
	if len(got) != 2 || names(got)["Bob"] == false || names(got)["carlos9"] == false {
		t.Errorf("int via %%v: want [Bob carlos9], got %v", names(got))
	}

	got, _ = FilterByProperty(objs, &Filter{Property: "tags", Value: `\[linux dev\]`, Options: []FilterOptions{}})
	if len(got) != 1 || got[0].Name != "alice" {
		t.Errorf("[]string renders as \"[linux dev]\": want [alice], got %v", names(got))
	}

	got, _ = FilterByProperty(objs, &Filter{Property: "secret", Value: "^<nil>$", Options: []FilterOptions{}})
	if len(got) != 3 || names(got)["alice"] == false {
		t.Errorf("nil *string renders \"<nil>\": want the 3 nil-secrets, got %v", names(got))
	}

	got, _ = FilterByProperty(objs, &Filter{Property: "nested", Value: ".*", Options: []FilterOptions{}})
	if len(got) != 0 {
		t.Errorf("struct-kind property must never match, got %v", names(got))
	}
}

// TestGetProperty_PointerElement pins intentional delta 1 (doc.go): pointer
// element types resolve instead of panicking, and value-type behaviour is
// unchanged (parity with the legacy engine asserted there).
func TestGetProperty_PointerElement(t *testing.T) {
	objs := loadFixtures(t).Objects
	ptrs := make([]*fixture, 0, len(objs))
	for i := range objs {
		ptrs = append(ptrs, &objs[i])
	}
	ptrs = append(ptrs, nil) // nil element must not panic either

	got, err := FilterByProperty(ptrs, &Filter{Property: "name", Value: "^root$", Options: []FilterOptions{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "root" {
		t.Errorf("pointer elements: want [root], got %d objects (err %v)", len(got), err)
	}

	// value-type path unchanged: identical outcomes to the legacy engine
	legA, legB := copies(t, objs)
	legacyFilterParsed, err := data.ParseFilter("name=^root$")
	if err != nil {
		t.Fatal(err)
	}
	want, err := data.FilterByProperty(legA, legacyFilterParsed)
	if err != nil {
		t.Fatal(err)
	}
	mine, err := FilterByProperty(legB, &Filter{Property: "name", Value: "^root$", Options: []FilterOptions{}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, mine) {
		t.Errorf("value-type behaviour drifted from legacy")
	}
}

func names(fs []fixture) map[string]bool {
	out := make(map[string]bool, len(fs))
	for _, f := range fs {
		out[f.Name] = true
	}
	return out
}

// TestDivergence_PointerTraversal pins the ONE documented divergence (doc.go
// delta 1, task spec §2 "verify during implementation"): the legacy layer
// PANICS on any dotted path traversing a pointer field — verified live:
// "reflect: call of reflect.Value.Type on zero Value". The query package
// resolves those paths (reflect.Indirect). Parity corpus skips them.
func TestDivergence_PointerTraversal(t *testing.T) {
	objs := loadFixtures(t).Objects
	f, err := ParseFilter("authentication.username=u-root")
	if err != nil {
		t.Fatal(err)
	}

	panicked := func() (r any) {
		defer func() { r = recover() }()
		legacyParsed, _ := legacyFilter("authentication.username=u-root")
		_, _ = data.FilterByProperty(objs, legacyParsed)
		return nil
	}()
	if panicked == nil {
		t.Fatal("expected legacy panic on pointer traversal (verified behaviour), got none")
	}

	got, err := FilterByProperty(objs, f)
	if err != nil || len(got) != 1 || got[0].Name != "root" {
		t.Fatalf("query must resolve pointer path: want [root], got %v (err %v)", names(got), err)
	}
}

// TestParityVsLegacy_Filter runs the whole fixtures.json corpus through BOTH
// engines on independent copies — the differential AC-parity proof (D-4).
// Skips pointer-traversal filters: those are intentional divergence territory
// (TestDivergence_PointerTraversal), not a parity gap.
func TestParityVsLegacy_Filter(t *testing.T) {
	doc := loadFixtures(t)
	for _, in := range doc.Filters {
		in := in
		lower := strings.ToLower(in)
		if strings.HasPrefix(lower, "auth.") || strings.HasPrefix(lower, "authentication.") {
			continue
		}
		t.Run(in, func(t *testing.T) {
			a, b := copies(t, doc.Objects)

			lf, lerr := data.ParseFilter(in)
			qf, qerr := ParseFilter(in)
			if (lerr == nil) != (qerr == nil) {
				t.Fatalf("parse error mismatch: legacy %v, query %v", lerr, qerr)
			}
			if lerr != nil {
				if lerr.Error() != qerr.Error() {
					t.Fatalf("parse error text differs: %q vs %q", lerr, qerr)
				}
				return
			}
			if !reflect.DeepEqual(lf, translate(qf)) && !(lf == nil && qf == nil) {
				t.Fatalf("parsed struct differs: %+v vs %+v", lf, qf)
			}

			legacyOut, lerr2 := data.FilterByProperty(a, lf)
			queryOut, qerr2 := FilterByProperty(b, qf)
			if (lerr2 == nil) != (qerr2 == nil) {
				t.Fatalf("filter error mismatch: %v vs %v", lerr2, qerr2)
			}
			if lerr2 != nil && lerr2.Error() != qerr2.Error() {
				t.Fatalf("filter error text differs: %q vs %q", lerr2, qerr2)
			}
			if !reflect.DeepEqual(legacyOut, queryOut) {
				t.Fatalf("output differs for %q: %v vs %v", in, names(legacyOut), names(queryOut))
			}
		})
	}
}

// translate converts a query.Filter to a data.Filter for DeepEqual (the two
// Filter structs are field-identical by transcription).
func translate(f *Filter) *data.Filter {
	if f == nil {
		return nil
	}
	return &data.Filter{Property: f.Property, Value: f.Value, Options: translateOptions(f.Options)}
}

func translateOptions(opts []FilterOptions) []data.FilterOptions {
	if opts == nil {
		return nil
	}
	out := make([]data.FilterOptions, 0, len(opts))
	for _, o := range opts {
		out = append(out, data.FilterOptions(o))
	}
	return out
}
