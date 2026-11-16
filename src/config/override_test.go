package config

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
)

// newTestConfig builds an isolated Config with a synthetic config-file layer
// (same package: internals are reachable without touching global state
// beyond what New would touch anyway — globalConfig is left untouched).
func newTestConfig(env map[string]string) *Config {
	return &Config{
		mode:       "api",
		ctx:        basecontext.NewBaseContext(),
		fileFormat: "yaml",
		filename:   "prldevops_config",
		config:     ConfigFile{Environment: env},
	}
}

func TestGetKey_PrecedenceMatrixWithOverrideResolver(t *testing.T) {
	restore := SetOverrideResolver(nil)
	defer func() { SetOverrideResolver(restore) }()

	cfg := newTestConfig(map[string]string{"SETTINGS_TEST_KEY": "file_value"})

	// 1. No override resolver installed: config file layer answers.
	if got := cfg.GetKey("SETTINGS_TEST_KEY"); got != "file_value" {
		t.Fatalf("expected file_value, got %q", got)
	}

	// 2. Environment beats the config file (pre-existing posture).
	t.Setenv("SETTINGS_TEST_KEY", "env_value")
	if got := cfg.GetKey("SETTINGS_TEST_KEY"); got != "env_value" {
		t.Fatalf("expected env_value, got %q", got)
	}

	// 3. DB override beats env + file.
	SetOverrideResolver(func(key string) (string, bool) {
		if key == "SETTINGS_TEST_KEY" {
			return "db_value", true
		}
		return "", false
	})
	if got := cfg.GetKey("SETTINGS_TEST_KEY"); got != "db_value" {
		t.Fatalf("expected db_value, got %q", got)
	}

	// 4. Present-but-empty override is a deliberate clear (found semantics).
	SetOverrideResolver(func(key string) (string, bool) {
		if key == "SETTINGS_TEST_KEY" {
			return "", true
		}
		return "", false
	})
	if got := cfg.GetKey("SETTINGS_TEST_KEY"); got != "" {
		t.Fatalf("expected empty string (explicit clear), got %q", got)
	}

	// 5. Resolver miss (found=false) falls back to env unchanged.
	SetOverrideResolver(func(string) (string, bool) { return "", false })
	if got := cfg.GetKey("SETTINGS_TEST_KEY"); got != "env_value" {
		t.Fatalf("expected env_value fallback, got %q", got)
	}

	// 6. Removing the resolver restores the byte-identical old posture.
	SetOverrideResolver(nil)
	if got := cfg.GetKey("SETTINGS_TEST_KEY"); got != "env_value" {
		t.Fatalf("expected env_value after removal, got %q", got)
	}
}

func TestSetOverrideResolver_ReturnsPrevious(t *testing.T) {
	restore := SetOverrideResolver(nil)
	defer func() { SetOverrideResolver(restore) }()

	first := func(string) (string, bool) { return "a", true }
	second := func(string) (string, bool) { return "b", true }

	SetOverrideResolver(first)
	previous := SetOverrideResolver(second)
	if previous == nil {
		t.Fatal("expected the previous resolver to be returned")
	}
	cfg := newTestConfig(nil)
	if got := cfg.GetKey("ANY_KEY"); got != "b" {
		t.Fatalf("second resolver must win, got %q", got)
	}
}

func TestConfigFileValue_ReadsFileLayerAndFlagFallback(t *testing.T) {
	cfg := newTestConfig(map[string]string{"Some_Key": "file_value"})
	value, ok := cfg.ConfigFileValue("SOME_KEY") // EqualFold lookup
	if !ok || value != "file_value" {
		t.Fatalf("expected case-insensitive file hit, got %q ok=%v", value, ok)
	}
	if _, ok := cfg.ConfigFileValue("NOT_SET_ANYWHERE"); ok {
		t.Fatal("expected miss for unknown key")
	}
}
