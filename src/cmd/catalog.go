package cmd

import (
	"fmt"
	"os"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/pdfile"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
	"github.com/Parallels/prl-devops-service/serviceprovider/parallelsdesktop"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
	"github.com/cjlapao/common-go/helper"
)

func processCatalog(ctx basecontext.ApiContext, operation string, filePath string) {
	// processing the command help
	if helper.GetFlagSwitch(constants.HELP_FLAG, false) || helper.GetCommandAt(1) == "help" {
		processHelp(constants.CATALOG_COMMAND)
		os.Exit(0)
	}
	ctx.ToggleLogTimestamps(false)
	_ = os.Setenv(constants.SOURCE_ENV_VAR, "catalog")

	if operation != "list" {
		if filePath == "" {
			ctx.LogInfof("The filePath is empty")
			filePath = helper.GetFlagValue(constants.FILE_FLAG, "")
			if filePath == "" {
				fmt.Println("Could not find a file to process, did you miss adding the flag --file=<file>?")
				return
			}
		} else {
			if !helper.FileExists(filePath) {
				ctx.LogErrorf("File with path %v does not exists, exiting", filePath)
				os.Exit(1)
			}
		}
	}

	switch operation {
	case "run":
		processCatalogRunCmd(ctx, filePath)
	case "list":
		processCatalogListCmd(ctx, filePath)
	case "push":
		fmt.Println("Starting push, this can take a while...")
		processCatalogPushCmd(ctx, filePath)
	case "pull":
		fmt.Println("Starting pull, this can take a while...")
		processCatalogPullCmd(ctx, filePath)
	case "delete":
		fmt.Println("Not implemented yet")
	case "import":
		fmt.Println("Not implemented yet")
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

func catalogInitPdFile(ctx basecontext.ApiContext, cmd string, filepath string) *pdfile.PDFileService {
	var pdFile *models.PDFile
	var diag *diagnostics.PDFileDiagnostics
	if filepath != "" {
		pdFile, diag = pdfile.Load(ctx, filepath)
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
		pdFile = models.NewPdFile()
		catalogGetFlags(pdFile)
	}
	if pdFile.Destination == "" {
		pdService := parallelsdesktop.New(ctx)
		if pdService != nil {
			info, err := pdService.GetInfo()
			if err != nil {
				ctx.LogErrorf("Error getting info from parallels desktop: %v", err)
			}
			if err == nil && info != nil {
				pdFile.Destination = info.VMHome
			}
		}
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

	svc := pdfile.NewPDFileService(ctx, pdFile)

	validationDiag := svc.Validate()
	if validationDiag.HasErrors() {
		ctx.EnableLog()
		ctx.ToggleLogTimestamps(false)
		ctx.LogErrorf("There was errors validating the pd file:")
		for _, err := range validationDiag.Errors() {
			ctx.LogErrorf("  - %v", err)
		}

		os.Exit(1)
	}

	return svc
}

func catalogGetFlags(pdFile *models.PDFile) {
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
			pdFile.Authentication = &models.PDFileAuthentication{}
		}

		pdFile.Authentication.Username = helper.GetFlagValue(constants.PD_FILE_USERNAME_FLAG, "")
		pdFile.Authentication.Password = helper.GetFlagValue(constants.PD_FILE_PASSWORD_FLAG, "")
		pdFile.Authentication.ApiKey = helper.GetFlagValue(constants.PD_FILE_API_KEY_FLAG, "")
	}
}

func processCatalogRunCmd(ctx basecontext.ApiContext, filepath string) {
	pdFile := catalogInitPdFile(ctx, "", filepath)

	out, diags := pdFile.Run(ctx)
	if diags.HasErrors() {
		for _, err := range diags.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	ctx.LogInfof("%v", out)
}

func processCatalogListCmd(ctx basecontext.ApiContext, filepath string) {
	svc := catalogInitPdFile(ctx, "list", filepath)

	out, diags := svc.Run(ctx)
	if diags.HasErrors() {
		fmt.Println(diags.Errors())
		os.Exit(1)
	}

	ctx.LogInfof("%v", out)
}

func processCatalogPushCmd(ctx basecontext.ApiContext, filePath string) {
	svc := catalogInitPdFile(ctx, "push", filePath)

	out, diags := svc.Run(ctx)

	if diags.HasErrors() {
		for _, err := range diags.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	// Stop the progress bar by printing a new line
	fmt.Println()

	ctx.LogInfof("%v", out)
}

func processCatalogPullCmd(ctx basecontext.ApiContext, filePath string) {
	svc := catalogInitPdFile(ctx, "pull", filePath)

	out, diags := svc.Run(ctx)
	if diags.HasErrors() {
		for _, err := range diags.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	ctx.LogInfof("%v", out)
}

func processCatalogImportCmd(ctx basecontext.ApiContext, filePath string) {
	file := helper.GetFlagValue(constants.FILE_FLAG, "")
	if file == "" {
		fmt.Println("Missing file flag")
		return
	}

	svc := catalogInitPdFile(ctx, "import", filePath)

	out, diags := svc.Run(ctx)
	if diags.HasErrors() {
		for _, err := range diags.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	ctx.LogInfof("%v", out)
}
