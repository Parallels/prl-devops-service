package parity

// Credential-shape removal (feature #417, slice security-01). See the
// same-named helper in package database for the rationale: fixture values
// are derived at runtime so no credential-shaped literal appears in source;
// they are deterministic per seed and never asserted verbatim.

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

func fixtureSecret(parts ...string) string {
	sum := sha256.Sum256([]byte("security-01|" + strings.Join(parts, "|")))
	return "Fx1!" + base64.RawURLEncoding.EncodeToString(sum[:9])
}

func testDSN(user, secret, host, path string) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s", user, secret, host, path)
}
