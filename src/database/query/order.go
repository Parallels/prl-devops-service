package query

// Transcription of data/order_layer.go:8-55 (PBI #421 task 02, D-4).
// Verbatim except the #425 T6 determinism patch, which is now part of the
// parity CONTRACT and must stay byte-identical in behaviour to the legacy
// twin (data/order_layer.go): in-place sort of the caller's slice,
// FieldByName(...).String() keying (non-string/unknown fields get reflect's
// non-string placeholder — effectively ties), ties resolved by the unique ID
// string field ascending when present, SliceStable so any residual tie stays
// input-ordered.

import (
	"reflect"
	"sort"
)

type OrderDirection int

const (
	OrderDirectionAsc OrderDirection = iota
	OrderDirectionDesc
)

type Order struct {
	Property  string
	Direction OrderDirection
}

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

	// #425 T6: mirrors the legacy twin exactly — property key, then unique ID
	// ascending; SliceStable keeps residual ties input-ordered.
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
