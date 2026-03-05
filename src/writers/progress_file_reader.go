package writers

import (
	"fmt"
	"os"
	"strings"
	"sync" // Added sync import for sync.Mutex
	"sync/atomic"

	"github.com/Parallels/prl-devops-service/jobs/tracker"
)

type ProgressFileReader struct {
	ns            *tracker.JobProgressService
	file          *os.File
	correlationId string
	size          int64
	read          int64
	prefix        string
	jobId         string
	currentAction string
	mu            sync.Mutex // Added mu
}

func NewProgressFileReader(file *os.File, size int64, action string) *ProgressFileReader {
	return &ProgressFileReader{
		file:          file,
		size:          size,
		ns:            tracker.GetProgressService(),
		currentAction: action,
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

func (pr *ProgressFileReader) SetJobId(jobId string) {
	pr.jobId = jobId
}

func (pr *ProgressFileReader) SetCurrentAction(action string) {
	pr.currentAction = action
}

func (pr *ProgressFileReader) Size() int64 {
	return pr.size
}

func (pr *ProgressFileReader) Read(p []byte) (int, error) {
	n, err := pr.file.Read(p)
	if n > 0 {
		newRead := atomic.AddInt64(&pr.read, int64(n))
		if pr.size > 0 {
			percentage := float64(newRead) * 100 / float64(pr.size)
			if pr.ns != nil {
				message := pr.prefix
				if pr.file.Name() != "" {
					message = fmt.Sprintf("%s %s", message, pr.file.Name())
				}

				if pr.jobId != "" && !strings.HasPrefix(message, "["+pr.jobId+"]") {
					message = fmt.Sprintf("[%s] %s", pr.jobId, message)
				}

				msg := tracker.NewJobProgressMessage(pr.correlationId, message, percentage).
					SetCurrentSize(newRead).
					SetTotalSize(pr.size).
					SetJobId(pr.jobId).
					SetCurrentAction(pr.currentAction).
					SetFilename(pr.file.Name())
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
		percentage := float64(newRead) * 100 / float64(pr.size)
		if pr.ns != nil {
			message := pr.prefix
			if pr.file.Name() != "" {
				message = fmt.Sprintf("%s %s", message, pr.file.Name())
			}

			if pr.jobId != "" && !strings.HasPrefix(message, "["+pr.jobId+"]") {
				message = fmt.Sprintf("[%s] %s", pr.jobId, message)
			}

			msg := tracker.NewJobProgressMessage(pr.correlationId, message, percentage).
				SetCurrentSize(newRead).
				SetTotalSize(pr.size).
				SetJobId(pr.jobId).
				SetCurrentAction(pr.currentAction).
				SetFilename(pr.file.Name())
			pr.ns.Notify(msg)
		}
	}

	return n, err
}

func (r *ProgressFileReader) Seek(offset int64, whence int) (int64, error) {
	return r.file.Seek(offset, whence)
}
