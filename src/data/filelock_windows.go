//go:build windows

package data

import "os"

// lockFileDescriptor is a no-op on Windows. The multi-instance scenario the lock
// guards against is the Linux/macOS service deployment; on Windows we rely on the
// in-process mutex only. A LockFileEx-based implementation can be added here if
// concurrent Windows instances ever become a supported configuration.
func lockFileDescriptor(_ *os.File) error { return nil }

// unlockFileDescriptor is a no-op on Windows. See lockFileDescriptor.
func unlockFileDescriptor(_ *os.File) error { return nil }
