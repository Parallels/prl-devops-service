package packertemplates

import (
	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/data"
	"github.com/Parallels/pd-api-service/data/models"
)

func AddMacOs14_0Manual(ctx *basecontext.BaseContext, svc *data.JsonDatabase) error {
	ubuntu2304Template := models.PackerTemplate{
		ID:           "macos14_23A344_ipsw",
		Name:         "macOS Sonoma 14.0",
		Description:  "This will create a macOS Sonoma 14.0 VM using a downloaded IPSW file, this will need user input to complete the installation.",
		PackerFolder: "macos",
		Internal:     true,
		Variables: map[string]string{
			"ipsw_url":      "https://updates.cdn-apple.com/2023FallFCS/fullrestores/042-54934/0E101AD6-3117-4B63-9BF1-143B6DB9270A/UniversalMac_14.0_23A344_Restore.ipsw",
			"ipsw_checksum": "sha256:c5a137b905a3f9fc4fb7bba16abfa625c9119154f93759f571aa1c915d3d9664",
		},
		Addons: []string{},
		Specs: map[string]string{
			"memory": "4096",
			"cpu":    "4",
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
		common.Logger.Error("Error validating macOS Sonoma 14.0 template: %s", err.Error())
		return err
	} else {
		if _, err := svc.AddPackerTemplate(ctx, &ubuntu2304Template); err != nil {
			if err.Error() != data.ErrPackerTemplateAlreadyExists.Error() {
				common.Logger.Error("Error adding macOS Sonoma 14.0 template: %s", err.Error())
				return err
			}
		} else {
			common.Logger.Info("macOS Sonoma 14.0 template added")
		}
	}

	return nil
}
