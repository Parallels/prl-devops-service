package seeds

import (
	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/serviceprovider"
	"github.com/Parallels/pd-api-service/startup/seeds/packertemplates"
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

	packertemplates.AddUbuntu23_04(ctx, svc)
	packertemplates.AddKaliLinux2023_3_gnome(ctx, svc)
	packertemplates.AddMacOs14_0Manual(ctx, svc)

	return nil
}
