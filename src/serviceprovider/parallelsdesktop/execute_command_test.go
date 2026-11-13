package parallelsdesktop

import (
	"reflect"
	"strings"
	"testing"
)

// buildExecArgs mirrors the prlctl exec interface. With no guest credentials
// it must produce exactly the plain form used before the user/password fields
// existed, so existing callers see byte-identical commands.
func TestBuildExecArgs(t *testing.T) {
	cases := []struct {
		name     string
		user     string
		password string
		want     []string
	}{
		{
			name: "no credentials is plain exec",
			want: []string{"exec", "vm-1", "id -un"},
		},
		{
			name: "user maps to --user",
			user: "build",
			want: []string{"exec", "vm-1", "--user", "build", "id -un"},
		},
		{
			name:     "user and password map to prlctl options",
			user:     "build",
			password: "s3cr3t",
			want:     []string{"exec", "vm-1", "--user", "build", "--password", "s3cr3t", "id -un"},
		},
		{
			name:     "password without user is never sent",
			password: "s3cr3t",
			want:     []string{"exec", "vm-1", "id -un"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildExecArgs("vm-1", tc.user, tc.password, "id -un")
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

// Guest credentials must never reach the service logs.
func TestRedactExecArgs(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "separates password value",
			args: []string{"exec", "vm-1", "--user", "build", "--password", "s3cr3t", "id -un"},
			want: "exec vm-1 --user build --password **** id -un",
		},
		{
			name: "masks equals form",
			args: []string{"exec", "vm-1", "--password=s3cr3t", "id -un"},
			want: "exec vm-1 --password=**** id -un",
		},
		{
			name: "no password passes through",
			args: []string{"exec", "vm-1", "id -un"},
			want: "exec vm-1 id -un",
		},
		{
			name: "host-side AsUser prefix keeps working",
			args: []string{"sudo", "-u", "vmowner", "prlctl", "exec", "vm-1", "--user", "build", "--password", "s3cr3t", "id -un"},
			want: "sudo -u vmowner prlctl exec vm-1 --user build --password **** id -un",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := redactExecArgs(tc.args)
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
			if strings.Contains(got, "s3cr3t") {
				t.Fatalf("password leaked into rendered args: %q", got)
			}
		})
	}
}
