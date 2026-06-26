package common

import (
	"database/sql/driver"
	"encoding/json"
)

// JSONObject is a generic type for JSON marshaling/unmarshaling of single objects
type JSONObject[T any] struct {
	Data T
}

// Value implements driver.Valuer interface for JSON marshaling
func (j JSONObject[T]) Value() (driver.Value, error) {
	return json.Marshal(j.Data)
}

// Scan implements sql.Scanner interface for JSON unmarshaling
func (j *JSONObject[T]) Scan(value interface{}) error {
	if value == nil {
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

	return json.Unmarshal(bytes, &j.Data)
}

// Get returns the underlying data
func (j JSONObject[T]) Get() T {
	return j.Data
}

// Set sets the underlying data
func (j *JSONObject[T]) Set(data T) {
	j.Data = data
}
