package data

import (
	"reflect"
	"sort"
)

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v1.3.0.
type OrderDirection int

const (
	OrderDirectionAsc OrderDirection = iota
	OrderDirectionDesc
)

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v1.3.0.
type Order struct {
	Property  string
	Direction OrderDirection
}

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v1.3.0.
func OrderByProperty[T interface{}](objects []T, order *Order) ([]T, error) {
	if order == nil {
		return objects, nil
	}

	if order.Direction == OrderDirectionAsc {
		return orderByPropertyAsc(objects, order.Property)
	} else {
		return orderByPropertyDesc(objects, order.Property)
	}
}

func orderByPropertyAsc[T interface{}](objects []T, property string) ([]T, error) {
	rv := reflect.ValueOf(objects)
	// swap := reflect.Swapper(objects)

	// #425 T6: ties resolve by unique ID ascending (see desc twin). SliceStable
	// keeps any residual tie (no ID field) input-ordered.
	sort.SliceStable(objects, func(i, j int) bool {
		iVal := reflect.Indirect(rv.Index(i)).FieldByName(property).String()
		jVal := reflect.Indirect(rv.Index(j)).FieldByName(property).String()
		if iVal != jVal {
			return iVal < jVal
		}
		return idTieLess(rv, i, j)
	})

	return objects, nil
}

func orderByPropertyDesc[T interface{}](objects []T, property string) ([]T, error) {
	rv := reflect.ValueOf(objects)
	// swap := reflect.Swapper(objects)

	// #425 T6: sort.Slice over pre-shuffled authz input (the map-based
	// filter/authz wrap scrambles the seed order) left same-key permutations
	// engine-run-specific, so every l[0]/"delete first row" addressing in
	// the parity suite could diverge between engines. Deterministic total
	// order now: property key first, then unique ID ascending whenever the
	// element exposes an ID string field; absent ID preserves the old
	// return-false (residual ties fall back to stability). Note: equal
	// UpdatedAt strings are rare (RFC3339Nano stamps) — the deeper string-
	// vs-chronological flaw of that format is backlog T14, production scope.
	sort.SliceStable(objects, func(i, j int) bool {
		iVal := reflect.Indirect(rv.Index(i)).FieldByName(property).String()
		jVal := reflect.Indirect(rv.Index(j)).FieldByName(property).String()
		if iVal != jVal {
			return iVal > jVal
		}
		return idTieLess(rv, i, j)
	})

	return objects, nil
}

// idTieLess compares the ID string field of objects[i] vs objects[j]; false
// when either element lacks a string ID (residual tie) or IDs are equal.
func idTieLess(rv reflect.Value, i, j int) bool {
	iID := reflect.Indirect(rv.Index(i)).FieldByName("ID")
	jID := reflect.Indirect(rv.Index(j)).FieldByName("ID")
	if !iID.IsValid() || !jID.IsValid() || iID.Kind() != reflect.String || jID.Kind() != reflect.String {
		return false
	}
	return iID.String() < jID.String()
}
