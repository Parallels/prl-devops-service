package startup

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/startup/seeds"
)

func SeedDefaults() error {
	if err := seeds.SeedDefaultClaims(); err != nil {
		common.Logger.Error("Error seeding default claims: %s", err.Error())
		return err
	}
	if err := seeds.SeedDefaultRoles(); err != nil {
		common.Logger.Error("Error seeding default roles: %s", err.Error())
		return err
	}
	if err := seeds.SeedDefaultUsers(); err != nil {
		common.Logger.Error("Error seeding admin user: %s", err.Error())
		return err
	}
	if err := seeds.SeedDefaultVirtualMachineTemplates(); err != nil {
		common.Logger.Error("Error seeding default virtual machine templates: %s", err.Error())
		return err
	}

	return nil
}
