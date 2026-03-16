package version

// Compare compares two SemVer values and returns:
//
//	-1  if a is older than b
//	 0  if a and b are the same version
//	+1  if a is newer than b
//
// Numeric components (Major, Minor, Patch) are compared first.
// When all numeric parts are equal the channel acts as a tiebreaker:
//
//	stable (2) > beta (1) > canary (0)
//
// Examples:
//
//	Compare("1.1.0-canary", "1.0.0-stable") → +1  (higher number wins)
//	Compare("1.0.0-canary", "1.0.0-beta")   → -1  (same number; beta > canary)
//	Compare("1.0.0-beta",   "1.0.0-stable") → -1  (same number; stable > beta)
//	Compare("1.0.0-stable", "1.0.0-stable") →  0
func Compare(a, b *SemVer) int {
	for _, pair := range [3][2]int{
		{a.Major, b.Major},
		{a.Minor, b.Minor},
		{a.Patch, b.Patch},
	} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}

	// Numeric parts are equal — use channel weight as tiebreaker.
	wa, wb := channelWeight(a.Channel), channelWeight(b.Channel)
	if wa < wb {
		return -1
	}
	if wa > wb {
		return 1
	}
	return 0
}

// CompareStrings parses a and b then delegates to Compare.
// Unparseable strings are treated as version "0.0.0-canary" (oldest possible).
func CompareStrings(a, b string) int {
	va, err := Parse(a)
	if err != nil {
		va = &SemVer{Channel: ChannelCanary}
	}
	vb, err := Parse(b)
	if err != nil {
		vb = &SemVer{Channel: ChannelCanary}
	}
	return Compare(va, vb)
}
