// package processors contains command processors for PDFile parsing
package processors

import (
	"strings"

	"github.com/Parallels/prl-devops-service/pdfile/models"
)

func getCommand(line string) *models.PDFileCommand {
	command := models.PDFileCommand{}
	if line == "" {
		return nil
	}
	if line[0] == '#' {
		return nil
	}

	parts := strings.Split(line, " ")
	command.Command = strings.ToUpper(parts[0])
	if len(parts) > 1 {
		command.Argument = strings.TrimSpace(strings.Join(parts[1:], " "))
	}

	return &command
}

func getBoolValue(arg string) bool {
	return strings.EqualFold(arg, "true") ||
		arg == "1" ||
		strings.EqualFold(arg, "t") ||
		strings.EqualFold(arg, "yes")
}
