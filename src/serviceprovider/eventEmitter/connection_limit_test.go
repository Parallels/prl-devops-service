package eventemitter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMultipleConnectionsPerIPAllowed(t *testing.T) {
	// This is a build-time constant for the release build
	result := isMultipleConnectionsPerIPAllowed()
	// Currently it returns true in both builds
	assert.True(t, result, "Multiple connections per IP is currently allowed")
}
