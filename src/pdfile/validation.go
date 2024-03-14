package pdfile

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
)

func (p *PDFileService) Validate() *diagnostics.PDFileDiagnostics {
	diag := diagnostics.NewPDFileDiagnostics()
	hasFromOrTo := false
	hasProvider := false
	hasProviderName := false
	hasAuthentication := false
	isUsernamePresent := false
	isPasswordPresent := false
	isApiKeyPresent := false
	runCMD := ""

	for i, line := range p.pdfile.Raw {
		if line == "" {
			continue
		}
		if line[0] == '#' {
			continue
		}

		parts := strings.Split(line, " ")
		switch strings.ToUpper(parts[0]) {
		case "FROM":
			hasFromOrTo = true
			continue
		case "TO":
			hasFromOrTo = true
			continue
		case "AUTHENTICATE":
			switch strings.ToUpper(parts[1]) {
			case "USERNAME":
				isUsernamePresent = true
				continue
			case "PASSWORD":
				isPasswordPresent = true
				continue
			case "API_KEY":
				isApiKeyPresent = true
				continue
			default:
				diag.AddError(errors.New("invalid authentication type"))
			}
		case "CATALOG_ID":
			continue
		case "VERSION":
			continue
		case "ARCHITECTURE":
			continue
		case "LOCAL_PATH":
			continue
		case "ROLE":
			continue
		case "CLAIM":
			continue
		case "TAG":
			continue
		case "MACHINE_NAME":
			continue
		case "OWNER":
			continue
		case "START_AFTER_PULL":
			continue
		case "DESTINATION":
			continue
		case "DESCRIPTION":
			continue
		case "PROVIDER":
			namePart := ""
			hasProvider = true
			subparts := strings.Split(parts[1], ";")
			if len(subparts) > 1 {
				namePart = subparts[0]
			} else {
				namePart = parts[1]
			}
			if strings.ToUpper(namePart) == "NAME" || namePart != "" {
				hasProviderName = true
				continue
			}
		case "RUN":
			runCMD = strings.Join(parts[1:], " ")
			continue
		case "IMPORT":
			runCMD = "import"
			continue
		case "PUSH":
			runCMD = "push"
			continue
		case "PULL":
			runCMD = "pull"
			continue
		case "DELETE":
			runCMD = "delete"
			continue
		case "DO":
			runCMD = strings.Join(parts[1:], " ")
			continue
		case "INSECURE":
			diag.AddWarning(fmt.Errorf("insecure flag found at line %d, do not use in production", i))
			continue
		default:
			diag.AddError(fmt.Errorf("invalid command %v at line %d", parts[0], i))
		}
	}

	if hasProvider {
		if !hasProviderName {
			diag.AddError(fmt.Errorf("provider name not found"))
		}
	} else {
		if runCMD == "push" ||
			runCMD == "pull" ||
			runCMD == "delete" ||
			runCMD == "import" ||
			runCMD == "list" {
			diag.AddError(fmt.Errorf("provider command not found"))
		}
	}

	if runCMD == "" && p.pdfile.Command == "" {
		diag.AddError(fmt.Errorf("RUN command not found"))
	}

	if !hasFromOrTo {
		diag.AddError(fmt.Errorf("from command not found"))
	}

	if hasAuthentication {
		if isUsernamePresent && !isPasswordPresent {
			diag.AddError(fmt.Errorf("username was found but password was not found"))
		}
		if !isUsernamePresent && isPasswordPresent {
			diag.AddError(fmt.Errorf("password was found but username was not found"))
		}
		if !isUsernamePresent && !isPasswordPresent && !isApiKeyPresent {
			diag.AddError(fmt.Errorf("authentication was found but username, password or api key was not found"))
		}
	}

	cmd := runCMD
	if cmd == "" {
		cmd = p.pdfile.Command
	}

	switch cmd {
	case "pull":
		if p.pdfile.MachineName == "" {
			diag.AddError(fmt.Errorf("machine name not found in pd file"))
		}
		if p.pdfile.Destination == "" {
			diag.AddError(fmt.Errorf("destination not found in pd file"))
		}
		if p.pdfile.Owner == "" {
			diag.AddError(fmt.Errorf("owner not found in pd file"))
		}
	case "push":
		if p.pdfile.LocalPath == "" {
			diag.AddError(fmt.Errorf("local path not found in pd file"))
		}
		if p.pdfile.CatalogId == "" {
			diag.AddError(fmt.Errorf("catalog id not found in pd file"))
		}
		if p.pdfile.Provider == nil {
			diag.AddError(fmt.Errorf("provider not found in pd file"))
		}
	}

	return diag
}
