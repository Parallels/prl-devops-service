package seeds

import (
	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/data"
	"github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/serviceprovider"
)

func SeedDefaultVirtualMachineTemplates() error {
	ctx := basecontext.NewRootBaseContext()
	svc := serviceprovider.Get().JsonDatabase
	err := svc.Connect(ctx)
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	defer svc.Disconnect(ctx)

	AddUbuntu23_04(ctx, svc)
	AddKaliLinux2023_3_gnome(ctx, svc)

	return nil
}

func AddUbuntu23_04(ctx *basecontext.BaseContext, svc *data.JsonDatabase) error {
	ubuntu2304Template := models.PackerTemplate{
		ID:           "ubuntu-23.04",
		Name:         "Ubuntu 23.04",
		Description:  "Ubuntu 23.04 Packer template",
		PackerFolder: "ubuntu",
		Internal:     true,
		Variables: map[string]string{
			"iso_url":      "https://cdimage.ubuntu.com/releases/23.04/release/ubuntu-23.04-live-server-arm64.iso",
			"iso_checksum": "sha256:ad306616e37132ee00cc651ac0233b0e24b0b6e5e93b4a8ad36aa30c95b74e8c",
		},
		Addons: []string{
			"developer",
		},
		Specs: map[string]string{
			"memory": "2048",
			"cpu":    "2",
			"disk":   "20480",
		},
		RequiredRoles: []string{
			constants.USER_ROLE,
			constants.SUPER_USER_ROLE,
		},
		RequiredClaims: []string{
			constants.LIST_PACKER_TEMPLATE_CLAIM,
			constants.CREATE_PACKER_TEMPLATE_CLAIM,
			constants.UPDATE_PACKER_TEMPLATE_CLAIM,
			constants.DELETE_PACKER_TEMPLATE_CLAIM,
		},
	}

	if err := ubuntu2304Template.Validate(); err != nil {
		common.Logger.Error("Error validating Ubuntu 23.04 template: %s", err.Error())
		return err
	} else {
		if _, err := svc.AddPackerTemplate(ctx, &ubuntu2304Template); err != nil {
			if err.Error() != data.ErrPackerTemplateAlreadyExists.Error() {
				common.Logger.Error("Error adding Ubuntu 23.04 template: %s", err.Error())
				return err
			}
		} else {
			common.Logger.Info("Ubuntu 23.04 template added")
		}
	}

	return nil
}

func AddKaliLinux2023_3_gnome(ctx *basecontext.BaseContext, svc *data.JsonDatabase) error {
	ubuntu2304Template := models.PackerTemplate{
		ID:           "kali-linux-2023.3-gnome",
		Name:         "Kali Linux 2023.3 (Gnome)",
		Description:  "This will create a Kali Linux 2023.3 with the Gnome UI VM using automated Packer scripts.",
		PackerFolder: "kali-linux",
		Internal:     true,
		Variables: map[string]string{
			"iso_url":      "https://cdimage.kali.org/kali-2023.3/kali-linux-2023.3-installer-arm64.iso",
			"iso_checksum": "sha256:41e3997b31639ec45363181d4fff68a2b6a1a07ed2e3458f1bcd11f3f2d9db9c",
			"desktop":      "gnome",
		},
		Addons: []string{},
		Specs: map[string]string{
			"memory": "4096",
			"cpu":    "4",
			"disk":   "20480",
		},
		RequiredRoles: []string{
			constants.USER_ROLE,
			constants.SUPER_USER_ROLE,
		},
		RequiredClaims: []string{
			constants.LIST_PACKER_TEMPLATE_CLAIM,
			constants.CREATE_PACKER_TEMPLATE_CLAIM,
			constants.UPDATE_PACKER_TEMPLATE_CLAIM,
			constants.DELETE_PACKER_TEMPLATE_CLAIM,
		},
	}

	if err := ubuntu2304Template.Validate(); err != nil {
		common.Logger.Error("Error validating Kali Linux 2023.3 template: %s", err.Error())
		return err
	} else {
		if _, err := svc.AddPackerTemplate(ctx, &ubuntu2304Template); err != nil {
			if err.Error() != data.ErrPackerTemplateAlreadyExists.Error() {
				common.Logger.Error("Error adding Kali Linux 2023.3 template: %s", err.Error())
				return err
			}
		} else {
			common.Logger.Info("Kali Linux 2023.3 template added")
		}
	}

	return nil
}
