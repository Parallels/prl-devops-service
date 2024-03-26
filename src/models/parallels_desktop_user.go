package models

type ParallelsDesktopUsers []ParallelsDesktopUser

type ParallelsDesktopUser struct {
	Name        string `json:"NAME"`
	MNGSettings string `json:"MNG_SETTINGS"`
	DefVMHome   string `json:"DEF_VM_HOME"`
}
