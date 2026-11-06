package orchestrator

import (
	"testing"

	"github.com/Parallels/prl-devops-service/data/models"
)

func TestMergePreservedIp(t *testing.T) {
	stored := models.VirtualMachine{
		ID:                "vm-1",
		State:             "running",
		InternalIpAddress: "10.10.10.5",
		IpStatus:          "resolved",
	}

	t.Run("incoming-silence-keeps-stored-ip", func(t *testing.T) {
		incoming := models.VirtualMachine{ID: "vm-1", State: "running"}
		merged := mergePreservedIp(stored, incoming)
		if merged.InternalIpAddress != "10.10.10.5" || merged.IpStatus != "resolved" {
			t.Errorf("expected stored IP/status preserved, got %q/%q", merged.InternalIpAddress, merged.IpStatus)
		}
	})

	t.Run("incoming-ip-wins", func(t *testing.T) {
		incoming := models.VirtualMachine{ID: "vm-1", State: "running", InternalIpAddress: "10.10.10.9", IpStatus: "resolved"}
		merged := mergePreservedIp(stored, incoming)
		if merged.InternalIpAddress != "10.10.10.9" {
			t.Errorf("expected live IP to win, got %q", merged.InternalIpAddress)
		}
	})

	t.Run("incoming-stopped-drops-stored-ip", func(t *testing.T) {
		incoming := models.VirtualMachine{ID: "vm-1", State: "stopped"}
		merged := mergePreservedIp(stored, incoming)
		if merged.InternalIpAddress != "" {
			t.Errorf("expected no IP when VM stopped, got %q", merged.InternalIpAddress)
		}
	})

	t.Run("failed-status-carried-through", func(t *testing.T) {
		storedFailed := stored
		storedFailed.InternalIpAddress = ""
		storedFailed.IpStatus = "failed"
		storedFailed.IpStatusDetail = "no IP address obtained after 8 attempts within 4m40s"
		incoming := models.VirtualMachine{ID: "vm-1", State: "running"}
		merged := mergePreservedIp(storedFailed, incoming)
		if merged.IpStatus != "failed" || merged.IpStatusDetail == "" {
			t.Errorf("expected failed status carried through, got %q/%q", merged.IpStatus, merged.IpStatusDetail)
		}
	})

	t.Run("incoming-status-wins", func(t *testing.T) {
		incoming := models.VirtualMachine{ID: "vm-1", State: "running", IpStatus: "resolving"}
		merged := mergePreservedIp(stored, incoming)
		if merged.IpStatus != "resolving" {
			t.Errorf("expected incoming status to win, got %q", merged.IpStatus)
		}
	})
}
