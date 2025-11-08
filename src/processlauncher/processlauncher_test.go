package processlauncher

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMockProcessLauncher(t *testing.T) {
	// Create a temporary pipe to simulate stdout
	r, w, _ := os.Pipe()
	w.WriteString("test output")
	w.Close()

	mock := &MockProcessLauncher{
		LaunchFunc: func(cmd *exec.Cmd) (*os.File, error) {
			// Simulate successful launch by returning a file with some test data.
			return r, nil
		},
	}

	cmd := exec.Command("echo", "hello")
	file, err := mock.Start(cmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer file.Close()

	// Read from the returned file to verify content
	reader := bufio.NewReader(file)
	output, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		t.Fatalf("Expected no error reading from file, got %v", err)
	}

	if strings.TrimSpace(output) != "test output" {
		t.Fatalf("Expected 'test output', got '%s'", output)
	}
}

func TestRealProcessLauncher(t *testing.T) {
	real := &RealProcessLauncher{}
	cmd := exec.Command("echo", "hello")
	file, err := real.Start(cmd)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer file.Close()

	// Read from the returned file to verify content
	reader := bufio.NewReader(file)
	output, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		t.Fatalf("Expected no error reading from file, got %v", err)
	}

	if strings.TrimSpace(output) != "hello" {
		t.Fatalf("Expected 'hello', got '%s'", output)
	}
}
