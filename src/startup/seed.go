package startup

import (
	"fmt"

	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/startup/seeds"
)

func SeedDefaults() (err error) {
	defer func() {
		if r := recover(); r != nil {
			common.Logger.Error("PANIC in SeedDefaults: %v", r)
			err = fmt.Errorf("panic in SeedDefaults: %v", r)
		}
	}()

	if err = seeds.SeedDefaultConfig(); err != nil {
		common.Logger.Error("Error seeding default config: %s", err.Error())
		return fmt.Errorf("SeedDefaultConfig failed: %w", err)
	}
	if err = seeds.SeedDefaultClaims(); err != nil {
		common.Logger.Error("Error seeding default claims: %s", err.Error())
		return fmt.Errorf("SeedDefaultClaims failed: %w", err)
	}
	if err = seeds.SeedDefaultRoles(); err != nil {
		common.Logger.Error("Error seeding default roles: %s", err.Error())
		return fmt.Errorf("SeedDefaultRoles failed: %w", err)
	}
	if err = seeds.SeedDefaultRoleClaims(); err != nil {
		common.Logger.Error("Error seeding default role claims: %s", err.Error())
		return fmt.Errorf("SeedDefaultRoleClaims failed: %w", err)
	}
	if err = seeds.SeedDefaultUsers(); err != nil {
		common.Logger.Error("Error seeding admin user: %s", err.Error())
		return fmt.Errorf("SeedDefaultUsers failed: %w", err)
	}
	if err = seeds.SeedUsersMissingClaimsByRole(); err != nil {
		common.Logger.Error("Error syncing role-based user claims: %s", err.Error())
		return fmt.Errorf("SeedUsersMissingClaimsByRole failed: %w", err)
	}
	if err = seeds.SeedDefaultVirtualMachineTemplates(); err != nil {
		common.Logger.Error("Error seeding default virtual machine templates: %s", err.Error())
		return fmt.Errorf("SeedDefaultVirtualMachineTemplates failed: %w", err)
	}

	return nil
}
