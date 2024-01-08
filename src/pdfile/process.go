package pdfile

import (
	"errors"
	"fmt"
	"strings"
)

func Process(fileContent string) (*PDFile, *PDFileDiagnostics) {
	result := NewPdFile()
	result.raw = strings.Split(fileContent, "\n")
	diag := NewPDFileDiagnostics()

	for i, line := range result.raw {
		if line == "" {
			continue
		}
		if line[0] == '#' {
			continue
		}

		parts := strings.Split(line, " ")
		switch strings.ToUpper(parts[0]) {
		case "FROM":
			if parts[0] != "FROM" {
				diag.AddWarning(fmt.Errorf("command %v at line %d is not in upper case", parts[0], i))
			}
			result.From = parts[1]
		case "AUTHENTICATE":
			if result.Authentication == nil {
				result.Authentication = &PDFileAuthentication{}
			}
			switch strings.ToUpper(parts[1]) {
			case "USERNAME":
				result.Authentication.Username = strings.TrimSpace(strings.Join(parts[2:], " "))
			case "PASSWORD":
				result.Authentication.Password = strings.TrimSpace(strings.Join(parts[2:], " "))
			case "API_KEY":
				result.Authentication.ApiKey = strings.TrimSpace(strings.Join(parts[2:], " "))
			default:
				diag.AddError(errors.New("invalid authentication type"))
			}
		case "CATALOG_ID":
			result.CatalogId = strings.TrimSpace(strings.Join(parts[1:], " "))
		case "VERSION":
			result.Version = strings.TrimSpace(strings.Join(parts[1:], " "))
		case "ARCHITECTURE":
			result.Architecture = strings.TrimSpace(strings.Join(parts[1:], " "))
		case "LOCAL_PATH":
			result.LocalPath = strings.TrimSpace(strings.Join(parts[1:], " "))
		case "ROLE":
			rolesStr := strings.TrimSpace(strings.Join(parts[1:], " "))
			roles := strings.Split(rolesStr, ",")
			for i, role := range roles {
				roles[i] = strings.TrimSpace(role)
			}
			result.Roles = append(result.Roles, roles...)
		case "CLAIM":
			claimsStr := strings.TrimSpace(strings.Join(parts[1:], " "))
			claims := strings.Split(claimsStr, ",")
			for i, claim := range claims {
				claims[i] = strings.TrimSpace(claim)
			}
			result.Claims = append(result.Claims, claims...)
		case "TAG":
			tagStr := strings.TrimSpace(strings.Join(parts[1:], " "))
			tags := strings.Split(tagStr, ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
			result.Tags = append(result.Tags, tags...)
		case "DESCRIPTION":
			result.Description = strings.TrimSpace(strings.Join(parts[1:], " "))
		case "MACHINE_NAME":
			result.MachineName = strings.TrimSpace(strings.Join(parts[1:], " "))
		case "OWNER":
			result.Owner = strings.TrimSpace(strings.Join(parts[1:], " "))
		case "START_AFTER_PULL":
			value := strings.TrimSpace(strings.Join(parts[1:], " "))
			result.StartAfterPull = strings.ToLower(value) == "true" || value == "1"
		case "DESTINATION":
			result.Destination = strings.TrimSpace(strings.Join(parts[1:], " "))
		case "PROVIDER":
			if result.Provider == nil {
				result.Provider = &PDFileProvider{}
			}
			if strings.ToUpper(parts[1]) == "NAME" {
				result.Provider.Name = strings.TrimSpace(strings.Join(parts[2:], " "))
			} else {
				provider, err := result.ParseProvider(strings.TrimSpace(strings.Join(parts[1:], " ")))
				if err != nil {
					diag.AddError(err)
				}

				if result.Provider.Attributes == nil {
					result.Provider.Attributes = make(map[string]string)
				}

				for key, value := range provider.Attributes {
					result.Provider.Attributes[key] = value
				}
			}
		case "RUN":
			result.Command = strings.TrimSpace(strings.Join(parts[1:], " "))
		case "INSECURE":
			value := strings.TrimSpace(strings.Join(parts[1:], " "))
			if strings.ToLower(value) == "true" || value == "1" {
				result.Insecure = true
			}
		default:
			diag.AddError(fmt.Errorf("invalid command %v at line %d", parts[0], i))
		}
	}

	return result, diag
}
