package bruteforceguard

import (
	"time"
)

type BruteForceGuardOptions struct {
	maxFailedLoginAttempts int
	blockDuration          string
	incrementalWait        bool
}

func NewDefaultOptions() *BruteForceGuardOptions {
	return &BruteForceGuardOptions{
		maxFailedLoginAttempts: 5,
		blockDuration:          "5s",
		incrementalWait:        true,
	}
}

func (bfg *BruteForceGuardOptions) WithMaxLoginAttempts(attempts int) *BruteForceGuardOptions {
	if attempts < 1 {
		attempts = 1
	}
	bfg.maxFailedLoginAttempts = attempts
	return bfg
}

func (bfg *BruteForceGuardOptions) WithBlockDuration(duration string) *BruteForceGuardOptions {
	bfg.blockDuration = duration
	return bfg
}

func (bfg *BruteForceGuardOptions) WithIncrementalWait(incremental bool) *BruteForceGuardOptions {
	bfg.incrementalWait = incremental
	return bfg
}

func (bfg *BruteForceGuardOptions) BlockDuration() time.Duration {
	duration, err := time.ParseDuration(bfg.blockDuration)
	if err != nil {
		return time.Second * 5
	}

	return duration
}

func (bfg *BruteForceGuardOptions) MaxLoginAttempts() int {
	return bfg.maxFailedLoginAttempts
}

func (bfg *BruteForceGuardOptions) IncrementalWait() bool {
	return bfg.incrementalWait
}
