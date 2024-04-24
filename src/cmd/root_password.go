package cmd

import (
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/cjlapao/common-go/helper"
)

func processRootPassword(ctx basecontext.ApiContext) {
	ctx.LogInfof("Updating root password")
	rootPassword := helper.GetFlagValue(constants.PASSWORD_FLAG, "")
	if rootPassword != "" {
		serviceprovider.InitServices(ctx)
		db := serviceprovider.Get().JsonDatabase
		ctx.LogInfof("Database connection found, updating password")
		_ = db.Connect(ctx)
		if db != nil {
			err := db.UpdateRootPassword(ctx, rootPassword)
			if err != nil {
				panic(err)
			}
			_ = db.Disconnect(ctx)
		} else {
			panic(data.ErrDatabaseNotConnected)
		}
	} else {
		panic("No password provided")
	}
	ctx.LogInfof("Root password updated")

	os.Exit(0)
}
