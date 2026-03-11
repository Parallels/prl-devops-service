package models

import (
	"testing"
)

func TestPushCatalogManifestRequestValidate_MissingLocalPath(t *testing.T) {
	r := PushCatalogManifestRequest{
		CatalogId:    "test-catalog",
		Version:      "v1.0",
		Architecture: "x86_64",
		Connection:   "provider://something",
	}
	err := r.Validate()
	if err == nil {
		t.Fatal("expected error for missing local_path, got nil")
	}
	if err != ErrPushMissingLocalPath {
		t.Errorf("expected ErrPushMissingLocalPath, got %v", err)
	}
}

func TestPushCatalogManifestRequestValidate_MissingCatalogId(t *testing.T) {
	r := PushCatalogManifestRequest{
		LocalPath:    "/some/path",
		Version:      "v1.0",
		Architecture: "x86_64",
		Connection:   "provider://something",
	}
	err := r.Validate()
	if err == nil {
		t.Fatal("expected error for missing catalog_id, got nil")
	}
	if err != ErrPushMissingCatalogId {
		t.Errorf("expected ErrPushMissingCatalogId, got %v", err)
	}
}

func TestPushCatalogManifestRequestValidate_MissingConnection(t *testing.T) {
	r := PushCatalogManifestRequest{
		LocalPath:    "/some/path",
		CatalogId:    "test-catalog",
		Version:      "v1.0",
		Architecture: "x86_64",
	}
	err := r.Validate()
	if err == nil {
		t.Fatal("expected error for missing connection, got nil")
	}
	if err != ErrMissingConnection {
		t.Errorf("expected ErrMissingConnection, got %v", err)
	}
}

func TestPushCatalogManifestRequestValidate_MissingVersion(t *testing.T) {
	r := PushCatalogManifestRequest{
		LocalPath:    "/some/path",
		CatalogId:    "test-catalog",
		Architecture: "x86_64",
		Connection:   "provider://something",
	}
	err := r.Validate()
	if err == nil {
		t.Fatal("expected error for missing version, got nil")
	}
	if err != ErrPushMissingVersion {
		t.Errorf("expected ErrPushMissingVersion, got %v", err)
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
	err := r.Validate()
	if err == nil {
		t.Fatal("expected error for invalid architecture, got nil")
	}
	if err != ErrInvalidArchitecture {
		t.Errorf("expected ErrInvalidArchitecture, got %v", err)
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
		if err := r.Validate(); err != nil {
			t.Errorf("input %q: unexpected error: %v", tt.input, err)
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
	err := r.Validate()
	if err == nil {
		t.Fatal("expected error for version with illegal chars, got nil")
	}
	if err != ErrPushVersionInvalidChars {
		t.Errorf("expected ErrPushVersionInvalidChars, got %v", err)
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
	if err := r.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
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
	if err := r.Validate(); err != nil {
		t.Errorf("expected no error for valid request with min spec requirements, got %v", err)
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
