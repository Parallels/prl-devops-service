//go:build windows
// +build windows

package helpers

import (
	"golang.org/x/sys/windows"

	"github.com/Parallels/prl-devops-service/errors"
)

func GetFreeDiskSpace(folder string) (int64, error) {
	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64
	folderPtr, err := windows.UTF16PtrFromString(folder)
	if err != nil {
		return 0, errors.NewFromErrorWithCode(err, 500)
	}
	err = windows.GetDiskFreeSpaceEx(folderPtr, &freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes)
	if err != nil {
		return 0, errors.NewFromErrorWithCode(err, 500)
	}

	diskFreeSpaceInMb := int64(freeBytesAvailable) / (1024 * 1024)
	return diskFreeSpaceInMb, nil
}

func GetTotalDiskSpace(folder string) (int64, error) {
	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64
	folderPtr, err := windows.UTF16PtrFromString(folder)
	if err != nil {
		return 0, errors.NewFromErrorWithCode(err, 500)
	}
	err = windows.GetDiskFreeSpaceEx(folderPtr, &freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes)
	if err != nil {
		return 0, errors.NewFromErrorWithCode(err, 500)
	}

	diskTotalSpaceInMb := int64(totalNumberOfBytes) / (1024 * 1024)
	return diskTotalSpaceInMb, nil
}
