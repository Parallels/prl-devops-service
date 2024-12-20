package writers

import (
	"fmt"
	"os"
	"sync/atomic"

	"github.com/Parallels/prl-devops-service/notifications"
)

type ProgressFileReader struct {
	ns            *notifications.NotificationService
	file          *os.File
	correlationId string
	size          int64
	read          int64
	prefix        string
}

func NewProgressFileReader(file *os.File, size int64) *ProgressFileReader {
	return &ProgressFileReader{
		file: file,
		size: size,
		ns:   notifications.Get(),
	}
}

func (pr *ProgressFileReader) SetPrefix(prefix string) {
	pr.prefix = prefix
}

func (pr *ProgressFileReader) SetCorrelationId(correlationId string) {
	pr.correlationId = correlationId
}

func (pr *ProgressFileReader) CorrelationId() string {
	return pr.correlationId
}

func (pr *ProgressFileReader) Size() int64 {
	return pr.size
}

func (pr *ProgressFileReader) Read(p []byte) (int, error) {
	n, err := pr.file.Read(p)
	if n > 0 {
		newRead := atomic.AddInt64(&pr.read, int64(n))
		if pr.size > 0 {
			percentage := int(float64(newRead) * 100 / float64(pr.size))
			if pr.ns != nil {
				message := pr.prefix
				if message == "" {
					message = "Processing"
				}
				if pr.file.Name() != "" {
					message = fmt.Sprintf("%s %s", message, pr.file.Name())
				}
				msg := notifications.NewProgressNotificationMessage(pr.correlationId, message, percentage).
					SetCurrentSize(newRead).
					SetTotalSize(pr.size)
				pr.ns.Notify(msg)
			}
		}
	}

	return n, err
}

func (pr *ProgressFileReader) ReadAt(p []byte, off int64) (int, error) {
	n, err := pr.file.ReadAt(p, off)
	if err != nil {
		return n, err
	}

	newRead := atomic.AddInt64(&pr.read, int64(n))

	if pr.size > 0 {
		if newRead > pr.size {
			newRead = pr.size
		}
		percentage := int(float64(newRead) * 100 / float64(pr.size))
		if pr.ns != nil {
			message := pr.prefix
			if message == "" {
				message = "Processing"
			}
			if pr.file.Name() != "" {
				message = fmt.Sprintf("%s %s", message, pr.file.Name())
			}

			msg := notifications.NewProgressNotificationMessage(pr.file.Name(), message, percentage).
				SetCurrentSize(newRead).
				SetTotalSize(pr.size)
			pr.ns.Notify(msg)
		}
	}

	return n, err
}

func (r *ProgressFileReader) Seek(offset int64, whence int) (int64, error) {
	return r.file.Seek(offset, whence)
}
