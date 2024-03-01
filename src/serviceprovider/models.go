package serviceprovider

type ToolOperationResult struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	InstallPath string `json:"installPath"`
	Result      bool   `json:"result"`
	Message     string `json:"errorMessage"`
}
