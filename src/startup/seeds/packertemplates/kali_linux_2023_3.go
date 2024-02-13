package packertemplates

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
)

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
