package models

type VirtualMachineOperationResponse struct {
	ID        string `json:"id"`
	Operation string `json:"operation"`
	Status    string `json:"status"`
}
