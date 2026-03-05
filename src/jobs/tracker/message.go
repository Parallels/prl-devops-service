package tracker

import "time"

type JobMessage struct {
	correlationId        string
	Message              string
	CurrentProgress      float64
	JobId                string
	JobPercentage        float64
	CurrentAction        string
	CurrentActionStep    string
	Filename             string
	totalSize            int64
	currentSize          int64
	IsProgress           bool
	prefix               string
	closed               bool
	startingTime         time.Time
	lastNotificationTime time.Time
	Level                JobMessageLevel
}

func NewJobMessage(message string, level JobMessageLevel) *JobMessage {
	return &JobMessage{
		Message: message,
		Level:   level,
	}
}

func NewJobProgressMessage(correlationId string, message string, progress float64) *JobMessage {
	cid := normalizeCorrelationID(correlationId)
	return &JobMessage{
		correlationId:        cid,
		Message:              message,
		CurrentProgress:      progress,
		lastNotificationTime: time.Now(),
		IsProgress:           true,
	}
}

func (nm *JobMessage) String() string {
	return nm.Message
}

func (nm *JobMessage) SetCorrelationId(id string) *JobMessage {
	nm.correlationId = normalizeCorrelationID(id)
	return nm
}

func (nm *JobMessage) CorrelationId() string {
	return nm.correlationId
}

func (nm *JobMessage) SetTotalSize(size int64) *JobMessage {
	nm.totalSize = size
	return nm
}

func (nm *JobMessage) TotalSize() int64 {
	return nm.totalSize
}

func (nm *JobMessage) SetCurrentSize(size int64) *JobMessage {
	nm.currentSize = size
	return nm
}

func (nm *JobMessage) CurrentSize() int64 {
	return nm.currentSize
}

func (nm *JobMessage) SetPrefix(prefix string) *JobMessage {
	nm.prefix = prefix
	return nm
}

func (nm *JobMessage) SetStartingTime(startingTime time.Time) *JobMessage {
	nm.startingTime = startingTime
	return nm
}

func (nm *JobMessage) Prefix() string {
	return nm.prefix
}

func (nm *JobMessage) Closed() bool {
	return nm.closed
}

func (nm *JobMessage) SetJobId(jobId string) *JobMessage {
	nm.JobId = jobId
	return nm
}

func (nm *JobMessage) SetJobPercentage(jobPercentage float64) *JobMessage {
	nm.JobPercentage = jobPercentage
	return nm
}

func (nm *JobMessage) SetCurrentAction(action string) *JobMessage {
	nm.CurrentAction = action
	return nm
}

func (nm *JobMessage) SetCurrentActionStep(step string) *JobMessage {
	nm.CurrentActionStep = step
	return nm
}

func (nm *JobMessage) SetFilename(filename string) *JobMessage {
	nm.Filename = filename
	return nm
}

func (nm *JobMessage) Close() *JobMessage {
	nm.closed = true
	return nm
}
