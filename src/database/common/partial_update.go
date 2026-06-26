// Package common provides utility functions for the database layer of the application
package common

import (
	"reflect"
	"strings"
	"time"
)

// PartialUpdateMap generates a map of field updates by comparing the original entity
// with the updated entity. Only fields that have actually changed are included.
// This mimics MongoDB's behavior where only changed fields are updated.
func PartialUpdateMap(original, updated interface{}, alwaysUpdateFields ...string) map[string]interface{} {
	updates := make(map[string]interface{})

	// Get the original and updated values
	originalVal := reflect.ValueOf(original)
	updatedVal := reflect.ValueOf(updated)

	if originalVal.Kind() == reflect.Ptr {
		originalVal = originalVal.Elem()
	}
	if updatedVal.Kind() == reflect.Ptr {
		updatedVal = updatedVal.Elem()
	}

	// Ensure both values are of the same type
	if originalVal.Type() != updatedVal.Type() {
		// If types don't match, fall back to zero-value approach
		return partialUpdateMapZeroValue(updated, alwaysUpdateFields...)
	}

	// Compare fields and only update changed ones
	for i := 0; i < updatedVal.NumField(); i++ {
		field := updatedVal.Field(i)
		originalField := originalVal.Field(i)
		fieldType := updatedVal.Type().Field(i)

		// Handle embedded structs (like BaseModel)
		if fieldType.Anonymous && field.Kind() == reflect.Struct {
			// Recursively process embedded struct fields
			embeddedUpdates := processEmbeddedStruct(originalField, field, fieldType)
			for key, value := range embeddedUpdates {
				updates[key] = value
			}
			continue
		}

		// Skip fields that should not be updated
		if shouldSkipField(field, fieldType) {
			continue
		}

		// Get the JSON tag
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" {
			// If no json tag, use the field name as fallback
			jsonTag = fieldType.Name
		}

		// Remove omitempty from the tag
		if commaIndex := strings.Index(jsonTag, ","); commaIndex != -1 {
			jsonTag = jsonTag[:commaIndex]
		}

		// Check if the field has changed
		if hasChanged(originalField, field) {
			// For most types, only update if the new value is not zero
			// But for booleans and pointers, allow zero values as they are meaningful
			shouldUpdate := !isZeroValue(field) ||
				field.Kind() == reflect.Bool ||
				field.Kind() == reflect.Ptr

			if shouldUpdate {
				updates[jsonTag] = field.Interface()
			}
		}
	}

	// Always update specified fields
	for _, fieldName := range alwaysUpdateFields {
		switch fieldName {
		case "updated_at":
			updates["updated_at"] = time.Now()
		}
	}

	return updates
}

// processEmbeddedStruct handles embedded structs like BaseModel
func processEmbeddedStruct(originalField, updatedField reflect.Value, fieldType reflect.StructField) map[string]interface{} {
	updates := make(map[string]interface{})

	// Process each field in the embedded struct
	for i := 0; i < fieldType.Type.NumField(); i++ {
		embeddedFieldType := fieldType.Type.Field(i)

		// Get the field values from both original and updated
		originalEmbeddedField := originalField.Field(i)
		updatedEmbeddedField := updatedField.Field(i)

		// Skip fields that should not be updated
		if shouldSkipField(updatedEmbeddedField, embeddedFieldType) {
			continue
		}

		// Get the JSON tag
		jsonTag := embeddedFieldType.Tag.Get("json")
		if jsonTag == "" {
			// If no json tag, use the field name as fallback
			jsonTag = embeddedFieldType.Name
		}

		// Remove omitempty from the tag
		if commaIndex := strings.Index(jsonTag, ","); commaIndex != -1 {
			jsonTag = jsonTag[:commaIndex]
		}

		// Check if the field has changed
		if hasChanged(originalEmbeddedField, updatedEmbeddedField) {
			// For most types, only update if the new value is not zero
			// But for booleans and pointers, allow zero values as they are meaningful
			shouldUpdate := !isZeroValue(updatedEmbeddedField) ||
				updatedEmbeddedField.Kind() == reflect.Bool ||
				updatedEmbeddedField.Kind() == reflect.Ptr

			if shouldUpdate {
				updates[jsonTag] = updatedEmbeddedField.Interface()
			}
		}
	}

	return updates
}

