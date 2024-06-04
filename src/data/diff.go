package data

import "encoding/json"

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
