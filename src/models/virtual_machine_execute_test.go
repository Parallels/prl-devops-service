package models

import (
	"encoding/json"
	"strings"
	"testing"
)

// The execute request mirrors `prlctl exec`: user and password are the only
// credential knobs, and a password is meaningless without a user.
func TestVirtualMachineExecuteValidate_PasswordRequiresUser(t *testing.T) {
	cases := []struct {
		name    string
		request VirtualMachineExecuteCommandRequest
		wantErr bool
	}{
		{"no credentials", VirtualMachineExecuteCommandRequest{Command: "id -un"}, false},
		{"user only", VirtualMachineExecuteCommandRequest{Command: "id -un", User: "build"}, false},
		{"user and password", VirtualMachineExecuteCommandRequest{Command: "id -un", User: "build", Password: "s3cr3t"}, false},
		{"password without user", VirtualMachineExecuteCommandRequest{Command: "id -un", Password: "s3cr3t"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected validation error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no validation error, got %v", err)
			}
			if tc.wantErr && !strings.Contains(err.Error(), "password requires user") {
				t.Fatalf("expected password-requires-user error, got %v", err)
			}
		})
	}
}

// Old callers (e.g. the Jenkins plugin) keep sending the removed use_sudo and
// use_ssh keys; the request body mapper is json.Unmarshal, which ignores
// unknown fields, so those payloads must keep deserializing.
func TestVirtualMachineExecute_LegacyPayloadStillDeserializes(t *testing.T) {
	legacy := []byte(`{"command":"id -un","user":"build","use_sudo":true,"use_ssh":false}`)
	var request VirtualMachineExecuteCommandRequest
	if err := json.Unmarshal(legacy, &request); err != nil {
		t.Fatalf("legacy payload must still deserialize, got %v", err)
	}
	if request.Command != "id -un" || request.User != "build" {
		t.Fatalf("known fields must map, got %+v", request)
	}
	if request.Password != "" {
		t.Fatalf("password must stay empty, got %q", request.Password)
	}
}
