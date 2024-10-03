package cmd

import (
	"fmt"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/cjlapao/common-go/helper"
)

func processHelp(command string) {
	fmt.Printf("%v\n", constants.Name)
	fmt.Printf("\n")
	fmt.Printf("  Find out more at: https://github.com/Parallels/prl-devops-service\n")
	fmt.Printf("\n")
	switch command {
	case constants.API_COMMAND:
		processApiHelp()
	case constants.CATALOG_COMMAND:
		processCatalogHelp()
	case constants.TEST_COMMAND:
		processTestHelp()
	case constants.GENERATE_SECURITY_KEY_COMMAND:
		processGenerateSecurityKeyHelp()
	case constants.REVERSE_PROXY_COMMAND:
		processReverseProxyHelp()
	case constants.INSTALL_SERVICE_COMMAND:
		processInstallHelp()
	case constants.START_COMMAND,
		constants.STOP_COMMAND,
		constants.CLONE_COMMAND,
		constants.DELETE_COMMAND,
		constants.EXEC_COMMAND:
		command := helper.GetCommandAt(0)
		processParallelsDesktopHelp(command)
	default:
		processDefaultHelp()
	}
	fmt.Printf("\n")
}

func processDefaultHelp() {
	fmt.Printf("Usage:\n")
	fmt.Printf("\n")
	fmt.Printf("  %v [command] [flags]\n", constants.ExecutableName)
	fmt.Printf("\n")
	fmt.Printf("Available Commands:\n")
	fmt.Printf("\n")
	fmt.Printf("  %s\t\t\t Starts the API Service\n", constants.API_COMMAND)
	fmt.Printf("  %s\t\t Starts the Reverse Proxy Service\n", constants.REVERSE_PROXY_COMMAND)
	fmt.Printf("  %s\t\t Prints the API Catalog\n", constants.CATALOG_COMMAND)
	fmt.Printf("  %s\t\t Generates a new Security Key\n", constants.GENERATE_SECURITY_KEY_COMMAND)
	fmt.Printf("  %s\t\t Installs the API Service\n", constants.INSTALL_SERVICE_COMMAND)
	fmt.Printf("  %s\t\t Uninstalls the API Service\n", constants.UNINSTALL_SERVICE_COMMAND)
	fmt.Printf("  %s\t Updates the Root Password\n", constants.UPDATE_ROOT_PASSWORD_COMMAND)
	fmt.Printf("  %s\t\t\t Tests the Remote providers\n", constants.TEST_COMMAND)
	fmt.Printf("  %s\t\t Prints the Version\n", constants.VERSION_COMMAND)
	fmt.Printf("  %s\t\t\t Prints this help\n", constants.HELP_COMMAND)
}
