package helpers

import (
	"io"
	"sync"
)

type ProgressWriter struct {
	writer          io.Writer
	totalDownloaded int64
	progress        chan int
	size            int64
	mu              sync.Mutex
}

func NewProgressWriter(writer io.Writer, size int64, progress chan int) io.Writer {
	return &ProgressWriter{
		writer:   writer,
		progress: progress,
		size:     size,
	}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	n, err := pw.writer.Write(p)
	pw.totalDownloaded += int64(len(p))
	if err == nil {
		if pw.progress != nil && pw.size > 0 {
			pw.progress <- int(float32(pw.totalDownloaded*100) / float32(pw.size))
		}
	}
	return n, err
}

type ProgressReporter struct {
	Progress chan int
	Size     int64
}

func NewProgressReporter(size int64, progressChannel chan int) *ProgressReporter {
	return &ProgressReporter{
		Progress: progressChannel,
		Size:     size,
	}
}
