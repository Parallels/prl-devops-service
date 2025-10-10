package errors

import (
	"strings"
	"testing"
	"time"
)

func TestNewDiagnostics(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	if diag.Operation != "test-operation" {
		t.Errorf("Expected operation 'test-operation', got '%s'", diag.Operation)
	}

	if len(diag.Path) != 1 {
		t.Errorf("Expected 1 path entry, got %d", len(diag.Path))
	}

	if diag.Path[0].Operation != "test-operation" {
		t.Errorf("Expected path entry operation 'test-operation', got '%s'", diag.Path[0].Operation)
	}

	if diag.Path[0].Component != "diagnostics" {
		t.Errorf("Expected component 'diagnostics', got '%s'", diag.Path[0].Component)
	}

	if diag.Path[0].LineNumber <= 0 {
		t.Error("Expected line number to be set")
	}

	if diag.Path[0].FileName == "" {
		t.Error("Expected file name to be set")
	}

	if diag.Path[0].FunctionName == "" {
		t.Error("Expected function name to be set")
	}
}

func TestDiagnostics_AddPathEntry(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	diag.AddPathEntry("step1", "component1", map[string]interface{}{"key": "value"})
	diag.AddPathEntry("step2", "component2")

	if len(diag.Path) != 3 { // 1 initial + 2 added
		t.Errorf("Expected 3 path entries, got %d", len(diag.Path))
	}

	if diag.Path[1].Operation != "step1" {
		t.Errorf("Expected step1 operation, got '%s'", diag.Path[1].Operation)
	}

	if diag.Path[1].Component != "component1" {
		t.Errorf("Expected component1, got '%s'", diag.Path[1].Component)
	}

	if diag.Path[1].Metadata["key"] != "value" {
		t.Errorf("Expected metadata key=value, got %v", diag.Path[1].Metadata)
	}

	// Test duration calculation
	if diag.Path[1].Duration <= 0 {
		t.Error("Expected duration to be calculated")
	}

	if diag.Path[2].Duration <= 0 {
		t.Error("Expected duration to be calculated")
	}
}

func TestDiagnostics_AddPathEntry_NoMetadata(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	diag.AddPathEntry("step1", "component1")

	if len(diag.Path) != 2 {
		t.Errorf("Expected 2 path entries, got %d", len(diag.Path))
	}

	if diag.Path[1].Metadata == nil {
		t.Error("Expected metadata map to be initialized")
	}
}

func TestDiagnostics_AddError(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	diag.AddError("TEST_ERROR", "Test error message", "test-component")

	if !diag.HasErrors() {
		t.Error("Expected diagnostics to have errors")
	}

	if diag.GetErrorCount() != 1 {
		t.Errorf("Expected 1 error, got %d", diag.GetErrorCount())
	}

	errors := diag.GetErrors()
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	if errors[0].Code != "TEST_ERROR" {
		t.Errorf("Expected error code 'TEST_ERROR', got '%s'", errors[0].Code)
	}

	if errors[0].Message != "Test error message" {
		t.Errorf("Expected error message 'Test error message', got '%s'", errors[0].Message)
	}

	if errors[0].Component != "test-component" {
		t.Errorf("Expected component 'test-component', got '%s'", errors[0].Component)
	}

	if errors[0].Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestDiagnostics_AddError_WithMetadata(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	metadata := map[string]interface{}{"key": "value", "number": 42}
	diag.AddError("TEST_ERROR", "Test error message", "test-component", metadata)

	errors := diag.GetErrors()
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}

	if errors[0].Metadata["key"] != "value" {
		t.Errorf("Expected metadata key=value, got %v", errors[0].Metadata["key"])
	}

	if errors[0].Metadata["number"] != 42 {
		t.Errorf("Expected metadata number=42, got %v", errors[0].Metadata["number"])
	}
}

