package startup

// Boot-time identity wiring. Two things are pinned here that cannot be tested
// from the packages they serve: that the identity is settled from the right
// sources in the right order, and that renaming it through the settings service
// takes the rows with it.

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/helpers"
	str "github.com/stretchr/testify/require"
)

type recordingRenamer struct {
	calls [][2]string
	err   error
}

func (r *recordingRenamer) RenameJobOwner(_ basecontext.ApiContext, from, to string) error {
	r.calls = append(r.calls, [2]string{from, to})

	return r.err
}

func resetIdentity(t *testing.T) {
	t.Helper()

	helpers.ResetNodeIdentityForTests()
	t.Cleanup(helpers.ResetNodeIdentityForTests)
}

func TestInstallNodeIdentityDerivesFromTheMachineID(t *testing.T) {
	resetIdentity(t)
	t.Setenv(constants.OrchestratorNodeIDEnvVar, "")

	installNodeIdentity(basecontext.NewRootBaseContext(), func() (string, error) {
		return `  "6E2BF1C4-1B2A" `, nil // ioreg quotes the UUID; telemetry strips the same quotes
	})

	str.Equal(t, "6E2BF1C4-1B2A", helpers.NodeID())
	str.Equal(t, "6E2BF1C4-1B2A", helpers.MachineID())
}

func TestInstallNodeIdentityFallsBackToTheHostname(t *testing.T) {
	resetIdentity(t)
	t.Setenv(constants.OrchestratorNodeIDEnvVar, "")

	host, err := os.Hostname()
	str.NoError(t, err)

	// Windows has no platform unique id implemented: the hostname takes over and
	// the operator is told so in the log rather than left with an unexplained id.
	installNodeIdentity(basecontext.NewRootBaseContext(), func() (string, error) {
		return "", errors.New("Not implemented")
	})

	str.Equal(t, strings.TrimSpace(host), helpers.NodeID())
}

func TestInstallNodeIdentityKeepsThePinAheadOfInference(t *testing.T) {
	resetIdentity(t)
	t.Setenv(constants.OrchestratorNodeIDEnvVar, "qa-A")

	installNodeIdentity(basecontext.NewRootBaseContext(), func() (string, error) {
		return "6E2BF1C4-1B2A", nil
	})

	str.Equal(t, "qa-A", helpers.NodeID())

	// The machine id does NOT follow the pin. It is the key an operator uses to
	// address this replica in the settings service; letting a name change move it
	// would leave the override unreachable under the name it replaced.
	str.Equal(t, "6E2BF1C4-1B2A", helpers.MachineID())
	str.Contains(t, helpers.NodeIDAliases(), "6E2BF1C4-1B2A")
}

func TestIdentityRenameHookAdoptsTheNameAndMovesItsRows(t *testing.T) {
	resetIdentity(t)
	previous := helpers.SetNodeIDForTests("node-a")
	t.Cleanup(func() { helpers.SetNodeIDForTests(previous) })

	renamer := &recordingRenamer{}
	hook := identityRenameHook(basecontext.NewRootBaseContext(), renamer)

	str.NoError(t, hook("node-a", "node-b"))
	str.Equal(t, "node-b", helpers.NodeID())
	str.Equal(t, [][2]string{{"node-a", "node-b"}}, renamer.calls)
}

func TestIdentityRenameHookSkipsANameThatHasNotChanged(t *testing.T) {
	resetIdentity(t)
	previous := helpers.SetNodeIDForTests("node-a")
	t.Cleanup(func() { helpers.SetNodeIDForTests(previous) })

	renamer := &recordingRenamer{}
	hook := identityRenameHook(basecontext.NewRootBaseContext(), renamer)

	str.NoError(t, hook("node-a", "node-a"))
	str.NoError(t, hook("", "node-a")) // re-saving the name this replica already answers to
	str.Empty(t, renamer.calls)        // an UPDATE keyed on the old name would match nothing anyway
}

func TestIdentityRenameHookOnDeleteReturnsToTheDerivedName(t *testing.T) {
	resetIdentity(t)
	t.Setenv(constants.OrchestratorNodeIDEnvVar, "qa-fallback")
	previous := helpers.SetNodeIDForTests("node-a")
	t.Cleanup(func() { helpers.SetNodeIDForTests(previous) })

	renamer := &recordingRenamer{}
	hook := identityRenameHook(basecontext.NewRootBaseContext(), renamer)

	// Deleting the override has to land on the same chain boot would use, not on
	// whatever name happened to be in force a moment before.
	str.NoError(t, hook("node-a", ""))
	str.Equal(t, "qa-fallback", helpers.NodeID())
	str.Equal(t, [][2]string{{"node-a", "qa-fallback"}}, renamer.calls)
}

func TestIdentityRenameHookReportsRowsThatDidNotFollow(t *testing.T) {
	resetIdentity(t)
	previous := helpers.SetNodeIDForTests("node-a")
	t.Cleanup(func() { helpers.SetNodeIDForTests(previous) })

	renamer := &recordingRenamer{err: errors.New("db is unreachable")}
	hook := identityRenameHook(basecontext.NewRootBaseContext(), renamer)

	err := hook("node-a", "node-b")

	// Partial, and told as such: the name is adopted and stored, so the failure is
	// the bookkeeping behind it — and the retry is a re-save, not a rollback.
	str.Error(t, err)
	str.Contains(t, err.Error(), "jobs")
	str.Contains(t, err.Error(), "re-save")
	str.Equal(t, "node-b", helpers.NodeID())
	str.Equal(t, [][2]string{{"node-a", "node-b"}}, renamer.calls)
}

func TestIdentityRenameHookWithoutTheJobsSideStillAdopts(t *testing.T) {
	resetIdentity(t)
	previous := helpers.SetNodeIDForTests("node-a")
	t.Cleanup(func() { helpers.SetNodeIDForTests(previous) })

	// A backend that cannot scope a rename (no local assertion available) must
	// not block the identity change; it reports what it could not do.
	hook := identityRenameHook(basecontext.NewRootBaseContext(), nil)
	str.NoError(t, hook("node-a", "node-b"))
	str.Equal(t, "node-b", helpers.NodeID())
}

func TestIdentityRenameHookRefusesAnUnusableName(t *testing.T) {
	resetIdentity(t)
	previous := helpers.SetNodeIDForTests("node-a")
	t.Cleanup(func() { helpers.SetNodeIDForTests(previous) })

	renamer := &recordingRenamer{}
	hook := identityRenameHook(basecontext.NewRootBaseContext(), renamer)

	// The settings layer validates values, so getting here means something else
	// wrote the row. Keeping the working name beats an owner nobody can match.
	str.Error(t, hook("node-a", "not\nusable"))
	str.Equal(t, "node-a", helpers.NodeID())
	str.Empty(t, renamer.calls)
}
