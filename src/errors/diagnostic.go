// Package errors provides a simple error and warning accumulator with path tracking.
// It allows for tracking execution paths and accumulating errors/warnings from nested operations.
package errors

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Diagnostics represents a collection of errors, warnings, and execution path
type Diagnostics struct {
	Operation string                 `json:"operation"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Path      []PathEntry            `json:"path"`
	Errors    []Error                `json:"errors"`
	Warnings  []Warning              `json:"warnings"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	mu        sync.RWMutex
}

// PathEntry represents a step in the execution path
type PathEntry struct {
	Operation    string                 `json:"operation"`
	Component    string                 `json:"component"`
	Timestamp    time.Time              `json:"timestamp"`
	Duration     time.Duration          `json:"duration"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	LineNumber   int                    `json:"line_number"`
	FileName     string                 `json:"file_name"`
	FunctionName string                 `json:"function_name"`
}

// Error represents an error with context
type Error struct {
	Code      string                 `json:"code,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Component string                 `json:"component,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Warning represents a warning with context
type Warning struct {
	Code      string                 `json:"code,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Component string                 `json:"component,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// New creates a new Diagnostics instance
func NewDiagnostics(operation string) *Diagnostics {
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	return &Diagnostics{
		Operation: operation,
		StartTime: time.Now(),
		Path: []PathEntry{
			{
				Operation:    operation,
				Component:    "diagnostics",
				Timestamp:    time.Now(),
				LineNumber:   line,
				FileName:     file,
				FunctionName: fn.Name(),
			},
		},
		Errors:   make([]Error, 0),
		Warnings: make([]Warning, 0),
		Metadata: make(map[string]interface{}),
	}
}

// AddPathEntry adds a new entry to the execution path
func (d *Diagnostics) AddPathEntry(operation, component string, metadata ...map[string]interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	entry := PathEntry{
		Operation:    operation,
		Component:    component,
		Timestamp:    time.Now(),
		LineNumber:   line,
		FileName:     file,
		FunctionName: fn.Name(),
	}

	if len(metadata) > 0 {
		entry.Metadata = metadata[0]
	} else {
		entry.Metadata = make(map[string]interface{})
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Calculate duration from previous entry
	if len(d.Path) > 0 {
		entry.Duration = entry.Timestamp.Sub(d.Path[len(d.Path)-1].Timestamp)
	}

	d.Path = append(d.Path, entry)
}

// AddError adds an error and logs it immediately
func (d *Diagnostics) AddError(code, message, component string, metadata ...map[string]interface{}) {
	error := Error{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Component: component,
	}

	if len(metadata) > 0 {
		error.Metadata = metadata[0]
	}

	d.mu.Lock()
	d.Errors = append(d.Errors, error)
	d.mu.Unlock()

}

// AddWarning adds a warning and logs it immediately
func (d *Diagnostics) AddWarning(code, message, component string, metadata ...map[string]interface{}) {
	warning := Warning{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Component: component,
	}

	if len(metadata) > 0 {
		warning.Metadata = metadata[0]
	}

	d.mu.Lock()
	d.Warnings = append(d.Warnings, warning)
	d.mu.Unlock()
}

// AddMetadata adds metadata to the diagnostics
func (d *Diagnostics) AddMetadata(key string, value interface{}) {
	d.mu.Lock()
	d.Metadata[key] = value
	d.mu.Unlock()
}

// GetAllMetadata returns a copy of all metadata
func (d *Diagnostics) GetAllMetadata() map[string]interface{} {
	if d == nil {
		return make(map[string]interface{})
	}
	d.mu.RLock()
	defer d.mu.RUnlock()

	metadata := make(map[string]interface{})
	for k, v := range d.Metadata {
		metadata[k] = v
	}
	return metadata
}

// Complete marks the diagnostics as complete
func (d *Diagnostics) Complete() {
	d.mu.Lock()
	now := time.Now()
	d.EndTime = &now
	d.mu.Unlock()
}

// HasErrors returns true if there are any errors
func (d *Diagnostics) HasErrors() bool {
	if d == nil {
		return false
	}
	d.mu.RLock()
	defer d.mu.RUnlock()

	return len(d.Errors) > 0
}

// HasWarnings returns true if there are any warnings
func (d *Diagnostics) HasWarnings() bool {
	if d == nil {
		return false
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.Warnings) > 0
}

// GetErrorCount returns the number of errors
func (d *Diagnostics) GetErrorCount() int {
	if d == nil {
		return 0
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.Errors)
}

// GetWarningCount returns the number of warnings
func (d *Diagnostics) GetWarningCount() int {
	if d == nil {
		return 0
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.Warnings)
}

// GetErrors returns a copy of all errors
func (d *Diagnostics) GetErrors() []Error {
	if d == nil {
		return []Error{}
	}
	d.mu.RLock()
	defer d.mu.RUnlock()

	errors := make([]Error, len(d.Errors))
	copy(errors, d.Errors)
	return errors
}

// GetWarnings returns a copy of all warnings
func (d *Diagnostics) GetWarnings() []Warning {
	if d == nil {
		return []Warning{}
	}
	d.mu.RLock()
	defer d.mu.RUnlock()

	warnings := make([]Warning, len(d.Warnings))
	copy(warnings, d.Warnings)
	return warnings
}

// GetPath returns a copy of the execution path
func (d *Diagnostics) GetPath() []PathEntry {
	if d == nil {
		return []PathEntry{}
	}
	d.mu.RLock()
	defer d.mu.RUnlock()

	path := make([]PathEntry, len(d.Path))
	copy(path, d.Path)
	return path
}

// GetDuration returns the total duration
func (d *Diagnostics) GetDuration() time.Duration {
	if d == nil {
		return 0
	}
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.EndTime != nil {
		return d.EndTime.Sub(d.StartTime)
	}
	return time.Since(d.StartTime)
}

// Append merges another diagnostics into this one
func (d *Diagnostics) Append(other *Diagnostics) {
	if other == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Merge errors
	d.Errors = append(d.Errors, other.GetErrors()...)

	// Merge warnings
	d.Warnings = append(d.Warnings, other.GetWarnings()...)

	// Merge metadata
	for key, value := range other.Metadata {
		d.Metadata[key] = value
	}

	// Merge path entries (append to the end)
	d.Path = append(d.Path, other.GetPath()...)
}

// Print prints a friendly summary of the diagnostics
func (d *Diagnostics) Print() {
	d.mu.RLock()
	defer d.mu.RUnlock()

	fmt.Printf("\n=== Diagnostics Summary ===\n")
	fmt.Printf("Operation: %s\n", d.Operation)
	fmt.Printf("Duration: %v\n", d.GetDuration())
	fmt.Printf("Errors: %d\n", len(d.Errors))
	fmt.Printf("Warnings: %d\n", len(d.Warnings))
	fmt.Printf("Path Entries: %d\n", len(d.Path))

	if len(d.Errors) > 0 {
		fmt.Printf("\n=== Errors ===\n")
		for i, err := range d.Errors {
			fmt.Printf("%d. [%s] %s (Component: %s)\n",
				i+1, err.Code, err.Message, err.Component)
		}
	}

	if len(d.Warnings) > 0 {
		fmt.Printf("\n=== Warnings ===\n")
		for i, warning := range d.Warnings {
			fmt.Printf("%d. [%s] %s (Component: %s)\n",
				i+1, warning.Code, warning.Message, warning.Component)
		}
	}

	if len(d.Path) > 0 {
		fmt.Printf("\n=== Execution Path ===\n")
		for i, entry := range d.Path {
			fmt.Printf("%d. %s (%s) - %s:%d\n",
				i+1, entry.Operation, entry.Component, entry.FileName, entry.LineNumber)
		}
	}

	fmt.Println()
}

// GetSummary returns a string summary of the diagnostics
func (d *Diagnostics) GetSummary() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Operation: %s\n", d.Operation))
	summary.WriteString(fmt.Sprintf("Duration: %v\n", d.GetDuration()))
	summary.WriteString(fmt.Sprintf("Errors: %d\n", len(d.Errors)))
	summary.WriteString(fmt.Sprintf("Warnings: %d\n", len(d.Warnings)))

	if len(d.Errors) > 0 {
		summary.WriteString("\nErrors:\n")
		for i, err := range d.Errors {
			summary.WriteString(fmt.Sprintf("  %d. [%s] %s\n", i+1, err.Code, err.Message))
		}
	}

	if len(d.Warnings) > 0 {
		summary.WriteString("\nWarnings:\n")
		for i, warning := range d.Warnings {
			summary.WriteString(fmt.Sprintf("  %d. [%s] %s\n", i+1, warning.Code, warning.Message))
		}
	}

	if len(d.Path) > 0 {
		summary.WriteString("\nExecution Path:\n")
		for i, entry := range d.Path {
			summary.WriteString(fmt.Sprintf("  %d. %s (%s) - %s:%d\n",
				i+1, entry.Operation, entry.Component, entry.FileName, entry.LineNumber))
		}
	}

	return summary.String()
}
