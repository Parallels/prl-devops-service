package seeds

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/startup/seeds/packertemplates"
)

func SeedDefaultVirtualMachineTemplates() error {
	ctx := basecontext.NewRootBaseContext()
	svc := serviceprovider.Get().JsonDatabase
	err := svc.Connect(ctx)
	if err != nil {
		common.Logger.Error("Error connecting to database: %s", err.Error())
		return err
	}

	_ = packertemplates.AddUbuntu23_04(ctx, svc)
	_ = packertemplates.AddKaliLinux2023_3_gnome(ctx, svc)
	_ = packertemplates.AddMacOs14_0Manual(ctx, svc)

	return nil
}
