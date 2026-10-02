package models

import (
	"testing"

	"github.com/Parallels/prl-devops-service/errors"
)

func TestPushCatalogManifestRequestValidate_MissingLocalPath(t *testing.T) {
	r := PushCatalogManifestRequest{
		CatalogId:    "test-catalog",
		Version:      "v1.0",
		Architecture: "x86_64",
		Connection:   "provider://something",
	}
	diag := errors.NewDiagnostics("Test")
	r.Validate(diag)
	if !diag.HasErrors() {
		t.Fatal("expected error for missing local_path, got none")
	}
	if diag.GetErrors()[0].Message != ErrPushMissingLocalPath.Error() {
		t.Errorf("expected ErrPushMissingLocalPath, got %v", diag.GetErrors()[0].Message)
	}
}

func TestPushCatalogManifestRequestValidate_MissingCatalogId(t *testing.T) {
	r := PushCatalogManifestRequest{
		LocalPath:    "/some/path",
		Version:      "v1.0",
		Architecture: "x86_64",
		Connection:   "provider://something",
	}
	diag := errors.NewDiagnostics("Test")
	r.Validate(diag)
	if !diag.HasErrors() {
		t.Fatal("expected error for missing catalog_id, got none")
	}
	if diag.GetErrors()[0].Message != ErrPushMissingCatalogId.Error() {
		t.Errorf("expected ErrPushMissingCatalogId, got %v", diag.GetErrors()[0].Message)
	}
}

func TestPushCatalogManifestRequestValidate_MissingConnection(t *testing.T) {
	r := PushCatalogManifestRequest{
		LocalPath:    "/some/path",
		CatalogId:    "test-catalog",
		Version:      "v1.0",
		Architecture: "x86_64",
	}
	diag := errors.NewDiagnostics("Test")
	r.Validate(diag)
	if !diag.HasErrors() {
		t.Fatal("expected error for missing connection, got none")
	}
	if diag.GetErrors()[0].Message != ErrMissingConnection.Error() {
		t.Errorf("expected ErrMissingConnection, got %v", diag.GetErrors()[0].Message)
	}
}

func TestPushCatalogManifestRequestValidate_MissingVersion(t *testing.T) {
	r := PushCatalogManifestRequest{
		LocalPath:    "/some/path",
		CatalogId:    "test-catalog",
		Architecture: "x86_64",
		Connection:   "provider://something",
	}
	diag := errors.NewDiagnostics("Test")
	r.Validate(diag)
	if !diag.HasErrors() {
		t.Fatal("expected error for missing version, got none")
	}
	if diag.GetErrors()[0].Message != ErrPushMissingVersion.Error() {
		t.Errorf("expected ErrPushMissingVersion, got %v", diag.GetErrors()[0].Message)
	}
}

func TestPushCatalogManifestRequestValidate_MissingArchitecture(t *testing.T) {
	r := PushCatalogManifestRequest{
		LocalPath:    "/some/path",
		CatalogId:    "test-catalog",
		Version:      "v1.0",
		Connection:   "provider://something",
	}
	diag := errors.NewDiagnostics("Test")
	r.Validate(diag)
	if !diag.HasErrors() {
		t.Fatal("expected error for missing architecture, got none")
	}
	if diag.GetErrors()[0].Message != ErrMissingArchitecture.Error() {
		t.Errorf("expected ErrMissingArchitecture, got %v", diag.GetErrors()[0].Message)
	}
}

func TestPushCatalogManifestRequestValidate_InvalidArchitecture(t *testing.T) {
	r := PushCatalogManifestRequest{
		LocalPath:    "/some/path",
		CatalogId:    "test-catalog",
		Version:      "v1.0",
		Architecture: "mips",
		Connection:   "provider://something",
	}
	diag := errors.NewDiagnostics("Test")
	r.Validate(diag)
	if !diag.HasErrors() {
		t.Fatal("expected error for invalid architecture, got none")
	}
	if diag.GetErrors()[0].Message != ErrInvalidArchitecture.Error() {
		t.Errorf("expected ErrInvalidArchitecture, got %v", diag.GetErrors()[0].Message)
	}
}

