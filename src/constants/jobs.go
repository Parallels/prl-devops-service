package constants

type JobState string

const (
	JobStatePending   JobState = "pending"
	JobStateRunning   JobState = "running"
	JobStateCompleted JobState = "completed"
	JobStateFailed    JobState = "failed"
)

func (js JobState) String() string {
	return string(js)
}

func (js JobState) IsValid() bool {
	switch js {
	case JobStatePending, JobStateRunning, JobStateCompleted, JobStateFailed:
		return true
	default:
		return false
	}
}

func GetAllJobStates() []JobState {
	return []JobState{
		JobStatePending,
		JobStateRunning,
		JobStateCompleted,
		JobStateFailed,
	}
}
