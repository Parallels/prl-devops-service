package cmd

import (
	"fmt"
	"os"

	"github.com/cjlapao/common-go/version"
)

var versionSvc = version.Get()

func processVersion() {
	fmt.Printf("%s\n", versionSvc.String())
	os.Exit(0)
}
