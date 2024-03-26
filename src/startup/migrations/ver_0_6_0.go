package migrations

import (
	"github.com/Parallels/prl-devops-service/common"
)

type Version0_6_0 struct{}

func (v Version0_6_0) Apply() error {
	versionTarget := "0.6.0"
	svc, err := Init()
	if err != nil {
		return err
	}

	compareResult, err := compareVersions(svc.schemaVersion, versionTarget)
	if err != nil {
		common.Logger.Error("Error comparing versions: %s", err.Error())
		return err
	}

	if compareResult == VersionEqualToTarget || compareResult == VersionHigherThanTarget {
		svc.Context.LogDebugf("Schema version is already %s, skipping migration", versionTarget)
		return nil
	}

	svc.Context.LogInfof("Applying migration to version %s", versionTarget)
	hosts, err := svc.DbService.GetOrchestratorHosts(svc.Context, "")
	if err != nil {
		return err
	}

	for _, host := range hosts {
		host.Enabled = true
		_, err = svc.DbService.UpdateOrchestratorHost(svc.Context, &host)
		if err != nil {
			return err
		}
	}

	err = svc.DbService.UpdateSchemaVersion(svc.Context, versionTarget)
	if err != nil {
		return err
	}
	return nil
}