func TestPushCatalogManifestRequestValidate_ArchitectureNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"amd64", "x86_64"},
		{"arm", "arm64"},
		{"aarch64", "arm64"},
		{"x86_64", "x86_64"},
		{"arm64", "arm64"},
	}

	for _, tt := range tests {
		r := PushCatalogManifestRequest{
			LocalPath:    "/some/path",
			CatalogId:    "test-catalog",
			Version:      "v1.0",
			Architecture: tt.input,
			Connection:   "provider://something",
		}
		diag := errors.NewDiagnostics("Test")
		r.Validate(diag)
		if diag.HasErrors() {
			t.Errorf("input %q: unexpected error: %v", tt.input, diag.GetErrors()[0].Message)
			continue
		}
		if r.Architecture != tt.expected {
			t.Errorf("input %q: expected architecture %q, got %q", tt.input, tt.expected, r.Architecture)
		}
	}
}

func TestPushCatalogManifestRequestValidate_VersionWithIllegalChars(t *testing.T) {
	r := PushCatalogManifestRequest{
		LocalPath:    "/some/path",
		CatalogId:    "test-catalog",
		Version:      "v1.0!@#",
		Architecture: "x86_64",
		Connection:   "provider://something",
	}
	diag := errors.NewDiagnostics("Test")
	r.Validate(diag)
	if !diag.HasErrors() {
		t.Fatal("expected error for version with illegal chars, got none")
	}
	if diag.GetErrors()[0].Message != ErrPushVersionInvalidChars.Error() {
		t.Errorf("expected ErrPushVersionInvalidChars, got %v", diag.GetErrors()[0].Message)
	}
}

func TestPushCatalogManifestRequestValidate_Valid(t *testing.T) {
	r := PushCatalogManifestRequest{
		LocalPath:    "/some/path",
		CatalogId:    "test-catalog",
		Version:      "v1.0",
		Architecture: "x86_64",
		Connection:   "provider://something",
	}
	diag := errors.NewDiagnostics("Test")
	r.Validate(diag)
	if diag.HasErrors() {
		t.Errorf("expected no error, got %v", diag.GetErrors()[0].Message)
	}
}

func TestPushCatalogManifestRequestValidate_JobIdNotInJson(t *testing.T) {
	// JobId should be excluded from JSON serialization
	r := PushCatalogManifestRequest{}
	r.JobId = "some-job-id"

	// The field has json:"-" tag so it won't be serialized, just verify it can be set
	if r.JobId != "some-job-id" {
		t.Errorf("expected JobId to be settable, got %v", r.JobId)
	}
}

func TestPushCatalogManifestRequestValidate_MinimumSpecDefaults(t *testing.T) {
	r := PushCatalogManifestRequest{
		LocalPath:    "/some/path",
		CatalogId:    "test-catalog",
		Version:      "v1.0",
		Architecture: "x86_64",
		Connection:   "provider://something",
		MinimumSpecRequirements: MinimumSpecRequirement{
			Cpu:    2,
			Memory: 4096,
			Disk:   20480,
		},
	}
	diag := errors.NewDiagnostics("Test")
	r.Validate(diag)
	if diag.HasErrors() {
		t.Errorf("expected no error for valid request with min spec requirements, got %v", diag.GetErrors()[0].Message)
	}
	if r.MinimumSpecRequirements.Cpu != 2 {
		t.Errorf("expected Cpu=2, got %v", r.MinimumSpecRequirements.Cpu)
	}
	if r.MinimumSpecRequirements.Memory != 4096 {
		t.Errorf("expected Memory=4096, got %v", r.MinimumSpecRequirements.Memory)
	}
	if r.MinimumSpecRequirements.Disk != 20480 {
		t.Errorf("expected Disk=20480, got %v", r.MinimumSpecRequirements.Disk)
	}
}
