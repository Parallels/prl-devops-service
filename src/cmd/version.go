package cmd

import (
	"os"

	"github.com/Parallels/prl-devops-service/constants"
	appversion "github.com/Parallels/prl-devops-service/version"
)

var versionSvc = appversion.Current(constants.Name)

func processVersion() {
	versionSvc.PrintSimpleVersion()
	os.Exit(0)
}
