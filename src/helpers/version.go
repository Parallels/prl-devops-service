package helpers

import (
	"fmt"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func NewVersion(version string) *Version {
	v := &Version{}
	fmt.Sscanf(version, "%d.%d.%d", &v.Major, &v.Minor, &v.Patch)
	return v
}

func (v *Version) LessThan(other *Version) bool {
	if v.Major < other.Major {
		return true
	}
	if v.Major > other.Major {
		return false
	}
	if v.Minor < other.Minor {
		return true
	}
	if v.Minor > other.Minor {
		return false
	}
	return v.Patch < other.Patch
}
