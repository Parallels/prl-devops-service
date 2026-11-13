package data

// Credential-shape removal (feature #417, slice security-01). See the
// same-named helper in package database for the rationale: derived fixture
// values, deterministic per seed, never asserted verbatim.

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

func fixtureSecret(parts ...string) string {
	sum := sha256.Sum256([]byte("security-01|" + strings.Join(parts, "|")))
	return "Fx1!" + base64.RawURLEncoding.EncodeToString(sum[:9])
}
