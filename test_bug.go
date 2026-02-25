package main

import (
	"fmt"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
)

func main() {
	ctx := basecontext.NewBaseContext()
	cfg := config.New(ctx)
	// simulate Linux config with MODE=orchestrator
	cfg.SetKey("MODE", "orchestrator")

	fmt.Printf("Before disable host: %v\n", cfg.GetEnabledModules())

	if cfg.IsModuleEnabled("host") {
		// Parallels Desktop is not available on Linux
		cfg.DisableModule("host")
	}

	fmt.Printf("After disable host: %v\n", cfg.GetEnabledModules())

	// Test with ENABLED_MODULES explicitly set
	cfg2 := config.New(ctx)
	cfg2.SetKey("ENABLED_MODULES", "api,orchestrator")
	fmt.Printf("With ENABLED_MODULES explicitly set: %v\n", cfg2.GetEnabledModules())
}
