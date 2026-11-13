package legacy_migrator

// AC-7 input-resolution units: precedence (env beats flag, Config.GetKey
// order) + accepted spellings; and the dry-run report's shape.

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
)

// setDryEnv sets the env var; "" means "definitively unset" (t.Setenv then
// Unsetenv — restoration still belongs to the testing framework).
func setDryEnv(t *testing.T, val string) {
	t.Helper()
	if val == "" {
		t.Setenv(constants.MIGRATE_DRY_RUN_ENV_VAR, "x")
		os.Unsetenv(constants.MIGRATE_DRY_RUN_ENV_VAR)
		return
	}
	t.Setenv(constants.MIGRATE_DRY_RUN_ENV_VAR, val)
}

// setArgs swaps os.Args (common-go helper.GetFlagValue reads it directly);
// sequential package tests make the swap safe, cleanup always restores.
func setArgs(t *testing.T, args ...string) {
	t.Helper()
	saved := os.Args
	os.Args = append([]string{"prl-devops-service"}, args...)
	t.Cleanup(func() { os.Args = saved })
}

func TestIsDryRunRequested(t *testing.T) {
	cases := []struct {
		name string
		env  string // "" = unset
		args []string
		want bool
	}{
		{"both unset", "", nil, false},
		{"env true", "true", nil, true},
		{"env 1", "1", nil, true},
		{"env yes", "yes", nil, true},
		{"env TRUE case-insensitive", "TRUE", nil, true},
		{"env false", "false", nil, false},
		{"env garbage", "maybe", nil, false},
		{"flag true no env", "", []string{"--migrate-dry-run=true"}, true},
		{"flag space value NOT supported (pinned common-go semantics)", "", []string{"--migrate-dry-run", "true"}, false},
		{"flag false env true (env wins)", "true", []string{"--migrate-dry-run=false"}, true},
		{"flag true env false (env wins)", "false", []string{"--migrate-dry-run=true"}, false},
		{"value-less flag inert", "", []string{"--migrate-dry-run"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			setDryEnv(t, c.env)
			if c.args != nil {
				setArgs(t, c.args...)
			} else {
				setArgs(t) // no flags at all
			}
			assert.Equal(t, c.want, IsDryRunRequested())
		})
	}
}

func TestReportDryRun(t *testing.T) {
	ctx := newRecordingCtx()
	ReportDryRun(ctx, Report{DryRun: true, Collections: []CollectionReport{
		{Name: "ApiKeys", LegacyCount: 2, WrittenCount: 2},
	}})
	joined := strings.Join(ctx.all(), "\n")
	assert.Contains(t, joined, "[dry-run] legacy migration NOT executed; 1 collection(s) would be written:")
	assert.Contains(t, joined, fmt.Sprintf("[dry-run] would write %d record(s) into %s", 2, "ApiKeys"))
	assert.Contains(t, joined, "[dry-run] no data was written (no marker; data.json untouched); migration did NOT run")
}
