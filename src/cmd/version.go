package cmd

import (
	"os"

	"github.com/cjlapao/common-go/version"
)

var versionSvc = version.Get()

func processVersion() {
	println(versionSvc.String())
	os.Exit(0)
}
