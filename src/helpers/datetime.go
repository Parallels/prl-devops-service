package helpers

import "time"

func GetUtcCurrentDateTime() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}
