package parallels_desktop

type ParallelsVirtualMachineState string

const (
	ParallelsVirtualMachineStateStopped   ParallelsVirtualMachineState = "stopped"
	ParallelsVirtualMachineStateRunning   ParallelsVirtualMachineState = "running"
	ParallelsVirtualMachineStateSuspended ParallelsVirtualMachineState = "suspended"
	ParallelsVirtualMachineStatePaused    ParallelsVirtualMachineState = "paused"
	ParallelsVirtualMachineStateUnknown   ParallelsVirtualMachineState = "unknown"
)

func (s ParallelsVirtualMachineState) String() string {
	return string(s)
}
