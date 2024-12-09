package common

import "regexp"

var (
	// Regular expressions to match the tags
	summaryRegex     = regexp.MustCompile(`^//\s*@Summary\s+(.*)$`)
	descriptionRegex = regexp.MustCompile(`^//\s*@Description\s+(.*)$`)
	contentRegex     = regexp.MustCompile(`^//\s*@Content\s+(.*)$`)
	tagsRegex        = regexp.MustCompile(`^//\s*@Tags\s+(.*)$`)
	rolesRegex       = regexp.MustCompile(`^//\s*@Roles\s+(.*)$`)
	claimsRegex      = regexp.MustCompile(`^//\s*@Claims\s+(.*)$`)
	paramRegex       = regexp.MustCompile(`^//\s*@Param\s+(.*)$`)
	successRegex     = regexp.MustCompile(`^//\s*@Success\s+(.*)$`)
	failureRegex     = regexp.MustCompile(`^//\s*@Failure\s+(.*)$`)
	routerRegex      = regexp.MustCompile(`^//\s*@Router\s+(.*)\s+\[(.*)\]$`)
	examplesRegex    = regexp.MustCompile(`^//\s*@Examples\s+(.*)$`)
)
