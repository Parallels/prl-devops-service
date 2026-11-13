package config

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/require"
)

// The VM count limit is a hard cap, so "unlimited" is the only fallback: an unset,
// invalid, zero or negative value must never be read as a limit of one or of the
// garbage value.
func TestMaxVmsPerHost(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  int64
	}{
		{"", 0}, {"0", 0}, {"1", 1}, {"10", 10}, {"-1", 0}, {"invalid", 0}, {"10.5", 0},
		{"9223372036854775808", 0}, {" 4 ", 4},
	} {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv(constants.MAX_VMS_PER_HOST_ENV_VAR, tc.value)
			require.Equal(t, tc.want, New(basecontext.NewBaseContext()).MaxVmsPerHost())
		})
	}
}

func TestMaxVmsPerHostMode(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  string
	}{
		{"", "any"}, {"any", "any"}, {"ANY", "any"}, {" any ", "any"},
		{"running", "running"}, {"Running", "running"}, {"RUNNING", "running"},
		{"stopped", "any"}, {"all", "any"},
	} {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv(constants.MAX_VMS_PER_HOST_MODE_ENV_VAR, tc.value)
			require.Equal(t, tc.want, New(basecontext.NewBaseContext()).MaxVmsPerHostMode())
		})
	}
}
