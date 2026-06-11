package models

// ParallelsDesktopLicense represents the parsed output of prlsrvctl info --license -j.
type ParallelsDesktopLicense struct {
	Status             string `json:"status"`
	Serial             string `json:"serial"`
	Expiration         string `json:"expiration"`
	MainPeriodEndsAt   string `json:"main_period_ends_at"`
	GracePeriodEndsAt  string `json:"grace_period_ends_at"`
	CPUTotal           int64  `json:"cpu_total"`
	MaxMemory          int64  `json:"max_memory"`
	Edition            string `json:"edition"`
	IsTrial            string `json:"is_trial"`
	IsVolume           string `json:"is_volume"`
	DeferredActivation string `json:"deferred_activation"`
	UUID               string `json:"uuid"`
}

// IsActive returns true if the license status is "ACTIVE".
func (l *ParallelsDesktopLicense) IsActive() bool {
	return l.Status == "ACTIVE"
}

// IsTrialLicense returns true if this is a trial license.
func (l *ParallelsDesktopLicense) IsTrialLicense() bool {
	return l.IsTrial == "yes"
}

// IsVolumeLicense returns true if this is a volume license.
func (l *ParallelsDesktopLicense) IsVolumeLicense() bool {
	return l.IsVolume == "yes"
}
