package processlauncher

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

type ProcessLauncher interface {
	Start(cmd *exec.Cmd) (*os.File, error)
}

// RealProcessLauncher implements ProcessLauncher using pty.Start.
type RealProcessLauncher struct{}

func (r *RealProcessLauncher) Start(cmd *exec.Cmd) (*os.File, error) {
	return pty.Start(cmd)
}

// MockProcessLauncher is a mock implementation of ProcessLauncher for testing purposes.
// It allows injecting custom behavior via the LaunchFunc field to simulate different scenarios without actual PTY dependencies.
type MockProcessLauncher struct {
	LaunchFunc func(cmd *exec.Cmd) (*os.File, error)
}

func (m *MockProcessLauncher) Start(cmd *exec.Cmd) (*os.File, error) {
	if m.LaunchFunc != nil {
		return m.LaunchFunc(cmd)
	}
	// Default behavior: return nil if no function is set (can be customized for tests).
	return nil, nil
}
