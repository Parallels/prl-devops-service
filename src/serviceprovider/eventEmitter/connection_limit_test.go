package eventemitter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMultipleConnectionsPerIPAllowed(t *testing.T) {
	// This is a build-time constant for the release build
	result := isMultipleConnectionsPerIPAllowed()
	assert.False(t, result, "Multiple connections per IP should not be allowed in release build")
}
