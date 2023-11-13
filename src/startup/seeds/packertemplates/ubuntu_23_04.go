package packertemplates

import (
	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/data"
	"github.com/Parallels/pd-api-service/data/models"
)

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
