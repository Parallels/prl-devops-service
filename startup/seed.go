package startup

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/services"
)

func SeedVirtualMachineTemplateDefaults() {
	svc := services.GetServices().JsonDatabase
	err := svc.Connect()
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return
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
	} else {
		if err := svc.AddVirtualMachineTemplate(&ubuntu2304Template); err != nil {
			if err.Error() != "Machine Template already exists" {
				common.Logger.Error("Error adding Ubuntu 23.04 template: %s", err.Error())
			}
		} else {
			common.Logger.Info("Ubuntu 23.04 template added")
		}
	}
}
