package models

import "errors"

type VirtualMachineTemplate struct {
	ID           string                     `json:"id"`
	Name         string                     `json:"name"`
	Owner        string                     `json:"owner"`
	Hostname     string                     `json:"hostname"`
	Type         VirtualMachineTemplateType `json:"type"`
	Description  string                     `json:"description"`
	PackerFolder string                     `json:"packer_folder"`
	RemoteUrl    string                     `json:"remote_url"`
	Variables    map[string]string          `json:"variables"`
	Addons       []string                   `json:"addons"`
	Specs        map[string]int             `json:"specs"`
	Defaults     map[string]string          `json:"defaults"`
}

func (m *VirtualMachineTemplate) Validate() error {
	if m.ID == "" {
		return errors.New("ID cannot be empty")
	}

	if m.Name == "" {
		return errors.New("Name cannot be empty")
	}

	if m.Type == -1 {
		return errors.New("Type cannot be empty")
	}

	if m.Type == VirtualMachineTemplateTypePacker && m.PackerFolder == "" {
		return errors.New("PackerFolder cannot be empty")
	}

	if m.Type == VirtualMachineTemplateTypeRemote && m.RemoteUrl == "" {
		return errors.New("RemoteUrl cannot be empty")
	}

	if m.Specs == nil {
		m.Specs = make(map[string]int)
		m.Specs["memory"] = 2048
		m.Specs["cpu"] = 2
		m.Specs["disk"] = 20480
	}

	if m.Defaults == nil {
		m.Defaults = make(map[string]string)
	}

	return nil
}
