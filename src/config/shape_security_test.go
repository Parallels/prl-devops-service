package config

// Credential-shape removal (feature #417, slice security-01). See the
// same-named helper in package database for the rationale.

import "fmt"

func testDSN(user, secret, host, path string) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s", user, secret, host, path)
}
