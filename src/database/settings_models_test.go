package database

import (
	"testing"

	"github.com/Parallels/prl-devops-service/data/models"
)

// TestAllGormModelsIncludesSettingsDomain guards the boot-registry contract:
// the five settings tables must ride AllGormModels or AutoMigrate never
// creates them on upgraded installs.
func TestAllGormModelsIncludesSettingsDomain(t *testing.T) {
	modelsList := AllGormModels()
	want := map[string]bool{
		"SettingDefinition": false,
		"SettingValue":      false,
		"SettingSyncState":  false,
		"NodeSetting":       false,
		"SettingMeta":       false,
	}
	for _, m := range modelsList {
		switch m.(type) {
		case *models.SettingDefinition:
			want["SettingDefinition"] = true
		case *models.SettingValue:
			want["SettingValue"] = true
		case *models.SettingSyncState:
			want["SettingSyncState"] = true
		case *models.NodeSetting:
			want["NodeSetting"] = true
		case *models.SettingMeta:
			want["SettingMeta"] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("AllGormModels missing %s", name)
		}
	}
}
