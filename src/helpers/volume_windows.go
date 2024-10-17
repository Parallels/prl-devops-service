//go:build windows
// +build windows

package helpers

import (
	"golang.org/x/sys/windows"
)

// Windows volume check using GetVolumeInformation
func isSameVolume(src, dst string) (bool, error) {
	srcVolume, err := getVolumeName(src)
	if err != nil {
		return false, err
	}
	dstVolume, err := getVolumeName(dst)
	if err != nil {
		return false, err
	}
	return srcVolume == dstVolume, nil
}

// Get volume name for Windows path
func getVolumeName(path string) (string, error) {
	h, err := windows.CreateFile(windows.StringToUTF16Ptr(path), windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(h)

	var volumeNameBuffer [windows.MAX_PATH]uint16
	err = windows.GetVolumeInformationByHandle(h, &volumeNameBuffer[0], uint32(len(volumeNameBuffer)), nil, nil, nil, nil, 0)
	if err != nil {
		return "", err
	}

	return windows.UTF16ToString(volumeNameBuffer[:]), nil
}