func TestDiagnostics_AddWarning(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	diag.AddWarning("TEST_WARNING", "Test warning message", "test-component")

	if !diag.HasWarnings() {
		t.Error("Expected diagnostics to have warnings")
	}

	if diag.GetWarningCount() != 1 {
		t.Errorf("Expected 1 warning, got %d", diag.GetWarningCount())
	}

	warnings := diag.GetWarnings()
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	if warnings[0].Code != "TEST_WARNING" {
		t.Errorf("Expected warning code 'TEST_WARNING', got '%s'", warnings[0].Code)
	}

	if warnings[0].Message != "Test warning message" {
		t.Errorf("Expected warning message 'Test warning message', got '%s'", warnings[0].Message)
	}

	if warnings[0].Component != "test-component" {
		t.Errorf("Expected component 'test-component', got '%s'", warnings[0].Component)
	}

	if warnings[0].Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestDiagnostics_AddWarning_WithMetadata(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	metadata := map[string]interface{}{"key": "value", "number": 42}
	diag.AddWarning("TEST_WARNING", "Test warning message", "test-component", metadata)

	warnings := diag.GetWarnings()
	if len(warnings) != 1 {
		t.Fatalf("Expected 1 warning, got %d", len(warnings))
	}

	if warnings[0].Metadata["key"] != "value" {
		t.Errorf("Expected metadata key=value, got %v", warnings[0].Metadata["key"])
	}

	if warnings[0].Metadata["number"] != 42 {
		t.Errorf("Expected metadata number=42, got %v", warnings[0].Metadata["number"])
	}
}

func TestDiagnostics_AddMetadata(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	diag.AddMetadata("key1", "value1")
	diag.AddMetadata("key2", 123)
	diag.AddMetadata("key3", true)
	diag.AddMetadata("key4", map[string]string{"nested": "value"})

	diag.mu.RLock()
	metadata := diag.Metadata
	diag.mu.RUnlock()

	if metadata["key1"] != "value1" {
		t.Errorf("Expected metadata key1=value1, got %v", metadata["key1"])
	}

	if metadata["key2"] != 123 {
		t.Errorf("Expected metadata key2=123, got %v", metadata["key2"])
	}

	if metadata["key3"] != true {
		t.Errorf("Expected metadata key3=true, got %v", metadata["key3"])
	}

	if metadata["key4"].(map[string]string)["nested"] != "value" {
		t.Errorf("Expected metadata key4.nested=value, got %v", metadata["key4"])
	}
}

func TestDiagnostics_Complete(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	// Wait a bit to ensure different timestamps
	time.Sleep(1 * time.Millisecond)
	diag.Complete()

	if diag.EndTime == nil {
		t.Error("Expected EndTime to be set")
	}

	duration := diag.GetDuration()
	if duration <= 0 {
		t.Errorf("Expected positive duration, got %v", duration)
	}

	// Test that duration is calculated correctly
	expectedDuration := diag.EndTime.Sub(diag.StartTime)
	if duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, duration)
	}
}

func TestDiagnostics_GetDuration_NotComplete(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	// Wait a bit
	time.Sleep(1 * time.Millisecond)

	duration := diag.GetDuration()
	if duration <= 0 {
		t.Errorf("Expected positive duration, got %v", duration)
	}

	// Should be approximately the time since creation
	if duration < time.Millisecond {
		t.Errorf("Expected duration to be at least 1ms, got %v", duration)
	}
}

func TestDiagnostics_GetSummary(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	diag.AddPathEntry("step1", "component1")
	diag.AddError("ERROR1", "Error 1", "component1")
	diag.AddWarning("WARNING1", "Warning 1", "component1")

	summary := diag.GetSummary()

	if !strings.Contains(summary, "Operation: test-operation") {
		t.Error("Expected summary to contain operation name")
	}

	if !strings.Contains(summary, "Errors: 1") {
		t.Error("Expected summary to contain error count")
	}

	if !strings.Contains(summary, "Warnings: 1") {
		t.Error("Expected summary to contain warning count")
	}

	if !strings.Contains(summary, "ERROR1") {
		t.Error("Expected summary to contain error code")
	}

	if !strings.Contains(summary, "WARNING1") {
		t.Error("Expected summary to contain warning code")
	}
}

