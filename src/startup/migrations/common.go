package migrations

import (
	"errors"
	"strconv"
	"strings"
)

type VersionComparisonResult int

const (
	VersionLowerThanTarget  VersionComparisonResult = -1
	VersionEqualToTarget    VersionComparisonResult = 0
	VersionHigherThanTarget VersionComparisonResult = 1
)

func compareVersions(version, targetVersion string) (VersionComparisonResult, error) {
	parts := strings.Split(version, ".")
	if len(parts) > 3 {
		return -1, errors.New("invalid version format")
	}
	if len(parts) == 1 {
		parts = append(parts, "0")
	}
	if len(parts) == 2 {
		parts = append(parts, "0")
	}

	targetParts := strings.Split(targetVersion, ".")
	if len(targetParts) > 3 {
		return -1, errors.New("invalid version format")
	}

	if len(targetParts) == 1 {
		targetParts = append(targetParts, "0")
	}
	if len(targetParts) == 2 {
		targetParts = append(targetParts, "0")
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return -1, errors.New("invalid version format")
	}
	targetMajor, err := strconv.Atoi(targetParts[0])
	if err != nil {
		return -1, errors.New("invalid version format")
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1, errors.New("invalid version format")
	}
	targetMinor, err := strconv.Atoi(targetParts[1])
	if err != nil {
		return -1, errors.New("invalid version format")
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return -1, errors.New("invalid version format")
	}
	targetPatch, err := strconv.Atoi(targetParts[2])
	if err != nil {
		return -1, errors.New("invalid version format")
	}

	if major > targetMajor {
		return 1, nil
	}
	if major < targetMajor {
		return -1, nil
	}
	if minor > targetMinor {
		return 1, nil
	}
	if minor < targetMinor {
		return -1, nil
	}
	if patch > targetPatch {
		return 1, nil
	}
	if patch < targetPatch {
		return -1, nil
	}

	return 0, nil
}
