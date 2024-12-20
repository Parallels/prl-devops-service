package writers

import (
	"io"
	"sync"
)

type ByteSliceWriterAt struct {
	data []byte
	mu   sync.Mutex
}

func NewByteSliceWriterAt(size int64) *ByteSliceWriterAt {
	return &ByteSliceWriterAt{
		data: make([]byte, size),
	}
}

func (b *ByteSliceWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if off < 0 || int(off) >= len(b.data) {
		return 0, io.EOF
	}

	n = copy(b.data[off:], p)
	if n < len(p) {
		err = io.EOF
	}
	return n, err
}

func (b *ByteSliceWriterAt) Bytes() []byte {
	return b.data
}