func TestDiagnostics_GetSummary_NoErrorsOrWarnings(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	summary := diag.GetSummary()

	if !strings.Contains(summary, "Operation: test-operation") {
		t.Error("Expected summary to contain operation name")
	}

	if !strings.Contains(summary, "Errors: 0") {
		t.Error("Expected summary to contain error count")
	}

	if !strings.Contains(summary, "Warnings: 0") {
		t.Error("Expected summary to contain warning count")
	}

	if strings.Contains(summary, "Errors:\n") {
		t.Error("Expected summary to not contain errors section")
	}

	if strings.Contains(summary, "Warnings:\n") {
		t.Error("Expected summary to not contain warnings section")
	}
}

func TestDiagnostics_Append(t *testing.T) {
	parent := NewDiagnostics("parent-operation")
	child := NewDiagnostics("child-operation")

	child.AddError("CHILD_ERROR", "Child error", "child-component")
	child.AddWarning("CHILD_WARNING", "Child warning", "child-component")
	child.AddPathEntry("child_step", "child-component")
	child.AddMetadata("child_key", "child_value")

	parent.Append(child)

	if parent.GetErrorCount() != 1 {
		t.Errorf("Expected 1 error after append, got %d", parent.GetErrorCount())
	}
	if parent.GetWarningCount() != 1 {
		t.Errorf("Expected 1 warning after append, got %d", parent.GetWarningCount())
	}
	if len(parent.Path) < 2 {
		t.Errorf("Expected appended path entries, got %d", len(parent.Path))
	}

	// Check metadata was merged
	parent.mu.RLock()
	metadata := parent.Metadata
	parent.mu.RUnlock()

	if metadata["child_key"] != "child_value" {
		t.Errorf("Expected metadata to be merged, got %v", metadata["child_key"])
	}
}

func TestDiagnostics_Append_NilChild(t *testing.T) {
	parent := NewDiagnostics("parent-operation")
	parent.AddError("PARENT_ERROR", "Parent error", "parent-component")

	// Should not panic
	parent.Append(nil)

	if parent.GetErrorCount() != 1 {
		t.Errorf("Expected 1 error after append nil, got %d", parent.GetErrorCount())
	}
}

func TestDiagnostics_Append_EmptyChild(t *testing.T) {
	parent := NewDiagnostics("parent-operation")
	child := NewDiagnostics("child-operation")

	parent.AddError("PARENT_ERROR", "Parent error", "parent-component")
	parent.Append(child)

	if parent.GetErrorCount() != 1 {
		t.Errorf("Expected 1 error after append empty child, got %d", parent.GetErrorCount())
	}
}

func TestDiagnostics_Print(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	diag.AddPathEntry("step1", "component1")
	diag.AddError("ERROR1", "Error 1", "component1")
	diag.AddWarning("WARNING1", "Warning 1", "component1")
	diag.AddMetadata("test_key", "test_value")

	// Capture output by redirecting stdout (this is a basic test)
	// In a real scenario, you might want to use a buffer or mock the output
	diag.Print()

	// Just verify it doesn't panic and completes
	if diag.Operation != "test-operation" {
		t.Error("Expected operation to remain unchanged")
	}
}

func TestDiagnostics_Print_Empty(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	// Should not panic
	diag.Print()

	if diag.Operation != "test-operation" {
		t.Error("Expected operation to remain unchanged")
	}
}

func TestDiagnostics_NilPointerHandling(t *testing.T) {
	var diag *Diagnostics

	// All methods should handle nil gracefully
	if diag.HasErrors() {
		t.Error("Expected nil diagnostics to not have errors")
	}

	if diag.HasWarnings() {
		t.Error("Expected nil diagnostics to not have warnings")
	}

	if diag.GetErrorCount() != 0 {
		t.Errorf("Expected nil diagnostics to have 0 errors, got %d", diag.GetErrorCount())
	}

	if diag.GetWarningCount() != 0 {
		t.Errorf("Expected nil diagnostics to have 0 warnings, got %d", diag.GetWarningCount())
	}

	if len(diag.GetErrors()) != 0 {
		t.Errorf("Expected nil diagnostics to return empty errors slice, got %d", len(diag.GetErrors()))
	}

	if len(diag.GetWarnings()) != 0 {
		t.Errorf("Expected nil diagnostics to return empty warnings slice, got %d", len(diag.GetWarnings()))
	}

	if len(diag.GetPath()) != 0 {
		t.Errorf("Expected nil diagnostics to return empty path slice, got %d", len(diag.GetPath()))
	}

	if diag.GetDuration() != 0 {
		t.Errorf("Expected nil diagnostics to return 0 duration, got %v", diag.GetDuration())
	}
}

