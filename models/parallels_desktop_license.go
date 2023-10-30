package models

type ParallelsDesktopLicense struct {
	Status               string `json:"status"`
	Serial               string `json:"serial"`
	Expiration           string `json:"expiration"`
	MainPeriodEndsAt     string `json:"main_period_ends_at"`
	GracePeriodEndsAt    string `json:"grace_period_ends_at"`
	CPUTotal             int64  `json:"cpu_total"`
	MaxMemory            int64  `json:"max_memory"`
	Edition              string `json:"edition"`
	IsVolume             string `json:"is_volume"`
	AdvancedRestrictions string `json:"advanced_restrictions"`
	DeferredActivation   string `json:"deferred_activation"`
	UUID                 string `json:"uuid"`
}
