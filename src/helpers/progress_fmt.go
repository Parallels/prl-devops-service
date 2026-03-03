package helpers

import (
	"fmt"
	"time"
)

func FormatStreamingProgress(prefix string, percent float64, currentSize int64, totalSize int64, startTime time.Time) string {
	msg := fmt.Sprintf("%s %.1f%% [%s/%s]", prefix, percent, formatSizeSI(currentSize), formatSizeSI(totalSize))
	elapsed := time.Since(startTime)
	if elapsed <= 0 || currentSize <= 0 {
		return msg
	}

	bytesPerSecond := float64(currentSize) / elapsed.Seconds()
	if bytesPerSecond > 0 {
		msg += fmt.Sprintf(" %s/s", formatSizeSI(int64(bytesPerSecond)))
	}

	if totalSize > currentSize && bytesPerSecond > 0 {
		remaining := float64(totalSize-currentSize) / bytesPerSecond
		msg += fmt.Sprintf(" ETA: %s", (time.Duration(remaining) * time.Second).Round(time.Second))
	}

	return msg
}

func formatSizeSI(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
