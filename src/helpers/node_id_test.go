package helpers

import (
	"os"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/constants"
	str "github.com/stretchr/testify/require"
)

// Node identity decides which replica owns a reservation and whose jobs a booting
// replica may recover, so the resolution order is worth pinning: a regression that
// gives two replicas the same id shows up as each one failing the other's running
// jobs, a long way from here.
//
// The order under test is the one the operator gets to influence:
// configured value for this machine > ORCHESTRATOR_NODE_ID > machine id > hostname.
// Settings outrank the pin because every other key works that way; the pin stays
// as the way back when a settings row has been mis-edited.
//
// Every test starts from ResetNodeIdentityForTests: the identity is process-wide
// and resolved once, which is right in production and would make these tests order
// dependent without the seam.

func resetIdentity(t *testing.T) {
	t.Helper()

	ResetNodeIdentityForTests()
	t.Cleanup(ResetNodeIdentityForTests)
}

func withNodeIDSource(t *testing.T, source func() (string, bool)) {
	t.Helper()

	nodeIDMu.Lock()
	nodeIDSource = source
	nodeIDMu.Unlock()
}

func withMachineIDSource(t *testing.T, source func() (string, bool)) {
	t.Helper()

	machineIDMu.Lock()
	machineIDSource = source
	machineIDMu.Unlock()
}

func thisHostname(t *testing.T) string {
	t.Helper()

	host, err := os.Hostname()
	str.NoError(t, err)
	str.NotEmpty(t, host) // the fallback cases below are only meaningful with one to compare against

	return strings.TrimSpace(host)
}

func TestConfiguredIdentityBeatsThePin(t *testing.T) {
	resetIdentity(t)
	t.Setenv(constants.OrchestratorNodeIDEnvVar, "qa-A")
	withMachineIDSource(t, func() (string, bool) { return "6E2BF1C4-1B2A", true })
	withNodeIDSource(t, func() (string, bool) { return "orchestrator-1", true })

	// The settings layer is authoritative for every key in this service; making
	// identity the exception would mean the UI can show a value that is not what
	// runs.
	str.Equal(t, "orchestrator-1", resolveNodeID())
}

func TestPinBeatsTheMachineID(t *testing.T) {
	resetIdentity(t)
	t.Setenv(constants.OrchestratorNodeIDEnvVar, "qa-A")
	withMachineIDSource(t, func() (string, bool) { return "6E2BF1C4-1B2A", true })

	str.Equal(t, "qa-A", resolveNodeID()) // explicit operator intent outranks inference
}

func TestMachineIDIsUsedWhenNothingIsConfigured(t *testing.T) {
	resetIdentity(t)
	t.Setenv(constants.OrchestratorNodeIDEnvVar, "")
	// ioreg hands back the UUID quoted, exactly as telemetry strips it.
	withMachineIDSource(t, func() (string, bool) { return `  "6E2BF1C4-1B2A-4C3D-9E7A-0F1D2E3A4B5C" `, true })

	str.Equal(t, "6E2BF1C4-1B2A-4C3D-9E7A-0F1D2E3A4B5C", resolveNodeID())
}

func TestHostnameTakesOverWhenThePlatformHasNoID(t *testing.T) {
	resetIdentity(t)
	t.Setenv(constants.OrchestratorNodeIDEnvVar, "")
	host := thisHostname(t)

	// A platform without an id (Windows) reports an error: no identity, not "".
	withMachineIDSource(t, func() (string, bool) { return "", false })
	str.Equal(t, host, resolveNodeID())

	// An id that is nothing once normalised counts as no identity either.
	ResetNodeIdentityForTests()
	withMachineIDSource(t, func() (string, bool) { return `""`, true })
	str.Equal(t, host, resolveNodeID())

	// And with no source installed at all (boot order gone wrong), still the
	// hostname rather than the empty string.
	ResetNodeIdentityForTests()
	str.Equal(t, host, resolveNodeID())
}

func TestAnUnusableConfiguredValueFallsThroughRatherThanWinning(t *testing.T) {
	resetIdentity(t)
	t.Setenv(constants.OrchestratorNodeIDEnvVar, "qa-A")
	withMachineIDSource(t, func() (string, bool) { return "6E2BF1C4-1B2A", true })

	// A configured value that is empty, or not an identity at all, must not
	// outrank anything: falling through is what keeps a bad row from producing an
	// owner nobody can match rows against.
	for _, junk := range []string{"", `  `, `""`, "has\nnewline", strings.Repeat("x", maxIdentityBytes+1)} {
		ResetNodeIdentityForTests()
		withNodeIDSource(t, func() (string, bool) { return junk, true })
		withMachineIDSource(t, func() (string, bool) { return "6E2BF1C4-1B2A", true })

		str.Equal(t, "qa-A", resolveNodeID(), "value %q should not have been accepted", junk)
	}
}

