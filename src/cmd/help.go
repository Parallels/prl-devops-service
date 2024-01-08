package cmd

import (
	"fmt"

	"github.com/Parallels/pd-api-service/constants"
)

func processHelp(command string) {
	fmt.Printf("Parallels Desktop API Service manager\n")
	fmt.Printf("\n")
	fmt.Printf("  Find out more at: https://github.com/Parallesl/pd-api-service\n")
	fmt.Printf("\n")
	switch command {
	case constants.API_COMMAND:
		processApiHelp()
	case constants.CATALOG_COMMAND:
		processCatalogHelp()
	default:
		processDefaultHelp()
	}
	fmt.Printf("\n")
}

func processDefaultHelp() {
	fmt.Printf("Usage:\n")
	fmt.Printf("\n")
	fmt.Printf("  pd-api-service [command] [flags]\n")
	fmt.Printf("\n")
	fmt.Printf("Available Commands:\n")
	fmt.Printf("\n")
	fmt.Printf("  %s\t\t Starts the API Service\n", constants.API_COMMAND)
	fmt.Printf("  %s\t Prints the API Catalog\n", constants.CATALOG_COMMAND)
	fmt.Printf("  %s\t Generates a new Security Key\n", constants.GENERATE_SECURITY_KEY_COMMAND)
	fmt.Printf("  %s\t Installs the API Service\n", constants.INSTALL_SERVICE_COMMAND)
	fmt.Printf("  %s\t Uninstalls the API Service\n", constants.UNINSTALL_SERVICE_COMMAND)
	fmt.Printf("  %s\t\t Tests the Remote providers\n", constants.TEST_COMMAND)
	fmt.Printf("  %s\t Prints the Version\n", constants.VERSION_COMMAND)
	fmt.Printf("  %s\t\t Prints this help\n", constants.HELP_COMMAND)
}
