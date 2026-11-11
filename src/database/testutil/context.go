package testutil

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
)

// RecordingApiContext wraps the real BaseContext (embed-and-override, not a
// 14-stub fake — the interface has 14 methods, count per
// basecontext/api_context.go) and records the 5 log levels instead of
// emitting them. #418 task-05 pattern, exported for the parity waves.
type RecordingApiContext struct {
	basecontext.ApiContext
	mu    sync.Mutex
	lines []string
}

// RecordingContext returns the recorder both as concrete type (assertions)
// and as the interface (hand it to production code under test).
func RecordingContext() (*RecordingApiContext, basecontext.ApiContext) {
	inner := basecontext.NewBaseContext()
	inner.DisableLog()
	rec := &RecordingApiContext{ApiContext: inner}
	return rec, rec
}

func (r *RecordingApiContext) record(level, format string, a ...interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lines = append(r.lines, level+": "+fmt.Sprintf(format, a...))
}

func (r *RecordingApiContext) LogInfof(format string, a ...interface{}) {
	r.record("info", format, a...)
}
func (r *RecordingApiContext) LogErrorf(format string, a ...interface{}) {
	r.record("error", format, a...)
}
func (r *RecordingApiContext) LogDebugf(format string, a ...interface{}) {
	r.record("debug", format, a...)
}
func (r *RecordingApiContext) LogWarnf(format string, a ...interface{}) {
	r.record("warn", format, a...)
}
func (r *RecordingApiContext) LogTracef(format string, a ...interface{}) {
	r.record("trace", format, a...)
}

// Lines returns a copy of everything captured, in order.
func (r *RecordingApiContext) Lines() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.lines...)
}

// Containing filters Lines() by substring.
func (r *RecordingApiContext) Containing(substr string) []string {
	var out []string
	for _, l := range r.Lines() {
		if strings.Contains(l, substr) {
			out = append(out, l)
		}
	}
	return out
}

// HasLevel reports whether a line with the given prefix ("warn", "error",
// …) contains substr.
func (r *RecordingApiContext) HasLevel(level, substr string) bool {
	return len(r.matchLevel(level, substr)) > 0
}

func (r *RecordingApiContext) matchLevel(level, substr string) []string {
	var out []string
	for _, l := range r.Containing(level + ": ") {
		if strings.Contains(l, substr) {
			out = append(out, l)
		}
	}
	return out
}
