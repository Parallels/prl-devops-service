package config

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/require"
)

func TestDbGhostJobTimeoutMinutes(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  int
	}{
		{"", 5}, {"30", 30}, {"1", 1}, {"0", 5}, {"-1", 5},
		{"invalid", 5}, {"153722868", 5},
	} {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv(constants.GhostJobTimeoutMinutesEnvVar, tc.value)
			cfg := New(basecontext.NewBaseContext())
			require.Equal(t, tc.want, cfg.DbGhostJobTimeoutMinutes())
		})
	}
}
