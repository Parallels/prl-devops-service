package models_test

// Credential-shape removal (feature #417, slice security-01). See the
// same-named helper in package database for the rationale. Unlike the other
// packages, this fixture lands in committed golden JSON files (user.json),
// so the value is a self-describing zero-entropy placeholder rather than a
// digest — committed test data must never resemble a real secret either.

import "strings"

func fixtureSecret(parts ...string) string {
	return "placeholder-" + strings.Join(parts, "-")
}