// hasChanged compares two reflect.Value objects and returns true if they are different
func hasChanged(original, updated reflect.Value) bool {
	// Handle different types
	if original.Kind() != updated.Kind() {
		return true
	}

	switch original.Kind() {
	case reflect.String:
		return original.String() != updated.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return original.Int() != updated.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return original.Uint() != updated.Uint()
	case reflect.Float32, reflect.Float64:
		return original.Float() != updated.Float()
	case reflect.Bool:
		return original.Bool() != updated.Bool()
	case reflect.Ptr:
		if original.IsNil() && updated.IsNil() {
			return false
		}
		if original.IsNil() || updated.IsNil() {
			return true
		}
		return hasChanged(original.Elem(), updated.Elem())
	case reflect.Struct:
		// For structs, compare each field
		for i := 0; i < original.NumField(); i++ {
			if hasChanged(original.Field(i), updated.Field(i)) {
				return true
			}
		}
		return false
	case reflect.Slice, reflect.Array:
		if original.Len() != updated.Len() {
			return true
		}
		for i := 0; i < original.Len(); i++ {
			if hasChanged(original.Index(i), updated.Index(i)) {
				return true
			}
		}
		return false
	case reflect.Map:
		if original.Len() != updated.Len() {
			return true
		}
		// For maps, we'll consider them different if they have different keys or values
		// This is a simplified comparison - in practice, you might want more sophisticated map comparison
		return !reflect.DeepEqual(original.Interface(), updated.Interface())
	case reflect.Interface:
		if original.IsNil() && updated.IsNil() {
			return false
		}
		if original.IsNil() || updated.IsNil() {
			return true
		}
		return !reflect.DeepEqual(original.Interface(), updated.Interface())
	default:
		// For other types, use DeepEqual as fallback
		return !reflect.DeepEqual(original.Interface(), updated.Interface())
	}
}

// partialUpdateMapZeroValue is a fallback function that only updates non-zero values
// This is used when the original and updated entities are of different types
func partialUpdateMapZeroValue(updated interface{}, alwaysUpdateFields ...string) map[string]interface{} {
	updates := make(map[string]interface{})

	v := reflect.ValueOf(updated)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Check if it's a struct type
	if v.Kind() != reflect.Struct {
		// If it's not a struct, return empty updates
		// Always update specified fields even for non-struct types
		for _, fieldName := range alwaysUpdateFields {
			switch fieldName {
			case "updated_at":
				updates["updated_at"] = time.Now()
			}
		}
		return updates
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip fields that should not be updated
		if shouldSkipField(field, fieldType) {
			continue
		}

		// Only include non-zero values
		if !isZeroValue(field) {
			jsonTag := fieldType.Tag.Get("json")
			if jsonTag == "" {
				// If no json tag, use the field name as fallback
				jsonTag = fieldType.Name
			}

			// Remove omitempty from the tag
			if commaIndex := strings.Index(jsonTag, ","); commaIndex != -1 {
				jsonTag = jsonTag[:commaIndex]
			}

			updates[jsonTag] = field.Interface()
		}
	}

	// Always update specified fields
	for _, fieldName := range alwaysUpdateFields {
		switch fieldName {
		case "updated_at":
			updates["updated_at"] = time.Now()
		}
	}

	return updates
}

// isZeroValue checks if a reflect.Value represents a zero value
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	default:
		return false
	}
}

// shouldSkipField determines if a field should be skipped during partial updates
func shouldSkipField(field reflect.Value, fieldType reflect.StructField) bool {
	// Skip unexported fields
	if !field.CanSet() {
		return true
	}

	// Skip fields with gorm:"-" tag
	gormTag := fieldType.Tag.Get("gorm")
	if strings.Contains(gormTag, "-") {
		return true
	}

	// Skip fields with json:"-" tag
	jsonTag := fieldType.Tag.Get("json")
	if jsonTag == "-" {
		return true
	}

	// Skip fields that are typically managed by the system
	fieldName := fieldType.Name
	skipFields := []string{"ID", "CreatedAt", "DeletedAt"} // Removed UpdatedAt and Slug from skip list
	for _, skipField := range skipFields {
		if fieldName == skipField {
			return true
		}
	}

	return false
}
