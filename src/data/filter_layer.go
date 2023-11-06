package data

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type Filter struct {
	Property string
	Value    string
	Options  []FilterOptions
}

type FilterOptions int

const (
	FilterOptionsNone FilterOptions = iota
	FilterOptionsCaseInsensitive
)

func ParseFilter(filter string) (*Filter, error) {
	if filter == "" {
		return nil, nil
	}
	optionParts := strings.Split(filter, ",")
	if len(optionParts) == 1 {
		filterParts := strings.Split(filter, "=")
		if len(filterParts) != 2 {
			return nil, fmt.Errorf("invalid filter: %s", filter)
		}
		return &Filter{
			Property: filterParts[0],
			Value:    filterParts[1],
			Options:  make([]FilterOptions, 0),
		}, nil
	} else if len(optionParts) >= 2 {
		options := make([]FilterOptions, 0)
		for _, part := range optionParts[1:] {
			if strings.EqualFold(part, "i") {
				options = append(options, FilterOptionsCaseInsensitive)
			}
		}
		filterParts := strings.Split(optionParts[0], "=")
		if len(filterParts) != 2 {
			return nil, fmt.Errorf("invalid filter: %s", filter)
		}
		return &Filter{
			Property: filterParts[0],
			Value:    filterParts[1],
			Options:  options,
		}, nil
	} else {
		return nil, fmt.Errorf("invalid filter: %s", filter)
	}
}

func FilterByProperty[T interface{}](objects []T, filter *Filter) ([]T, error) {
	if filter == nil {
		return objects, nil
	}

	propertyName := filter.Property
	propertyValue := filter.Value

	// Filter the objects by the property value
	filteredObjects := make([]T, 0)
	for _, obj := range objects {
		objValue := reflect.ValueOf(obj)
		property := getProperty(objValue, propertyName)
		if propertyIsValid(property) {
			if len(filter.Options) > 0 {
				for _, option := range filter.Options {
					switch option {
					case FilterOptionsCaseInsensitive:
						propertyValue = strings.ToLower(fmt.Sprintf("(?i)%s", propertyValue))
					}
				}
			}
			exp, err := regexp.Compile(propertyValue)
			if err != nil {
				return nil, err
			}

			matched := exp.MatchString(fmt.Sprintf("%v", property.Interface()))
			if matched {
				filteredObjects = append(filteredObjects, obj)
			}
		}
	}

	return filteredObjects, nil
}

// Iterate through all of the map properties
func iterateMaps(value reflect.Value, f func(key, val reflect.Value)) {
	if value.Kind() != reflect.Map {
		return
	}

	for _, key := range value.MapKeys() {
		val := value.MapIndex(key)
		f(key, val)
	}
}

// Get the property value
func getProperty(value reflect.Value, propertyName string) reflect.Value {
	// Split the property name into parts
	parts := strings.Split(propertyName, ".")
	if len(parts) == 1 {
		// Get the type of the value
		valueType := value.Type()
		if valueType.Kind() == reflect.Map {
			if valueType.Key().Kind() == reflect.String && valueType.Elem().Kind() == reflect.String {
				return value.MapIndex(reflect.ValueOf(propertyName))
			} else {
				// Iterate through the map properties
				var result reflect.Value
				iterateMaps(value, func(key, val reflect.Value) {
					if strings.EqualFold(key.String(), propertyName) {
						result = val
					}
				})
				return result
			}
		}

		// Find the field with the specified property name
		for i := 0; i < valueType.NumField(); i++ {
			field := valueType.Field(i)
			if strings.EqualFold(field.Name, propertyName) {
				return value.Field(i)
			}
			tagValue := field.Tag.Get("json")
			tagValue = strings.Split(tagValue, ",")[0]
			if strings.EqualFold(tagValue, propertyName) {
				return value.Field(i)
			}
		}
	} else {
		// Get the type of the value
		valueType := value.Type()
		// Find the field with the specified property name
		for i := 0; i < valueType.NumField(); i++ {
			field := valueType.Field(i)
			if strings.EqualFold(field.Name, parts[0]) {
				return getProperty(value.FieldByName(parts[0]), strings.Join(parts[1:], "."))
			}
			tagValue := field.Tag.Get("json")
			tagValue = strings.Split(tagValue, ",")[0]
			if strings.EqualFold(tagValue, parts[0]) {
				return getProperty(value.Field(i), strings.Join(parts[1:], "."))
			}
		}
		return reflect.Value{}
	}

	return reflect.Value{}
}

func propertyIsValid(property reflect.Value) bool {
	// Check if the property is valid and can be converted to a string
	return property.IsValid() && property.CanInterface() && property.Type().Kind() != reflect.Struct
}
