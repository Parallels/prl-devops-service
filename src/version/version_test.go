package version

import (
	"testing"
)

// ── Parse ────────────────────────────────────────────────────────────────────

func TestParse(t *testing.T) {
	tests := []struct {
		input   string
		major   int
		minor   int
		patch   int
		channel Channel
		wantErr bool
	}{
		{"1.2.3", 1, 2, 3, ChannelStable, false},
		{"v1.2.3", 1, 2, 3, ChannelStable, false},
		{"release-v1.2.3", 1, 2, 3, ChannelStable, false},
		{"release-1.2.3", 1, 2, 3, ChannelStable, false},
		{"1.2.3-beta", 1, 2, 3, ChannelBeta, false},
		{"1.2.3-canary", 1, 2, 3, ChannelCanary, false},
		{"1.2.3-alpha", 1, 2, 3, ChannelCanary, false},
		{"1.2.3-dev", 1, 2, 3, ChannelCanary, false},
		{"v0.9.21", 0, 9, 21, ChannelStable, false},
		{"1.0", 1, 0, 0, ChannelStable, false},
		{"1", 1, 0, 0, ChannelStable, false},
		{"not-a-version", 0, 0, 0, ChannelCanary, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.Major != tt.major || got.Minor != tt.minor || got.Patch != tt.patch {
				t.Errorf("Parse(%q) = %d.%d.%d, want %d.%d.%d",
					tt.input, got.Major, got.Minor, got.Patch,
					tt.major, tt.minor, tt.patch)
			}
			if got.Channel != tt.channel {
				t.Errorf("Parse(%q) channel = %q, want %q", tt.input, got.Channel, tt.channel)
			}
		})
	}
}

// ── Channel predicates ───────────────────────────────────────────────────────

func TestChannelPredicates(t *testing.T) {
	stable := &SemVer{Channel: ChannelStable}
	beta := &SemVer{Channel: ChannelBeta}
	canary := &SemVer{Channel: ChannelCanary}

	if !stable.IsStable() || stable.IsBeta() || stable.IsCanary() {
		t.Error("stable: wrong predicates")
	}
	if stable.IsStable() == beta.IsStable() {
		t.Error("beta: IsStable should be false")
	}
	if !beta.IsBeta() || beta.IsCanary() {
		t.Error("beta: wrong predicates")
	}
	if !canary.IsCanary() || canary.IsBeta() || canary.IsStable() {
		t.Error("canary: wrong predicates")
	}
}

// ── String ───────────────────────────────────────────────────────────────────

func TestString(t *testing.T) {
	tests := []struct {
		v    *SemVer
		want string
	}{
		{&SemVer{Major: 1, Minor: 2, Patch: 3, Channel: ChannelStable}, "1.2.3"},
		{&SemVer{Major: 1, Minor: 2, Patch: 3, Channel: ChannelBeta}, "1.2.3-beta"},
		{&SemVer{Major: 1, Minor: 2, Patch: 3, Channel: ChannelCanary}, "1.2.3-canary"},
		{&SemVer{Major: 0, Minor: 9, Patch: 21, Channel: ChannelStable}, "0.9.21"},
	}
	for _, tt := range tests {
		if got := tt.v.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}

// ── Compare ──────────────────────────────────────────────────────────────────

func TestCompare(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		// Numeric wins
		{"higher major wins", "2.0.0", "1.9.9", 1},
		{"lower major loses", "1.0.0", "2.0.0", -1},
		{"higher minor wins", "1.1.0", "1.0.9", 1},
		{"higher patch wins", "1.0.1", "1.0.0", 1},
		{"canary beats older stable (numeric)", "1.1.0-canary", "1.0.0-stable", 1},

		// Channel tiebreak (same numeric)
		{"stable beats beta", "1.0.0", "1.0.0-beta", 1},
		{"stable beats canary", "1.0.0", "1.0.0-canary", 1},
		{"beta beats canary", "1.0.0-beta", "1.0.0-canary", 1},
		{"canary loses to stable", "1.0.0-canary", "1.0.0", -1},
		{"canary loses to beta", "1.0.0-canary", "1.0.0-beta", -1},
		{"beta loses to stable", "1.0.0-beta", "1.0.0", -1},

		// Equality
		{"equal stable", "1.2.3", "1.2.3", 0},
		{"equal beta", "1.2.3-beta", "1.2.3-beta", 0},
		{"equal canary", "1.2.3-canary", "1.2.3-canary", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareStrings(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("CompareStrings(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
