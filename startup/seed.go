package startup

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/services"
)

func SeedVirtualMachineTemplateDefaults() error {
	svc := services.GetServices().JsonDatabase
	err := svc.Connect()
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	defer svc.Disconnect()

	// adding Ubuntu Packer template
	ubuntu2304Template := models.VirtualMachineTemplate{
		ID:           "ubuntu-23.04",
		Name:         "Ubuntu 23.04",
		Description:  "Ubuntu 23.04 Packer template",
		Type:         models.VirtualMachineTemplateTypePacker,
		PackerFolder: "ubuntu",
		Variables: map[string]string{
			"iso_url":      "https://cdimage.ubuntu.com/releases/23.04/release/ubuntu-23.04-live-server-arm64.iso",
			"iso_checksum": "sha256:ad306616e37132ee00cc651ac0233b0e24b0b6e5e93b4a8ad36aa30c95b74e8c",
		},
		Addons: []string{
			"developer",
		},
		Specs: map[string]int{
			"memory": 2048,
			"cpu":    2,
			"disk":   20480,
		},
	}

	if err := ubuntu2304Template.Validate(); err != nil {
		common.Logger.Error("Error validating Ubuntu 23.04 template: %s", err.Error())
		return err
	} else {
		if err := svc.AddVirtualMachineTemplate(&ubuntu2304Template); err != nil {
			if err.Error() != "Machine Template already exists" {
				common.Logger.Error("Error adding Ubuntu 23.04 template: %s", err.Error())
				return err
			}
		} else {
			common.Logger.Info("Ubuntu 23.04 template added")
		}
	}

	return nil
}

func SeedDefaultRolesAndClaims() error {
	db := services.GetServices().JsonDatabase
	err := db.Connect()
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	defer db.Disconnect()

	if exists, _ := db.GetRole(constants.USER_ROLE); exists == nil {
		if err := db.CreateRole(&models.UserRole{
			ID:   helpers.GenerateId(),
			Name: constants.USER_ROLE,
		}); err != nil {
			common.Logger.Error("Error adding role: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetRole(constants.SUPER_USER_ROLE); exists == nil {
		if err := db.CreateRole(&models.UserRole{
			ID:   helpers.GenerateId(),
			Name: constants.SUPER_USER_ROLE,
		}); err != nil {
			common.Logger.Error("Error adding role: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.READ_ONLY_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.UserClaim{
			ID:   helpers.GenerateId(),
			Name: constants.READ_ONLY_CLAIM,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.CREATE_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.UserClaim{
			ID:   helpers.GenerateId(),
			Name: constants.CREATE_CLAIM,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.DELETE_CLAIM); exists == nil {

		if err := db.CreateClaim(&models.UserClaim{
			ID:   helpers.GenerateId(),
			Name: constants.DELETE_CLAIM,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.LIST_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.UserClaim{
			ID:   helpers.GenerateId(),
			Name: constants.LIST_CLAIM,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	// VMS
	if exists, _ := db.GetClaim(constants.CREATE_VM_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.UserClaim{
			ID:   helpers.GenerateId(),
			Name: constants.CREATE_VM_CLAIM,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.UPDATE_VM_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.UserClaim{
			ID:   helpers.GenerateId(),
			Name: constants.UPDATE_VM_CLAIM,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.UPDATE_VM_STATES_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.UserClaim{
			ID:   helpers.GenerateId(),
			Name: constants.UPDATE_VM_CLAIM,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.DELETE_VM_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.UserClaim{
			ID:   helpers.GenerateId(),
			Name: constants.DELETE_VM_CLAIM,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.LIST_VM_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.UserClaim{
			ID:   helpers.GenerateId(),
			Name: constants.LIST_VM_CLAIM,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	// TEMPLATES
	if exists, _ := db.GetClaim(constants.LIST_TEMPLATE_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.UserClaim{
			ID:   helpers.GenerateId(),
			Name: constants.LIST_TEMPLATE_CLAIM,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.CREATE_TEMPLATE_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.UserClaim{
			ID:   helpers.GenerateId(),
			Name: constants.CREATE_TEMPLATE_CLAIM,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.DELETE_TEMPLATE_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.UserClaim{
			ID:   helpers.GenerateId(),
			Name: constants.DELETE_TEMPLATE_CLAIM,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	if exists, _ := db.GetClaim(constants.UPDATE_TEMPLATE_CLAIM); exists == nil {
		if err := db.CreateClaim(&models.UserClaim{
			ID:   helpers.GenerateId(),
			Name: constants.CREATE_TEMPLATE_CLAIM,
		}); err != nil {
			common.Logger.Error("Error adding claim: %s", err.Error())
			return err
		}
	}

	db.Disconnect()

	return nil
}

func SeedAdminUser() error {
	db := services.GetServices().JsonDatabase
	err := db.Connect()
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	defer db.Disconnect()

	if exists, _ := db.GetUser("root"); exists != nil {
		return nil
	}

	suRole, err := db.GetRole(constants.SUPER_USER_ROLE)
	if err != nil {
		return err
	}
	claims, err := db.GetClaims()
	if err != nil {
		return err
	}

	if _, err := db.CreateUser(&models.User{
		ID:       services.GetServices().HardwareId,
		Name:     "Root",
		Username: "root",
		Email:    "root@localhost",
		Password: services.GetServices().HardwareSecret,
		Roles: []models.UserRole{
			*suRole,
		},
		Claims: claims,
	}); err != nil {
		common.Logger.Error("Error adding admin user: %s", err.Error())
		return err
	}

	db.Disconnect()

	return nil
}
