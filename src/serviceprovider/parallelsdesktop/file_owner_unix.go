//go:build linux || darwin
// +build linux darwin

package parallelsdesktop

import (
	"os"
	"syscall"
)

// getFileOwner extracts the Unix uid and gid from a FileInfo's underlying Stat_t.
func getFileOwner(info os.FileInfo) (uid int, gid int) {
	stat := info.Sys().(*syscall.Stat_t)
	return int(stat.Uid), int(stat.Gid)
}
