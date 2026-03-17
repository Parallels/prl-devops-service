package version

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// Channel represents the release channel a binary was built for.
type Channel string

const (
	ChannelStable Channel = "stable"
	ChannelBeta   Channel = "beta"
	ChannelCanary Channel = "canary"
)

// Build-time variables — injected via -ldflags at compile time.
var (
	buildVersion = "0.0.0"
	buildChannel = "stable"
	buildDate    = ""
	buildCommit  = ""
)

// SemVer holds a parsed semantic version with a release channel.
type SemVer struct {
	Major   int
	Minor   int
	Patch   int
	Channel Channel
	// Raw is the original unparsed string passed to Parse, or the ldflag value.
	Raw string
}

var (
	current     *SemVer
	currentOnce sync.Once
)

// channelWeight returns the ordering priority of a channel.
// Higher value means newer / more stable.
func channelWeight(c Channel) int {
	switch c {
	case ChannelStable:
		return 2
	case ChannelBeta:
		return 1
	default: // canary, alpha, dev, or anything unrecognised
		return 0
	}
}

// parseChannel maps a raw suffix string to a Channel.
// Anything unrecognised is treated as ChannelCanary (oldest/least stable).
func parseChannel(s string) Channel {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "stable", "":
		return ChannelStable
	case "beta":
		return ChannelBeta
	case "canary", "alpha", "dev", "development":
		return ChannelCanary
	default:
		return ChannelCanary
	}
}

// Parse parses a version string into a SemVer.
//
// Accepted formats:
//
//	"1.2.3"          → stable
//	"v1.2.3"         → stable  (leading v stripped)
//	"release-v1.2.3" → stable  (release- prefix stripped)
//	"1.2.3-beta"     → beta
//	"1.2.3-canary"   → canary
//
// Missing patch or minor segments default to 0.
func Parse(s string) (*SemVer, error) {
	raw := s
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "release-")
	s = strings.TrimPrefix(s, "v")

	ch := ChannelStable
	if idx := strings.Index(s, "-"); idx != -1 {
		ch = parseChannel(s[idx+1:])
		s = s[:idx]
	}

	parts := strings.Split(s, ".")
	nums := make([]int, 3)
	for i := 0; i < len(parts) && i < 3; i++ {
		n, err := strconv.Atoi(strings.TrimSpace(parts[i]))
		if err != nil {
			return nil, fmt.Errorf("version: invalid segment %q in %q", parts[i], raw)
		}
		nums[i] = n
	}

	return &SemVer{
		Major:   nums[0],
		Minor:   nums[1],
		Patch:   nums[2],
		Channel: ch,
		Raw:     raw,
	}, nil
}

// Get returns the singleton SemVer built from build-time ldflag variables.
// Safe for concurrent use.
func Get() *SemVer {
	currentOnce.Do(func() {
		v, err := Parse(buildVersion)
		if err != nil {
			v = &SemVer{Raw: buildVersion}
		}
		// Channel is set independently of the version string so that the
		// Makefile can inject them separately (e.g. VERSION=1.2.3, CHANNEL=beta).
		v.Channel = parseChannel(buildChannel)
		v.Raw = buildVersion
		current = v
	})
	return current
}

// String returns the canonical version string.
// Stable builds omit the channel suffix: "1.2.3".
// Pre-release builds include it: "1.2.3-beta", "1.2.3-canary".
func (v *SemVer) String() string {
	base := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Channel != ChannelStable {
		return base + "-" + string(v.Channel)
	}
	return base
}

// IsStable reports whether the binary was built for the stable channel.
func (v *SemVer) IsStable() bool { return v.Channel == ChannelStable }

// IsBeta reports whether the binary was built for the beta channel.
func (v *SemVer) IsBeta() bool { return v.Channel == ChannelBeta }

// IsCanary reports whether the binary was built for the canary channel.
func (v *SemVer) IsCanary() bool { return v.Channel == ChannelCanary }

// BuildDate returns the RFC3339 timestamp this binary was compiled, or an
// empty string when the ldflag was not provided.
func BuildDate() string { return buildDate }

// BuildCommit returns the short git commit hash this binary was built from,
// or an empty string when the ldflag was not provided.
func BuildCommit() string { return buildCommit }
