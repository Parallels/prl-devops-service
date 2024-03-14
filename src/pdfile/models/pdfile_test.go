package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPDFile_HasAuthentication(t *testing.T) {
	pdFile := &PDFile{}
	t.Run("Authentication is nil", func(t *testing.T) {

		assert.False(t, pdFile.HasAuthentication())
	})

	t.Run("ApiKey is not empty", func(t *testing.T) {
		pdFile = &PDFile{
			Authentication: &PDFileAuthentication{
				ApiKey: "api-key",
			},
		}
		assert.True(t, pdFile.HasAuthentication())
	})

	t.Run("Username and Password are not empty", func(t *testing.T) {
		pdFile = &PDFile{
			Authentication: &PDFileAuthentication{
				Username: "username",
				Password: "password",
			},
		}
		assert.True(t, pdFile.HasAuthentication())
	})

	t.Run("ApiKey is empty, Username and Password are empty", func(t *testing.T) {
		pdFile = &PDFile{
			Authentication: &PDFileAuthentication{},
		}
		assert.False(t, pdFile.HasAuthentication())
	})
}
