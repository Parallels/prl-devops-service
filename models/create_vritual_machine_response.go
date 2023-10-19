package models

type CreateVirtualMachineResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Owner        string `json:"owner"`
	CurrentState string `json:"current_state"`
}