func TestDeriveNodeIDIgnoresTheConfiguredValue(t *testing.T) {
	resetIdentity(t)
	t.Setenv(constants.OrchestratorNodeIDEnvVar, "qa-A")
	withNodeIDSource(t, func() (string, bool) { return "orchestrator-1", true })
	withMachineIDSource(t, func() (string, bool) { return "6E2BF1C4-1B2A", true })

	// Deleting the override has to land somewhere, and it cannot be the row that
	// was just deleted.
	str.Equal(t, "qa-A", DeriveNodeID())
}

func TestMachineIDFallsBackToTheHostname(t *testing.T) {
	resetIdentity(t)
	host := thisHostname(t)

	withMachineIDSource(t, func() (string, bool) { return " 6E2BF1C4-1B2A ", true })
	str.Equal(t, "6E2BF1C4-1B2A", MachineID())

	ResetNodeIdentityForTests()
	withMachineIDSource(t, func() (string, bool) { return "", false })
	str.Equal(t, host, MachineID())
}

func TestSetMachineIDSourceRefusesOnceObserved(t *testing.T) {
	resetIdentity(t)
	str.True(t, SetMachineIDSource(func() (string, bool) { return "6E2BF1C4-1B2A", true }))
	str.Equal(t, "6E2BF1C4-1B2A", MachineID())

	// Late installation must say so instead of being silently ignored.
	str.False(t, SetMachineIDSource(func() (string, bool) { return "too-late", true }))
	str.Equal(t, "6E2BF1C4-1B2A", MachineID())
}

func TestSetNodeIDSourceRefusesOnceIdentityIsResolved(t *testing.T) {
	resetIdentity(t)
	NodeID() // the one identity of this process, from whatever the environment gives

	str.False(t, SetNodeIDSource(func() (string, bool) { return "too-late", true }))
}

func TestNormalizeIdentityRefusesWhatCannotBeAnIdentity(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "abc-123", "abc-123"},
		{"quoted and padded", ` "abc-123" `, "abc-123"},
		{"lone quote", `"`, ""},
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
		{"embedded newline", "abc\n123", ""}, // would break the log lines it lands in
		{"tab", "abc\t123", ""},              // and the SQL literal quoted in a warning
		{"delete", "abc" + string(rune(0x7f)), ""},
		{"past the ceiling", strings.Repeat("x", maxIdentityBytes+1), ""},
		{"at the ceiling", strings.Repeat("x", maxIdentityBytes), strings.Repeat("x", maxIdentityBytes)},
	}

	for _, tc := range cases {
		str.Equal(t, tc.want, normalizeIdentity(tc.in), tc.name)
	}
}

func TestApplyNodeIDAdoptsUsableValuesAndRefusesTheRest(t *testing.T) {
	resetIdentity(t)
	t.Setenv(constants.OrchestratorNodeIDEnvVar, "")
	withMachineIDSource(t, func() (string, bool) { return "6E2BF1C4-1B2A", true })
	str.Equal(t, "6E2BF1C4-1B2A", NodeID())

	str.True(t, ApplyNodeID(` "orchestrator-2" `))
	str.Equal(t, "orchestrator-2", NodeID()) // adopted, and stays adopted

	// A refused value must leave the working identity in place: dropping to the
	// empty owner would re-merge every replica's rows into one bucket.
	for _, junk := range []string{"", "   ", `""`, "two\nlines"} {
		str.False(t, ApplyNodeID(junk), "value %q", junk)
		str.Equal(t, "orchestrator-2", NodeID())
	}
}

func TestNodeIDAliasesCarryEveryIdentityThisMachineCouldHaveUsed(t *testing.T) {
	resetIdentity(t)
	withMachineIDSource(t, func() (string, bool) { return "6E2BF1C4-1B2A", true })
	host := thisHostname(t)

	aliases := NodeIDAliases()

	// Rows written before a rename, or before identity moved from the hostname to
	// the machine id, are still this machine's work — a row nobody owns is a row
	// nobody will ever finish.
	str.Contains(t, aliases, "6E2BF1C4-1B2A")
	str.Contains(t, aliases, host)

	// A fourth identity from an earlier rename is not carried: the aliases are
	// what this machine can derive about itself today, not a history table.
	previous := SetNodeIDForTests("renamed-away")
	t.Cleanup(func() { SetNodeIDForTests(previous) })
	str.Contains(t, NodeIDAliases(), "renamed-away")
	str.Contains(t, NodeIDAliases(), "6E2BF1C4-1B2A")
}

func TestNodeIDAliasesDeduplicate(t *testing.T) {
	resetIdentity(t)
	host := thisHostname(t)

	// The common single-replica case: identity derived from the hostname, so all
	// three candidates are the same string and recovery should get one condition.
	SetNodeIDForTests(host)
	withMachineIDSource(t, func() (string, bool) { return "", false })

	str.Equal(t, []string{host}, NodeIDAliases())
}
