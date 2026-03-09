//go:build linux || darwin
// +build linux darwin

package stats

import "syscall"

// getCPUTimes returns user and system CPU time in seconds using getrusage.
func getCPUTimes() (userSeconds float64, systemSeconds float64, err error) {
	var rUsage syscall.Rusage
	if err = syscall.Getrusage(syscall.RUSAGE_SELF, &rUsage); err != nil {
		return 0, 0, err
	}
	userSeconds = float64(rUsage.Utime.Sec) + float64(rUsage.Utime.Usec)/1e6
	systemSeconds = float64(rUsage.Stime.Sec) + float64(rUsage.Stime.Usec)/1e6
	return userSeconds, systemSeconds, nil
}
