package cmd

import (
	"fmt"
	"os"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/pdfile"
	"github.com/Parallels/pd-api-service/serviceprovider/system"
	"github.com/cjlapao/common-go/helper"
)

func processCatalog(ctx basecontext.ApiContext) {
	subcommand := helper.GetCommandAt(1)
	// processing the command help
	if helper.GetFlagSwitch(constants.HELP_FLAG, false) || helper.GetCommandAt(1) == "help" {
		processHelp(constants.API_COMMAND)
		os.Exit(0)
	}

	switch subcommand {
	case "run":
		processCatalogRunCmd(ctx)
	case "list":
		processCatalogListCmd(ctx)
	case "push":
		fmt.Println("Starting push...")
		processCatalogPushCmd(ctx)
	case "pull":
		fmt.Println("Starting pull...")
		processCatalogPullCmd(ctx)
	case "delete":
		fmt.Println("delete")
	case "import":
		fmt.Println("import")
	default:
		processHelp(constants.CATALOG_COMMAND)
	}

	os.Exit(0)
}

func processCatalogHelp() {
	fmt.Println("Usage:")
	fmt.Printf("  %v %v <command>\n", constants.ExecutableName, constants.CATALOG_COMMAND)
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  run\t\t\t\t\tRun a catalog file")
	fmt.Println("  list\t\t\t\t\tList all catalogs")
	fmt.Println("  push <catalog>\t\t\tPush a catalog to the server")
	fmt.Println("  pull <catalog>\t\t\tPull a catalog from the server")
	fmt.Println("  delete <catalog>\t\t\tDelete a catalog from the server")
	fmt.Println("  import <catalog> <file>\t\tImport a catalog from a file")
}

func catalogInitPdFile(ctx basecontext.ApiContext, cmd string) *pdfile.PDFile {
	var pdFile *pdfile.PDFile
	var diag *pdfile.PDFileDiagnostics
	if helper.GetFlagValue(constants.FILE_FLAG, "") != "" {
		pdFile, diag = pdfile.Load(helper.GetFlagValue(constants.FILE_FLAG, ""))
		if diag.HasErrors() {
			ctx.EnableLog()
			ctx.ToggleLogTimestamps(false)
			ctx.LogErrorf("There was errors loading the pd file:")
			for _, err := range diag.Errors() {
				ctx.LogErrorf("  - %v", err)
			}
			os.Exit(1)
		}
		catalogGetFlags(pdFile)
	} else {
		pdFile = pdfile.NewPdFile()
		catalogGetFlags(pdFile)
	}

	if cmd != "" {
		pdFile.Command = cmd
	}

	if pdFile.Owner == "" {
		user, _ := system.Get().GetCurrentUser(ctx)
		if user != "" {
			pdFile.Owner = user
		}
	}

	validationDiag := pdFile.Validate()
	if validationDiag.HasErrors() {
		ctx.EnableLog()
		ctx.ToggleLogTimestamps(false)
		ctx.LogErrorf("There was errors validating the pd file:")
		for _, err := range validationDiag.Errors() {
			ctx.LogErrorf("  - %v", err)
		}

		os.Exit(1)
	}

	return pdFile
}

