package data

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
