package models

type PerformanceRequest struct {
	TestCount                   int `json:"test_count"`
	ConsecutiveCalls            int `json:"consecutive_calls"`
	TimeBetweenConsecutiveCalls int `json:"time_between_consecutive_calls"`
	TimeBetweenCalls            int `json:"time_between_calls"`
}
