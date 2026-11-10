package query

// Transcription of data/order_layer.go:8-55 (PBI #421 task 02, D-4).
// Verbatim: in-place sort of the caller's slice, FieldByName(...).String()
// keying (non-string/unknown fields get reflect's non-string placeholder —
// effectively ties), sort.Slice's unstable-but-input-ordered ties.

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

	sort.Slice(objects, func(i, j int) bool {
		iVal := reflect.Indirect(rv.Index(i)).FieldByName(property).String()
		jVal := reflect.Indirect(rv.Index(j)).FieldByName(property).String()
		return iVal < jVal
	})

	return objects, nil
}

func orderByPropertyDesc[T interface{}](objects []T, property string) ([]T, error) {
	rv := reflect.ValueOf(objects)
	// swap := reflect.Swapper(objects)

	sort.Slice(objects, func(i, j int) bool {
		iVal := reflect.Indirect(rv.Index(i)).FieldByName(property).String()
		jVal := reflect.Indirect(rv.Index(j)).FieldByName(property).String()
		return iVal > jVal
	})

	return objects, nil
}
