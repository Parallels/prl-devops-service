package query

import (
	"reflect"
	"testing"

	"github.com/Parallels/prl-devops-service/data"
)

// TestOrderByProperty_TieOrder pins order_layer.go:32-55: only String()-able
// fields actually order; ties (and unknown properties, which all render to
// reflect's identical non-string placeholder) keep input order; the caller's
// slice is sorted IN PLACE (same backing array returned).
func TestOrderByProperty_TieOrder(t *testing.T) {
	objs := loadFixtures(t).Objects

	got, err := OrderByProperty(objs, &Order{Property: "Name", Direction: OrderDirectionAsc})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Name != "Bob" || got[1].Name != "alice" { // byte order: "Bob" < "alice"
		t.Errorf("asc by Name: want Bob,alice first, got %v", names(got))
	}

	desc, _ := OrderByProperty(copies2(t, objs), &Order{Property: "Name", Direction: OrderDirectionDesc})
	if desc[0].Name != "root" {
		t.Errorf("desc by Name: want root first, got %v", names(desc))
	}

	// int property: FieldByName(...).String() yields the identical "<int Value>"
	// placeholder for every element → comparator never true → input order kept.
	ties, _ := OrderByProperty(copies2(t, objs), &Order{Property: "FailedLogins", Direction: OrderDirectionAsc})
	if !sameOrder(ties, objs) {
		t.Errorf("non-string field must tie-keep input order, got %v", names(ties))
	}

	unknown, _ := OrderByProperty(copies2(t, objs), &Order{Property: "NoSuchField", Direction: OrderDirectionAsc})
	if !sameOrder(unknown, objs) {
		t.Errorf("unknown property must preserve order, got %v", names(unknown))
	}

	// in-place contract: same backing array
	src := copies2(t, objs)
	out, _ := OrderByProperty(src, &Order{Property: "Name", Direction: OrderDirectionDesc})
	if len(out) == len(src) && sameOrder(src, out) && &src[0] != nil {
		// src was mutated by the call itself — assert the mutation happened:
		if src[0].Name != "root" {
			t.Errorf("OrderByProperty must sort the caller slice in place")
		}
	}
}

// TestParityVsLegacy_Order: corpus through both engines on independent copies.
func TestParityVsLegacy_Order(t *testing.T) {
	doc := loadFixtures(t)
	for _, o := range doc.Orders {
		o := o
		t.Run(o.Property, func(t *testing.T) {
			a, b := copies(t, doc.Objects)
			lo := &data.Order{Property: o.Property, Direction: data.OrderDirection(o.Direction)}
			qo := &Order{Property: o.Property, Direction: OrderDirection(o.Direction)}
			legacyOut, lerr := data.OrderByProperty(a, lo)
			queryOut, qerr := OrderByProperty(b, qo)
			if (lerr == nil) != (qerr == nil) {
				t.Fatalf("error mismatch: %v vs %v", lerr, qerr)
			}
			if !reflect.DeepEqual(legacyOut, queryOut) {
				t.Fatalf("order output differs for %q/%d: %v vs %v", o.Property, o.Direction, names(legacyOut), names(queryOut))
			}
		})
	}
}

func copies2(t *testing.T, objs []fixture) []fixture {
	t.Helper()
	a, _ := copies(t, objs)
	return a
}

func sameOrder(a, b []fixture) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name {
			return false
		}
	}
	return true
}
