package tracker

import (
	"encoding/base64"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
)

type JobMessageLevel int

const (
	JobMessageLevelInfo    JobMessageLevel = iota
	JobMessageLevelWarning JobMessageLevel = iota
	JobMessageLevelError   JobMessageLevel = iota
	JobMessageLevelDebug   JobMessageLevel = iota
)

const (
	minUpdateInterval         = 500 * time.Millisecond
	significantProgressChange = 1.0
	minSampleInterval         = 2 * time.Second
	uiProgressPushThreshold   = 1.0
)

type RateSample struct {
	Timestamp time.Time
	Size      int64
	Progress  float64
}

// JobStep defines a single step in a job workflow.
// Name is the unique identifier used internally for map keys and function targeting.
// DisplayName is the human-readable label shown in the UI; if empty, Name is used.
type JobStep struct {
	Name          string
	DisplayName   string
	Weight        float64
	Parallel      bool
	HasPercentage bool
}

type JobWorkflow struct {
	Message      string
	Steps        []JobStep
	StepProgress map[string]float64
	StepMessage  map[string]string
	StepError    map[string]string
	StepValue    map[string]int64
	StepTotal    map[string]int64
	StepFilename map[string]string
	StepState    map[string]constants.JobState
}

type ProgressTracker struct {
	CurrentProgress        float64
	JobPercentage          float64
	LastJobPercentageSent  float64
	LastStepPercentageSent float64
	LastLogPercentageSent  float64
	JobId                  string
	CurrentAction          string
	CurrentActionStep      string
	Filename               string
	LastUpdateTime         time.Time
	LastLogTime            time.Time
	Prefix                 string
	StartTime              time.Time
	TotalSize              int64
	CurrentSize            int64
	IsComplete             bool
	RateSamples            []RateSample
}

type SpeedTrend struct {
	Increasing bool
	Stable     bool
	Factor     float64
}

// ProgressRate contains rate information for a progress job message.
type ProgressRate struct {
	BytesPerSecond       float64
	RecentBytesPerSecond float64
	ProgressPerSecond    float64
}

func normalizeCorrelationID(id string) string {
	if id == "" {
		return ""
	}

	decoded, err := base64.StdEncoding.DecodeString(id)
	if err == nil && base64.StdEncoding.EncodeToString(decoded) == id {
		return id
	}

	return base64.StdEncoding.EncodeToString([]byte(id))
}

func decodeCorrelationID(id string) (string, error) {
	if id == "" {
		return "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}
