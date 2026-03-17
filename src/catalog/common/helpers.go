package common

import (
	"errors"
	"fmt"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
)

const (
	minPartSize int64 = 10 * 1024 * 1024 // 10 MB — well above the S3 5 MB protocol minimum
	maxPartSize int64 = 64 * 1024 * 1024 // 64 MB — 5 concurrent parts = 320 MB in-flight; minio acks quickly
	targetParts int64 = 200              // aim for ~200 parts to balance round-trip overhead vs. ack latency
)

// CalculatePartSize returns an upload part size appropriate for the given file
// size. It targets ~200 parts, clamped between 10 MB and 64 MB.
func CalculatePartSize(fileSize int64) int64 {
	if fileSize <= 0 {
		return minPartSize
	}
	size := fileSize / targetParts
	if size < minPartSize {
		return minPartSize
	}
	if size > maxPartSize {
		return maxPartSize
	}
	return size
}

func ValidateArch(arch string) (string, error) {
	currentArch := arch
	if arch == "" {
		ctx := basecontext.NewRootBaseContext()
		svcCtl := system.Get()
		arch, err := svcCtl.GetArchitecture(ctx)
		if err != nil {
			return "", errors.New("unable to determine architecture")
		}

		currentArch = arch
	}

	if currentArch == "amd64" {
		currentArch = "x86_64"
	}
	if currentArch == "arm" {
		currentArch = "arm64"
	}
	if currentArch == "aarch64" {
		currentArch = "arm64"
	}

	if currentArch != "x86_64" && currentArch != "arm64" {
		return "", errors.New("invalid architecture")
	}

	return currentArch, nil
}

func ValidatePath(path string, owner string) (string, error) {
	ctx := basecontext.NewRootBaseContext()
	if path == "" {
		prl := serviceprovider.Get().ParallelsDesktopService
		if prl == nil {
			return "", errors.New("Local Path is required and we are unable to determine it without Parallels Desktop Service")
		}
		userPath, err := prl.GetUserHome(ctx, owner)
		if err != nil {
			return "", fmt.Errorf("unable to determine user %v home for path", owner)
		}
		path = userPath
	}

	return path, nil
}
