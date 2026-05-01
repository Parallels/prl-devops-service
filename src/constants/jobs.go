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

// Ghost job cleanup configuration
const (
	// DefaultGhostJobTimeoutMinutes is the default inactivity threshold in minutes.
	// Jobs not updated within this window are considered "ghost jobs" and marked as failed.
	DefaultGhostJobTimeoutMinutes = 5

	// GhostJobCheckIntervalSeconds is how often the stale-job checker runs.
	GhostJobCheckIntervalSeconds = 5

	// GhostJobCanceledReason is the error message set when a ghost job is canceled.
	GhostJobCanceledReason = "Canceled Job"

	// GhostJobTimeoutMinutesEnvVar is the environment variable to override the default timeout.
	GhostJobTimeoutMinutesEnvVar = "GHOST_JOB_TIMEOUT_MINUTES"
)
