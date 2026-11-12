package data

import "encoding/json"

// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v1.3.0.
func Diff(a interface{}, b interface{}) bool {
	jsonA, err := json.Marshal(a)
	if err != nil {
		return false
	}
	jsonB, err := json.Marshal(b)
	if err != nil {
		return false
	}

	return string(jsonA) != string(jsonB)
}
