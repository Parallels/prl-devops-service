package parallels_desktop

type ParallelsVirtualMachineDesiredState string

const (
	ParallelsVirtualMachineDesiredStateStop    ParallelsVirtualMachineDesiredState = "stop"
	ParallelsVirtualMachineDesiredStateStart   ParallelsVirtualMachineDesiredState = "start"
	ParallelsVirtualMachineDesiredStatePause   ParallelsVirtualMachineDesiredState = "pause"
	ParallelsVirtualMachineDesiredStateSuspend ParallelsVirtualMachineDesiredState = "suspend"
	ParallelsVirtualMachineDesiredStateResume  ParallelsVirtualMachineDesiredState = "resume"
	ParallelsVirtualMachineDesiredStateReset   ParallelsVirtualMachineDesiredState = "reset"
	ParallelsVirtualMachineDesiredStateRestart ParallelsVirtualMachineDesiredState = "restart"
	ParallelsVirtualMachineDesiredStateUnknown ParallelsVirtualMachineDesiredState = "unknown"
)

func (s ParallelsVirtualMachineDesiredState) String() string {
	return string(s)
}

func ParallelsVirtualMachineDesiredStateFromString(s string) ParallelsVirtualMachineDesiredState {
	switch s {
	case "stop":
		return ParallelsVirtualMachineDesiredStateStop
	case "start":
		return ParallelsVirtualMachineDesiredStateStart
	case "pause":
		return ParallelsVirtualMachineDesiredStatePause
	case "suspend":
		return ParallelsVirtualMachineDesiredStateSuspend
	case "resume":
		return ParallelsVirtualMachineDesiredStateResume
	case "reset":
		return ParallelsVirtualMachineDesiredStateReset
	case "restart":
		return ParallelsVirtualMachineDesiredStateRestart
	default:
		return ParallelsVirtualMachineDesiredStateUnknown
	}
}
