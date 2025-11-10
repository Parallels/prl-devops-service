package models

type AdditionalInfo struct {
	VmStateName string `json:"Vm state name"`
}

type ParallelsServiceEvent struct {
	Timestamp      string          `json:"Timestamp"`
	VMID           string          `json:"VM ID"`
	EventName      string          `json:"Event name"`
	AdditionalInfo *AdditionalInfo `json:"Additional info,omitempty"`
}
