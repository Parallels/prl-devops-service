//go:build windows
// +build windows

package stats

import (
	"syscall"
	"unsafe"
)

// getCPUTimes returns user and system CPU time in seconds using GetProcessTimes.
// Windows FILETIME values are in 100-nanosecond intervals.
func getCPUTimes() (userSeconds float64, systemSeconds float64, err error) {
	handle := syscall.Handle(^uintptr(0)) // pseudo-handle for current process

	var creationTime, exitTime, kernelTime, userTime syscall.Filetime
	err = syscall.GetProcessTimes(
		handle,
		&creationTime,
		&exitTime,
		&kernelTime,
		&userTime,
	)
	if err != nil {
		return 0, 0, err
	}

	// Convert FILETIME (100ns intervals) to seconds
	userSeconds = float64(*(*uint64)(unsafe.Pointer(&userTime))) / 1e7
	systemSeconds = float64(*(*uint64)(unsafe.Pointer(&kernelTime))) / 1e7
	return userSeconds, systemSeconds, nil
}
