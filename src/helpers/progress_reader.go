package helpers

import (
	"os"
	"sync/atomic"
)

type ProgressReader struct {
	file     *os.File
	size     int64
	read     int64
	progress chan int
}

func NewProgressReader(file *os.File, size int64, progress chan int) *ProgressReader {
	return &ProgressReader{
		file:     file,
		size:     size,
		progress: progress,
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	return pr.file.Read(p)
}

func (r *ProgressReader) ReadAt(p []byte, off int64) (int, error) {
	n, err := r.file.ReadAt(p, off)
	if err != nil {
		return n, err
	}

	atomic.AddInt64(&r.read, int64(n))

	if r.progress != nil && r.size > 0 {
		go func() {
			r.progress <- int(float32(r.read*100/2) / float32(r.size))
		}()
	}

	return n, err
}

func (r *ProgressReader) Seek(offset int64, whence int) (int64, error) {
	return r.file.Seek(offset, whence)
}
