package constants

type JobState string

const (
	JobStateInit      JobState = "init"
	JobStatePending   JobState = "pending"
	JobStateRunning   JobState = "running"
	JobStateCompleted JobState = "completed"
	JobStateFailed    JobState = "failed"
	JobStateSkipped   JobState = "skipped"
)

func (js JobState) String() string {
	return string(js)
}

func (js JobState) IsValid() bool {
	switch js {
	case JobStateInit, JobStatePending, JobStateRunning, JobStateCompleted, JobStateFailed, JobStateSkipped:
		return true
	default:
		return false
	}
}

func GetAllJobStates() []JobState {
	return []JobState{
		JobStateInit,
		JobStatePending,
		JobStateRunning,
		JobStateCompleted,
		JobStateFailed,
		JobStateSkipped,
	}
}