func TestDiagnostics_ConcurrentAccess(t *testing.T) {
	diag := NewDiagnostics("concurrent-test")

	// Test concurrent access to AddPathEntry
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			diag.AddPathEntry("concurrent-step", "component", map[string]interface{}{"index": index})
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 11 path entries (1 initial + 10 concurrent)
	if len(diag.Path) != 11 {
		t.Errorf("Expected 11 path entries, got %d", len(diag.Path))
	}
}

func TestDiagnostics_ConcurrentErrorsAndWarnings(t *testing.T) {
	diag := NewDiagnostics("concurrent-test")

	// Test concurrent access to AddError and AddWarning
	done := make(chan bool, 20)
	for i := 0; i < 10; i++ {
		go func(index int) {
			diag.AddError("CONCURRENT_ERROR", "Concurrent error", "component", map[string]interface{}{"index": index})
			done <- true
		}(i)
		go func(index int) {
			diag.AddWarning("CONCURRENT_WARNING", "Concurrent warning", "component", map[string]interface{}{"index": index})
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should have 10 errors and 10 warnings
	if diag.GetErrorCount() != 10 {
		t.Errorf("Expected 10 errors, got %d", diag.GetErrorCount())
	}

	if diag.GetWarningCount() != 10 {
		t.Errorf("Expected 10 warnings, got %d", diag.GetWarningCount())
	}
}

func TestDiagnostics_MultipleAppends(t *testing.T) {
	parent := NewDiagnostics("parent-operation")
	child1 := NewDiagnostics("child1-operation")
	child2 := NewDiagnostics("child2-operation")

	child1.AddError("CHILD1_ERROR", "Child 1 error", "child1-component")
	child2.AddWarning("CHILD2_WARNING", "Child 2 warning", "child2-component")

	parent.Append(child1)
	parent.Append(child2)

	if parent.GetErrorCount() != 1 {
		t.Errorf("Expected 1 error after multiple appends, got %d", parent.GetErrorCount())
	}

	if parent.GetWarningCount() != 1 {
		t.Errorf("Expected 1 warning after multiple appends, got %d", parent.GetWarningCount())
	}
}

func TestDiagnostics_PathEntryMetadata(t *testing.T) {
	diag := NewDiagnostics("test-operation")

	metadata1 := map[string]interface{}{"key1": "value1", "number": 42}
	metadata2 := map[string]interface{}{"key2": "value2", "bool": true}

	diag.AddPathEntry("step1", "component1", metadata1)
	diag.AddPathEntry("step2", "component2", metadata2)

	if len(diag.Path) != 3 { // 1 initial + 2 added
		t.Errorf("Expected 3 path entries, got %d", len(diag.Path))
	}

	if diag.Path[1].Metadata["key1"] != "value1" {
		t.Errorf("Expected metadata key1=value1, got %v", diag.Path[1].Metadata["key1"])
	}

	if diag.Path[1].Metadata["number"] != 42 {
		t.Errorf("Expected metadata number=42, got %v", diag.Path[1].Metadata["number"])
	}

	if diag.Path[2].Metadata["key2"] != "value2" {
		t.Errorf("Expected metadata key2=value2, got %v", diag.Path[2].Metadata["key2"])
	}

	if diag.Path[2].Metadata["bool"] != true {
		t.Errorf("Expected metadata bool=true, got %v", diag.Path[2].Metadata["bool"])
	}
}
