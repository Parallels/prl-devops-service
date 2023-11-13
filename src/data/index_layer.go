package data

import (
	"reflect"
	"strings"
)

func GetRecordIndex[T interface{}](objects []T, propertyIndexer string, propertyValue string) (int, error) {
	index := -1
	propertyName := propertyIndexer

	for i, obj := range objects {
		objValue := reflect.ValueOf(obj)
		property := getProperty(objValue, propertyName)
		if propertyIsValid(property) {
			switch property.Interface().(type) {
			case string:
				if strings.EqualFold(property.Interface().(string), propertyValue) {
					index = i
					break
				}
			case bool:
				index = -1
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
				if property.Interface() == propertyValue {
					index = i
					break
				}
			default:
				index = -1
			}
		}
	}

	return index, nil
}
