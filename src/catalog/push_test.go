package catalog

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/constants"
)

func TestPushWithExistingJob_SetsJobId(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	svc := NewManifestService(ctx)

	r := &models.PushCatalogManifestRequest{
		LocalPath:    "/nonexistent/path",
		CatalogId:    "test-catalog",
		Version:      "v1.0",
		Architecture: "x86_64",
		Connection:   "invalid://connection",
	}

	jobId := "test-job-123"
	// PushWithExistingJob should set the JobId on the request before calling Push
	// We can observe this because the request is passed by pointer
	_ = svc.PushWithExistingJob(jobId, r)

	if r.JobId != jobId {
		t.Errorf("expected JobId to be %q, got %q", jobId, r.JobId)
	}
}

func TestPushWithExistingJob_EmptyJobId(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	svc := NewManifestService(ctx)

	r := &models.PushCatalogManifestRequest{
		LocalPath:    "/nonexistent/path",
		CatalogId:    "test-catalog",
		Version:      "v1.0",
		Architecture: "x86_64",
		Connection:   "invalid://connection",
	}

	result := svc.PushWithExistingJob("", r)

	// With no matching provider the push should return an error
	if !result.HasErrors() {
		t.Error("expected errors for invalid connection, got none")
	}
}

func TestPush_NoMatchingProvider(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	svc := NewManifestService(ctx)

	r := &models.PushCatalogManifestRequest{
		LocalPath:    "/nonexistent/path",
		CatalogId:    "test-catalog",
		Version:      "v1.0",
		Architecture: "x86_64",
		Connection:   "unknown-provider://invalid-connection-string-that-matches-nothing",
	}

	result := svc.Push(r)

	if !result.HasErrors() {
		t.Error("expected errors when no provider matches, got none")
	}
}

func TestPush_NilRequest(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	svc := NewManifestService(ctx)

	// Push with a zero-value request (all empty fields) should fail
	r := &models.PushCatalogManifestRequest{}
	result := svc.Push(r)

	if !result.HasErrors() {
		t.Error("expected errors for empty push request, got none")
	}
}

func TestAsyncPush_NoJobManager(t *testing.T) {
	// When the job manager is not available, AsyncPush should return gracefully
	ctx := basecontext.NewRootBaseContext()
	svc := NewManifestService(ctx)

	r := &models.PushCatalogManifestRequest{
		LocalPath:    "/nonexistent/path",
		CatalogId:    "test-catalog",
		Version:      "v1.0",
		Architecture: "x86_64",
		Connection:   "invalid://connection",
	}

	// This should not panic even when the job manager is unavailable
	// (global job manager is nil in test context)
	done := make(chan struct{})
	go func() {
		defer close(done)
		svc.AsyncPush("test-job-id", r)
	}()
	<-done
}

func TestPushActionConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
	}{
		{"ActionPushValidateStage", constants.ActionPushValidateStage},
		{"ActionPushCompressStage", constants.ActionPushCompressStage},
		{"ActionPushCheckRemoteStage", constants.ActionPushCheckRemoteStage},
		{"ActionPushUploadPackStage", constants.ActionPushUploadPackStage},
		{"ActionPushUploadMetaStage", constants.ActionPushUploadMetaStage},
		{"ActionPushRegisterStage", constants.ActionPushRegisterStage},
	}

	for _, tt := range tests {
		if tt.constant == "" {
			t.Errorf("constant %s should not be empty", tt.name)
		}
	}
}

func TestApplyMinimumSpecRequirements_AllFields(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	svc := NewManifestService(ctx)

	r := &models.PushCatalogManifestRequest{
		MinimumSpecRequirements: models.MinimumSpecRequirement{
			Cpu:    4,
			Memory: 8192,
			Disk:   51200,
		},
	}
	manifest := models.NewVirtualMachineCatalogManifest()

	svc.applyMinimumSpecRequirements(r, manifest)

	if manifest.MinimumSpecRequirements == nil {
		t.Fatal("expected MinimumSpecRequirements to be set, got nil")
	}
	if manifest.MinimumSpecRequirements.Cpu != 4 {
		t.Errorf("expected Cpu=4, got %v", manifest.MinimumSpecRequirements.Cpu)
	}
	if manifest.MinimumSpecRequirements.Memory != 8192 {
		t.Errorf("expected Memory=8192, got %v", manifest.MinimumSpecRequirements.Memory)
	}
	if manifest.MinimumSpecRequirements.Disk != 51200 {
		t.Errorf("expected Disk=51200, got %v", manifest.MinimumSpecRequirements.Disk)
	}
}

func TestApplyMinimumSpecRequirements_ZeroValues(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	svc := NewManifestService(ctx)

	r := &models.PushCatalogManifestRequest{
		MinimumSpecRequirements: models.MinimumSpecRequirement{},
	}
	manifest := models.NewVirtualMachineCatalogManifest()

	svc.applyMinimumSpecRequirements(r, manifest)

	// When all requirements are zero, MinimumSpecRequirements should remain nil
	if manifest.MinimumSpecRequirements != nil {
		t.Errorf("expected MinimumSpecRequirements to remain nil for zero values")
	}
}

func TestApplyMinimumSpecRequirements_PartialFields(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	svc := NewManifestService(ctx)

	r := &models.PushCatalogManifestRequest{
		MinimumSpecRequirements: models.MinimumSpecRequirement{
			Cpu: 2,
		},
	}
	manifest := models.NewVirtualMachineCatalogManifest()

	svc.applyMinimumSpecRequirements(r, manifest)

	if manifest.MinimumSpecRequirements == nil {
		t.Fatal("expected MinimumSpecRequirements to be set")
	}
	if manifest.MinimumSpecRequirements.Cpu != 2 {
		t.Errorf("expected Cpu=2, got %v", manifest.MinimumSpecRequirements.Cpu)
	}
	if manifest.MinimumSpecRequirements.Memory != 0 {
		t.Errorf("expected Memory=0 (unset), got %v", manifest.MinimumSpecRequirements.Memory)
	}
}