func catalogGetFlags(pdFile *pdfile.PDFile) {
	if helper.GetFlagValue(constants.PD_FILE_FROM_FLAG, "") != "" {
		pdFile.From = helper.GetFlagValue(constants.PD_FILE_FROM_FLAG, "")
	}
	if helper.GetFlagValue(constants.PD_FILE_ARCHITECTURE_FLAG, "") != "" {
		pdFile.Architecture = helper.GetFlagValue(constants.PD_FILE_ARCHITECTURE_FLAG, "")
	}
	if helper.GetFlagValue(constants.PD_FILE_CATALOG_ID_FLAG, "") != "" {
		pdFile.CatalogId = helper.GetFlagValue(constants.PD_FILE_CATALOG_ID_FLAG, "")
	}
	if helper.GetFlagValue(constants.PD_FILE_VERSION_FLAG, "") != "" {
		pdFile.Version = helper.GetFlagValue(constants.PD_FILE_VERSION_FLAG, "")
	}
	if helper.GetFlagValue(constants.PD_FILE_INSECURE_FLAG, "") != "" {
		pdFile.Insecure = helper.GetFlagSwitch(constants.PD_FILE_INSECURE_FLAG, false)
	}
	if helper.GetFlagValue(constants.PD_FILE_ROLE_FLAG, "") != "" {
		pdFile.Roles = helper.GetFlagArrayValue(constants.PD_FILE_ROLE_FLAG)
	}
	if helper.GetFlagValue(constants.PD_FILE_CLAIM_FLAG, "") != "" {
		pdFile.Claims = helper.GetFlagArrayValue(constants.PD_FILE_CLAIM_FLAG)
	}
	if helper.GetFlagValue(constants.PD_FILE_TAG_FLAG, "") != "" {
		pdFile.Tags = helper.GetFlagArrayValue(constants.PD_FILE_TAG_FLAG)
	}
	// if helper.GetFlagValue(constants.PD_FILE_PROVIDER_FLAG, "") != "" {
	//   pdFile.Provider = helper.GetFlagValue(constants.PD_FILE_PROVIDER_FLAG, "")
	// }
	if helper.GetFlagValue(constants.PD_FILE_RUN_FLAG, "") != "" {
		pdFile.Command = helper.GetFlagValue(constants.PD_FILE_RUN_FLAG, "")
	}
	if helper.GetFlagValue(constants.PD_FILE_DESTINATION_FLAG, "") != "" {
		pdFile.Destination = helper.GetFlagValue(constants.PD_FILE_DESTINATION_FLAG, "")
	}
	if helper.GetFlagValue(constants.PD_FILE_DESCRIPTION_FLAG, "") != "" {
		pdFile.Description = helper.GetFlagValue(constants.PD_FILE_DESCRIPTION_FLAG, "")
	}
	if helper.GetFlagValue(constants.PD_FILE_OWNER_FLAG, "") != "" {
		pdFile.Owner = helper.GetFlagValue(constants.PD_FILE_OWNER_FLAG, "")
	}
	if helper.GetFlagSwitch(constants.PD_FILE_START_AFTER_PULL_FLAG, false) {
		pdFile.StartAfterPull = helper.GetFlagSwitch(constants.PD_FILE_START_AFTER_PULL_FLAG, false)
	}
	if helper.GetFlagValue(constants.PD_FILE_MACHINE_NAME_FLAG, "") != "" {
		pdFile.MachineName = helper.GetFlagValue(constants.PD_FILE_MACHINE_NAME_FLAG, "")
	}
	if helper.GetFlagValue(constants.PD_FILE_TAG_FLAG, "") != "" {
		pdFile.Tags = helper.GetFlagArrayValue(constants.PD_FILE_TAG_FLAG)
	}

	if helper.GetFlagValue(constants.PD_FILE_USERNAME_FLAG, "") != "" || helper.GetFlagValue(constants.PD_FILE_API_KEY_FLAG, "") != "" {
		if pdFile.Authentication == nil {
			pdFile.Authentication = &pdfile.PDFileAuthentication{}
		}

		pdFile.Authentication.Username = helper.GetFlagValue(constants.PD_FILE_USERNAME_FLAG, "")
		pdFile.Authentication.Password = helper.GetFlagValue(constants.PD_FILE_PASSWORD_FLAG, "")
		pdFile.Authentication.ApiKey = helper.GetFlagValue(constants.PD_FILE_API_KEY_FLAG, "")
	}
}

func processCatalogRunCmd(ctx basecontext.ApiContext) {
	pdFile := catalogInitPdFile(ctx, "")

	out, diags := pdFile.Run(ctx)
	if diags.HasErrors() {
		for _, err := range diags.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	fmt.Printf("%s\n", out)
}

func processCatalogListCmd(ctx basecontext.ApiContext) {
	pdFile := catalogInitPdFile(ctx, "list")

	out, diags := pdFile.Run(ctx)
	if diags.HasErrors() {
		fmt.Println(diags.Errors())
		os.Exit(1)
	}

	fmt.Printf("%s\n", out)
}

func processCatalogPushCmd(ctx basecontext.ApiContext) {
	file := helper.GetFlagValue(constants.FILE_FLAG, "")
	if file == "" {
		fmt.Println("Missing file flag")
		return
	}

	pdFile := catalogInitPdFile(ctx, "push")

	out, diags := pdFile.Run(ctx)
	if diags.HasErrors() {
		for _, err := range diags.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	fmt.Printf("%s\n", out)
}

func processCatalogPullCmd(ctx basecontext.ApiContext) {
	file := helper.GetFlagValue(constants.FILE_FLAG, "")
	if file == "" {
		fmt.Println("Missing file flag")
		return
	}

	pdFile := catalogInitPdFile(ctx, "pull")

	out, diags := pdFile.Run(ctx)
	if diags.HasErrors() {
		for _, err := range diags.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	fmt.Printf("%s\n", out)
}
