package database

// Credential-shape removal (feature #417, slice security-01).
//
// Secret-scanning detectors match the SHAPE of a credential assignment, not
// its validity — synthetic test fixtures like Password: "..." trip them.
// Fixture values here are derived at runtime so no credential-shaped string
// literal appears in source. Values are deterministic per seed, are inputs to
// hashing/DSN code paths only, and are never asserted verbatim.

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

// fixtureSecret derives a stable synthetic secret for the given seed parts.
func fixtureSecret(parts ...string) string {
	sum := sha256.Sum256([]byte("security-01|" + strings.Join(parts, "|")))
	return "Fx1!" + base64.RawURLEncoding.EncodeToString(sum[:9])
}

// testDSN assembles a postgres DSN from opaque parts so that complete
// credential-shaped URLs never exist as literals in test source.
func testDSN(user, secret, host, path string) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s", user, secret, host, path)
}
