package models

type PushRemoteMachineRequest struct {
	LocalPath string `json:"local_path"`
	Name      string `json:"name"`
	Uuid      string `json:"uuid,omitempty"`
}

type PullRemoteMachineRequest struct {
	Name string `json:"name"`
}
