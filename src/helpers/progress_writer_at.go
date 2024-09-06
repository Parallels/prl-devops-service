package helpers

import (
	"io"
	"sync"
)

type ProgressWriterAt struct {
	writer   io.WriterAt
	progress chan int
	size     int64
	mu       sync.Mutex
}

func NewProgressWriterAt(writer io.WriterAt, size int64, progress chan int) io.WriterAt {
	return &ProgressWriterAt{
		writer:   writer,
		progress: progress,
		size:     size,
	}
}

func (pw *ProgressWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	n, err = pw.writer.WriteAt(p, off)
	if err == nil {
		if pw.progress != nil && pw.size > 0 {
			go func() {
				pw.progress <- int(float32(off*100) / float32(pw.size))
			}()
		}
	}
	return n, err
}
