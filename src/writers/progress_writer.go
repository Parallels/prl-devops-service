package writers

import (
	"fmt"
	"io"
	"sync"

	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/notifications"
)

type ProgressWriter struct {
	ns             *notifications.NotificationService
	writer         io.Writer
	correlationId  string
	totalProcessed int64
	size           int64
	filename       string
	prefix         string
	mu             sync.Mutex
}

func NewProgressWriter(writer io.Writer, size int64) *ProgressWriter {
	return &ProgressWriter{
		correlationId: helpers.GenerateId(),
		ns:            notifications.Get(),
		writer:        writer,
		size:          size,
	}
}

func (pr *ProgressWriter) SetFilename(filename string) {
	pr.filename = filename
}

func (pr *ProgressWriter) SetPrefix(prefix string) {
	pr.prefix = prefix
}

func (pr *ProgressWriter) SetCorrelationId(correlationId string) {
	pr.correlationId = correlationId
}

func (pr *ProgressWriter) CorrelationId() string {
	return pr.correlationId
}

func (pr *ProgressWriter) Size() int64 {
	return pr.size
}

func (pr *ProgressWriter) GetWriterAt() io.WriterAt {
	if ra, ok := pr.writer.(io.WriterAt); ok {
		return ra
	}

	return nil
}

func (pw *ProgressWriter) WriteAt(p []byte, off int64) (n int, err error) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	if _, ok := pw.writer.(io.WriterAt); !ok {
		return 0, fmt.Errorf("underlying writer does not support WriteAt")
	}

	n, err = pw.writer.(io.WriterAt).WriteAt(p, off)
	if err == nil {
		if pw.size > 0 {
			percentage := int(float32(off*100) / float32(pw.size))
			if pw.ns != nil {
				prefix := pw.prefix
				if prefix == "" {
					prefix = "Processing"
				}
				if pw.filename != "" {
					prefix = fmt.Sprintf("%s %s", prefix, pw.filename)
				}
				msg := notifications.NewProgressNotificationMessage(pw.correlationId, prefix, percentage).
					SetCurrentSize(off).
					SetTotalSize(pw.size)
				pw.ns.Notify(msg)
			}
		}
	}
	return n, err
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	n, err := pw.writer.Write(p)
	pw.totalProcessed += int64(len(p))
	if err == nil {
		if pw.size > 0 {
			percentage := int(float32(pw.totalProcessed*100) / float32(pw.size))
			if pw.ns != nil {
				prefix := pw.prefix
				if prefix == "" {
					prefix = "Processing"
				}
				if pw.filename != "" {
					prefix = fmt.Sprintf("%s %s", prefix, pw.filename)
				}
				msg := notifications.NewProgressNotificationMessage(pw.correlationId, prefix, percentage).
					SetCurrentSize(pw.totalProcessed).
					SetTotalSize(pw.size)

				pw.ns.Notify(msg)
			}
		}
	}
	return n, err
}
