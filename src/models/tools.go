package models

import "github.com/Parallels/prl-devops-service/errors"

type InstallToolsRequest struct {
	All   bool                               `json:"all"`
	RunAs string                             `json:"run_as"`
	Tools map[string]InstallToolsRequestItem `json:"tools"`
}

type InstallToolsRequestItem struct {
	Version string            `json:"version"`
	Flags   map[string]string `json:"flags"`
}

func (i *InstallToolsRequest) Validate() error {
	if i.Tools == nil && !i.All {
		return errors.New("tools cannot be empty")
	}
	for _, tool := range i.Tools {
		if tool.Version == "" {
			tool.Version = "latest"
		}
		if tool.Flags == nil {
			tool.Flags = make(map[string]string)
		}
	}

	return nil
}

type InstallToolsResponse struct {
	Success        bool                                `json:"success"`
	InstalledTools map[string]InstallToolsResponseItem `json:"installed_tools"`
}

type InstallToolsResponseItem struct {
	Success      bool   `json:"success"`
	Version      string `json:"version,omitempty"`
	ErrorMessage string `json:"message,omitempty"`
}

type UninstallToolsRequest struct {
	All                   bool                                 `json:"all"`
	UninstallDependencies bool                                 `json:"uninstall_dependencies"`
	RunAs                 string                               `json:"run_as"`
	Tools                 map[string]UninstallToolsRequestItem `json:"tools"`
}

type UninstallToolsRequestItem struct {
	UninstallDependencies bool              `json:"uninstall_dependencies"`
	Flags                 map[string]string `json:"flags"`
}

func (u *UninstallToolsRequest) Validate() error {
	if u.Tools == nil && !u.All {
		return errors.New("tools cannot be empty")
	}
	for _, tool := range u.Tools {
		if tool.Flags == nil {
			tool.Flags = make(map[string]string)
		}
	}

	return nil
}

type UninstallToolsResponse struct {
	Success          bool                                  `json:"success"`
	UninstalledTools map[string]UninstallToolsResponseItem `json:"uninstalled_tools"`
}

type UninstallToolsResponseItem struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"message,omitempty"`
}
