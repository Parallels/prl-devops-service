package migrations

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
)

// MockSeedWorker is a mock implementation of SeedWorker for testing
type MockSeedWorker struct {
	Name        string
	Description string
	ShouldFail  bool
	FailOnUp    bool
	FailOnDown  bool
	UpCalled    bool
	DownCalled  bool
}

// NewMockSeedWorker creates a new mock seed worker
func NewMockSeedWorker(name, description string) *MockSeedWorker {
	return &MockSeedWorker{
		Name:        name,
		Description: description,
		ShouldFail:  false,
		FailOnUp:    false,
		FailOnDown:  false,
		UpCalled:    false,
		DownCalled:  false,
	}
}

// GetName returns the name of the mock worker
func (m *MockSeedWorker) GetName() string {
	return m.Name
}

// GetDescription returns the description of the mock worker
func (m *MockSeedWorker) GetDescription() string {
	return m.Description
}

// GetVersion returns the version (always 1 for mock)
func (m *MockSeedWorker) GetVersion() int {
	return 1
}

func (m *MockSeedWorker) GetOrder() int {
	return 1
}

// Up simulates the up migration
func (m *MockSeedWorker) Up(ctx basecontext.BaseContext) *errors.Diagnostics {
	m.UpCalled = true

	diag := errors.NewDiagnostics("mock_up")
	defer diag.Complete()

	if m.ShouldFail || m.FailOnUp {
		diag.AddError("MOCK_UP_FAILED", "Mock up migration failed", "mock_worker", map[string]interface{}{
			"worker_name": m.Name,
		})
		return diag
	}

	diag.AddPathEntry("mock_up_success", "mock_worker", map[string]interface{}{
		"worker_name": m.Name,
	})

	return diag
}

// Down simulates the down migration
func (m *MockSeedWorker) Down(ctx basecontext.BaseContext) *errors.Diagnostics {
	m.DownCalled = true

	diag := errors.NewDiagnostics("mock_down")
	defer diag.Complete()

	if m.ShouldFail || m.FailOnDown {
		diag.AddError("MOCK_DOWN_FAILED", "Mock down migration failed", "mock_worker", map[string]interface{}{
			"worker_name": m.Name,
		})
		return diag
	}

	diag.AddPathEntry("mock_down_success", "mock_worker", map[string]interface{}{
		"worker_name": m.Name,
	})

	return diag
}
