//go:build windows
// +build windows

package parallelsdesktop

import "os"

// getFileOwner returns 0, 0 on Windows — Unix ownership is not applicable.
func getFileOwner(info os.FileInfo) (uid int, gid int) {
	return 0, 0
}
