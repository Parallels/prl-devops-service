package helpers

import (
	"strings"
	"testing"
	"time"
)

func TestGetUtcCurrentDateTime(t *testing.T) {
	expectedFormat := "2006-01-02T15:04:05.999999999Z07:00"
	expectedTime := time.Now().UTC().Format(expectedFormat)

	actualTime := GetUtcCurrentDateTime()

	if strings.Split(actualTime, ".")[0] != strings.Split(expectedTime, ".")[0] {
		t.Errorf("Expected %s, but got %s", expectedTime, actualTime)
	}
}
