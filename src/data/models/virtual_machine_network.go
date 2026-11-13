package models

type VirtualMachineNetworkInformation struct {
	Conditioned string                                      `json:"conditioned"`
	Inbound     VirtualMachineNetworkInformationBound       `json:"inbound"`
	Outbound    VirtualMachineNetworkInformationBound       `json:"outbound"`
	IPAddresses []VirtualMachineNetworkInformationIPAddress `json:"ip_addresses"`
}

func (m *VirtualMachineNetworkInformation) Diff(source VirtualMachineNetworkInformation) bool {
	if m.Conditioned != source.Conditioned {
		return true
	}

	if m.Inbound.Diff(source.Inbound) {
		return true
	}

	if m.Outbound.Diff(source.Outbound) {
		return true
	}

	if len(m.IPAddresses) != len(source.IPAddresses) {
		return true
	}
	for i, v := range m.IPAddresses {
		if v.Diff(source.IPAddresses[i]) {
			return true
		}
	}

	return false
}

type VirtualMachineNetworkInformationIPAddress struct {
	Type string `json:"type"`
	IP   string `json:"ip"`
}

func (m *VirtualMachineNetworkInformationIPAddress) Diff(source VirtualMachineNetworkInformationIPAddress) bool {
	if m.Type != source.Type {
		return true
	}
	if m.IP != source.IP {
		return true
	}

	return false
}

type VirtualMachineNetworkInformationBound struct {
	Bandwidth  string `json:"bandwidth"`
	PacketLoss string `json:"packet_loss"`
	Delay      string `json:"delay"`
}

func (m *VirtualMachineNetworkInformationBound) Diff(source VirtualMachineNetworkInformationBound) bool {
	if m.Bandwidth != source.Bandwidth {
		return true
	}
	if m.PacketLoss != source.PacketLoss {
		return true
	}
	if m.Delay != source.Delay {
		return true
	}

	return false
}
