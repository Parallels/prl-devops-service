package writers

import (
	"fmt"
	"io"
	"sync/atomic"

	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/notifications"
)

type ProgressReader struct {
	ns            *notifications.NotificationService
	reader        io.Reader
	correlationId string
	size          int64
	read          int64
	filename      string
	prefix        string
}

func NewProgressReader(reader io.Reader, size int64) *ProgressReader {
	return &ProgressReader{
		correlationId: helpers.GenerateId(),
		ns:            notifications.Get(),
		reader:        reader,
		size:          size,
	}
}

func (pr *ProgressReader) SetFilename(filename string) {
	pr.filename = filename
}

func (pr *ProgressReader) SetPrefix(prefix string) {
	pr.prefix = prefix
}

func (pr *ProgressReader) SetCorrelationId(correlationId string) {
	pr.correlationId = correlationId
}

func (pr *ProgressReader) CorrelationId() string {
	return pr.correlationId
}

func (pr *ProgressReader) Size() int64 {
	return pr.size
}

func (pr *ProgressReader) GetReaderAt() io.ReaderAt {
	if ra, ok := pr.reader.(io.ReaderAt); ok {
		return ra
	}

	return nil
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		newRead := atomic.AddInt64(&pr.read, int64(n))
		if pr.size > 0 {
			percentage := float64(newRead) * 100 / float64(pr.size)
			if pr.ns != nil {
				prefix := pr.prefix
				if prefix == "" {
					prefix = "Processing"
				}
				if pr.filename != "" {
					prefix = fmt.Sprintf("%s %s", prefix, pr.filename)
				}
				msg := notifications.NewProgressNotificationMessage(pr.correlationId, prefix, percentage).
					SetCurrentSize(newRead).
					SetTotalSize(pr.size)
				pr.ns.Notify(msg)
			}
		}
	}
	return n, err
}

func (pr *ProgressReader) ReadAt(p []byte, off int64) (int, error) {
	ra, ok := pr.reader.(io.ReaderAt)
	if !ok {
		return 0, fmt.Errorf("underlying reader does not support ReadAt")
	}
	n, err := ra.ReadAt(p, off)
	if err != nil {
		return n, err
	}

	if err == io.EOF {
		newRead := atomic.AddInt64(&pr.read, int64(n))

		if pr.size > 0 {
			if newRead > pr.size {
				newRead = pr.size
			}
			percentage := float64(newRead) * 100 / float64(pr.size)
			if pr.ns != nil {
				prefix := pr.prefix
				if prefix == "" {
					prefix = "Processing"
				}
				if pr.filename != "" {
					prefix = fmt.Sprintf("%s %s", prefix, pr.filename)
				}
				msg := notifications.NewProgressNotificationMessage(pr.correlationId, prefix, percentage).
					SetCurrentSize(newRead).
					SetTotalSize(pr.size)
				pr.ns.Notify(msg)
			}
		}

	}

	return n, err
}

func (pr *ProgressReader) Seek(offset int64, whence int) (int64, error) {
	seeker, ok := pr.reader.(io.Seeker)
	if !ok {
		return 0, fmt.Errorf("underlying reader does not support Seek")
	}

	return seeker.Seek(offset, whence)
}
