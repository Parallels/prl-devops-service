package cmd

import (
	"os"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/data"
	"github.com/Parallels/pd-api-service/serviceprovider"
	"github.com/cjlapao/common-go/helper"
)

func processRootPassword(ctx basecontext.ApiContext) {
	ctx.LogInfof("Updating root password")
	rootPassword := helper.GetFlagValue(constants.PASSWORD_FLAG, "")
	if rootPassword != "" {
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
