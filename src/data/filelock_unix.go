//go:build !windows

package data

import (
	"os"
	"syscall"
)

// lockFileDescriptor acquires an exclusive advisory lock on the given file,
// blocking until it can be obtained. This guards the database save against
// multiple service instances pointing at the same database directory.
func lockFileDescriptor(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

// unlockFileDescriptor releases the advisory lock acquired by lockFileDescriptor.
func unlockFileDescriptor(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
