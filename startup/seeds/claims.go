package seeds

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/service_provider"
)

func SeedDefaultClaims() error {
	db := service_provider.Get().JsonDatabase
	err := db.Connect()
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	defer db.Disconnect()

	if exists, _ := db.GetClaim(constants.READ_ONLY_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.READ_ONLY_CLAIM,
			Name:     constants.READ_ONLY_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.CREATE_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.CREATE_CLAIM,
			Name:     constants.CREATE_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.DELETE_CLAIM); exists == nil {

		if err := db.CreateClaim(&models.Claim{
			ID:       constants.DELETE_CLAIM,
			Name:     constants.DELETE_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.LIST_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.LIST_CLAIM,
			Name:     constants.LIST_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	// VMS
	if exists, _ := db.GetClaim(constants.CREATE_VM_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.CREATE_VM_CLAIM,
			Name:     constants.CREATE_VM_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.UPDATE_VM_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.UPDATE_VM_CLAIM,
			Name:     constants.UPDATE_VM_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.UPDATE_VM_STATES_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.UPDATE_VM_CLAIM,
			Name:     constants.UPDATE_VM_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.DELETE_VM_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.DELETE_VM_CLAIM,
			Name:     constants.DELETE_VM_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.LIST_VM_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.LIST_VM_CLAIM,
			Name:     constants.LIST_VM_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}
	if exists, _ := db.GetClaim(constants.EXECUTE_COMMAND_VM_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.EXECUTE_COMMAND_VM_CLAIM,
			Name:     constants.EXECUTE_COMMAND_VM_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	// TEMPLATES
	if exists, _ := db.GetClaim(constants.LIST_TEMPLATE_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.LIST_TEMPLATE_CLAIM,
			Name:     constants.LIST_TEMPLATE_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.CREATE_TEMPLATE_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.CREATE_TEMPLATE_CLAIM,
			Name:     constants.CREATE_TEMPLATE_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.DELETE_TEMPLATE_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.DELETE_TEMPLATE_CLAIM,
			Name:     constants.DELETE_TEMPLATE_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.UPDATE_TEMPLATE_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.UPDATE_TEMPLATE_CLAIM,
			Name:     constants.UPDATE_TEMPLATE_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.CREATE_CATALOG_MANIFEST_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.CREATE_CATALOG_MANIFEST_CLAIM,
			Name:     constants.CREATE_CATALOG_MANIFEST_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.DELETE_CATALOG_MANIFEST_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.DELETE_CATALOG_MANIFEST_CLAIM,
			Name:     constants.DELETE_CATALOG_MANIFEST_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.UPDATE_CATALOG_MANIFEST_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.UPDATE_CATALOG_MANIFEST_CLAIM,
			Name:     constants.UPDATE_CATALOG_MANIFEST_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.LIST_CATALOG_MANIFEST_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.LIST_CATALOG_MANIFEST_CLAIM,
			Name:     constants.LIST_CATALOG_MANIFEST_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.PULL_CATALOG_MANIFEST_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.PULL_CATALOG_MANIFEST_CLAIM,
			Name:     constants.PULL_CATALOG_MANIFEST_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.PUSH_CATALOG_MANIFEST_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.Claim{
			ID:       constants.PUSH_CATALOG_MANIFEST_CLAIM,
			Name:     constants.PUSH_CATALOG_MANIFEST_CLAIM,
			Internal: true,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	db.Disconnect()

	return nil
}
