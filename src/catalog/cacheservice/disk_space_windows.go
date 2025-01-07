package cacheservice

import (
	"errors"
	"syscall"
	"unsafe"
)

func (cs *CacheService) getFreeDiskSpace() (int64, error) {
	// Convert Go string to UTF16 pointer
	lpDirectoryName, err := syscall.UTF16PtrFromString(cs.cacheFolder)
	if err != nil {
		return 0, err
	}

	// Load the kernel32 DLL and find the procedure
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	defer kernel32.Release()

	procGetDiskFreeSpaceExW := kernel32.MustFindProc("GetDiskFreeSpaceExW")

	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64

	// Call the Windows API
	r, _, callErr := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(lpDirectoryName)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)
	if r == 0 {
		// The call failed
		if callErr != syscall.Errno(0) {
			return 0, callErr
		}
		return 0, errors.New("GetDiskFreeSpaceExW failed, unknown error")
	}

	// freeBytesAvailable is the number of bytes available to the *caller* (which
	// can differ from the total free bytes if quotas are in use, but in most
	// cases is the same).
	// Convert bytes -> MB
	diskFreeSpaceInMB := freeBytesAvailable / (1024 * 1024)

	return int64(diskFreeSpaceInMB), nil
}
