package writers

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/jobs/tracker"
)

type ProgressWriter struct {
	ns             *tracker.JobProgressService
	writer         io.Writer
	correlationId  string
	totalProcessed int64
	size           int64
	filename       string
	prefix         string
	jobId          string
	currentAction  string
	mu             sync.Mutex
}

func NewProgressWriter(writer io.Writer, size int64, action string) *ProgressWriter {
	return &ProgressWriter{
		correlationId: helpers.GenerateId(),
		ns:            tracker.GetProgressService(),
		writer:        writer,
		size:          size,
		currentAction: action,
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

func (pr *ProgressWriter) SetJobId(jobId string) {
	pr.jobId = jobId
}

func (pr *ProgressWriter) SetCurrentAction(action string) {
	pr.currentAction = action
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
			percentage := float64(off*100) / float64(pw.size)
			if pw.ns != nil {
				prefix := pw.prefix
				if pw.filename != "" {
					prefix = fmt.Sprintf("%s %s", prefix, pw.filename)
				}

				if pw.jobId != "" && !strings.HasPrefix(prefix, "["+pw.jobId+"]") {
					prefix = fmt.Sprintf("[%s] %s", pw.jobId, prefix)
				}

				msg := tracker.NewJobProgressMessage(pw.correlationId, prefix, percentage).
					WithTransfer(off, pw.size).
					WithJob(pw.jobId, pw.currentAction).
					SetFilename(pw.filename)
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
			percentage := float64(pw.totalProcessed*100) / float64(pw.size)
			if pw.ns != nil {
				prefix := pw.prefix
				if pw.filename != "" {
					prefix = fmt.Sprintf("%s %s", prefix, pw.filename)
				}

				if pw.jobId != "" && !strings.HasPrefix(prefix, "["+pw.jobId+"]") {
					prefix = fmt.Sprintf("[%s] %s", pw.jobId, prefix)
				}

				msg := tracker.NewJobProgressMessage(pw.correlationId, prefix, percentage).
					WithTransfer(pw.totalProcessed, pw.size).
					WithJob(pw.jobId, pw.currentAction).
					SetFilename(pw.filename)

				pw.ns.Notify(msg)
			}
		}
	}
	return n, err
}
