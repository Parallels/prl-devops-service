package main

import (
	"github.com/Parallels/prl-devops-service/cmd"
	"github.com/Parallels/prl-devops-service/constants"

	"github.com/cjlapao/common-go/version"
)

var (
	ver        = "0.6.1"
	versionSvc = version.Get()
)

//	@title			Parallels Desktop DevOps Service
//	@version		0.6.1
//	@description	Parallels Desktop DevOps Service
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Parallels Desktop DevOps Support
//	@contact.url	https://forum.parallels.com/
//	@contact.email	carlos.lapao@parallels.com

//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@BasePath					/api
//	@securityDefinitions.apikey	ApiKeyAuth
//	@description				Type the api key in the input below.
//	@in							header
//	@name						X-Api-Key

//	@securityDefinitions.apikey	BearerAuth
//	@description				Type "Bearer" followed by a space and JWT token.
//	@in							header
//	@name						Authorization
func main() {
	versionSvc.Author = "Carlos Lapao"
	versionSvc.Name = constants.Name
	versionSvc.License = "Fair Source (https://fair.io)"
	// Reading the version from a string
	strVer, err := version.FromString(ver)
	if err == nil {
		versionSvc.Major = strVer.Major
		versionSvc.Minor = strVer.Minor
		versionSvc.Build = strVer.Build
		versionSvc.Rev = strVer.Rev
	}

	cmd.Process()
}
