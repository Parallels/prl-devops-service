package common

import (
	"database/sql/driver"
	"encoding/json"
)

type (
	StringSlice = JSONSlice[string]
)

// JSONSlice is a generic type for JSON marshaling/unmarshaling of slices
type JSONSlice[T any] []T

// Value implements driver.Valuer interface for JSON marshaling
func (j JSONSlice[T]) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner interface for JSON unmarshaling
func (j *JSONSlice[T]) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}

	return json.Unmarshal(bytes, j)
}
