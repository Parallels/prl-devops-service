package models

import "errors"

type VirtualMachineTemplateType int

const (
	VirtualMachineTemplateTypePacker VirtualMachineTemplateType = iota
	VirtualMachineTemplateTypeRemote
)

func (m VirtualMachineTemplateType) String() string {
	return [...]string{"packer", "parallels"}[m]
}

func (m VirtualMachineTemplateType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.String() + `"`), nil
}

func (m *VirtualMachineTemplateType) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case `"packer"`:
		*m = VirtualMachineTemplateTypePacker
	case `"parallels"`:
		*m = VirtualMachineTemplateTypeRemote
	default:
		return errors.New("invalid VirtualMachineTemplateType")
	}
	return nil
}
