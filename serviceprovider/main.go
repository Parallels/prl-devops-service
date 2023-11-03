package serviceprovider

import (
	"Parallels/pd-api-service/basecontext"
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/data"
	"Parallels/pd-api-service/errors"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/serviceprovider/git"
	"Parallels/pd-api-service/serviceprovider/interfaces"
	"Parallels/pd-api-service/serviceprovider/packer"
	"Parallels/pd-api-service/serviceprovider/parallelsdesktop"
	"Parallels/pd-api-service/serviceprovider/system"
	"Parallels/pd-api-service/serviceprovider/vagrant"
	sql_database "Parallels/pd-api-service/sql"
	"encoding/base64"
	"fmt"
	"strings"

	log "github.com/cjlapao/common-go-logger"
)

type ServiceProvider struct {
	RunningUser             string
	Logger                  *log.LoggerService
	System                  *system.SystemService
	ParallelsDesktopService *parallelsdesktop.ParallelsService
	GitService              *git.GitService
	PackerService           *packer.PackerService
	VagrantService          *vagrant.VagrantService
	MySqlService            *sql_database.MySQLService
	JsonDatabase            *data.JsonDatabase
	Services                []interfaces.Service
	HardwareInfo            *models.ParallelsDesktopInfo
	HardwareId              string
	HardwareSecret          string
	CurrentSystemUser       string
}

var globalProvider *ServiceProvider

func InitCatalogServices() {
	globalProvider = &ServiceProvider{
		Logger: common.Logger,
	}

	globalProvider.System = system.New()
	globalProvider.System.SetDependencies([]interfaces.Service{})
	globalProvider.Services = append(globalProvider.Services, globalProvider.System)
	ctx := basecontext.NewBaseContext()

	currentUser := "root"
	globalProvider.CurrentSystemUser = currentUser
	globalProvider.RunningUser = currentUser

	if globalProvider.RunningUser == "root" {
		dbLocation := "/etc/parallels-api-service"
		err := helpers.CreateDirIfNotExist(dbLocation)
		if err != nil {
			panic(err)
		}

		globalProvider.JsonDatabase = data.NewJsonDatabase(dbLocation + "/data.json")
		globalProvider.JsonDatabase.Connect(ctx)
		globalProvider.Logger.Info("Running as %s, using %s/data.json file", globalProvider.RunningUser, dbLocation)
		globalProvider.JsonDatabase.Save(ctx)
	} else {
		userHome, err := globalProvider.System.GetUserHome(ctx, currentUser)
		if err != nil {
			panic(err)
		}
		dbLocation := userHome + "/.parallels-api-service"
		err = helpers.CreateDirIfNotExist(dbLocation)
		if err != nil {
			panic(err)
		}

		globalProvider.JsonDatabase = data.NewJsonDatabase(dbLocation + "/data.json")
		globalProvider.JsonDatabase.Connect(ctx)
		globalProvider.Logger.Info("Running as %s, using %s/data.json file", globalProvider.RunningUser, dbLocation)
	}

	key := "00000000-0000-0000-0000-000000000000"
	hid := "XXX00000000000000000000000000000000"

	if shid, err := globalProvider.System.GetUniqueId(ctx); err == nil {
		ctx.LogInfo("Hardware ID: %s", shid)
		hid = shid
	}

	globalProvider.HardwareId = hid
	globalProvider.HardwareSecret = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", key, hid)))
}

func InitServices() {
	// Create a new Services struct and add the DB service
	globalProvider = &ServiceProvider{
		Logger: common.Logger,
	}

	globalProvider.System = system.New()
	globalProvider.System.SetDependencies([]interfaces.Service{})
	globalProvider.Services = append(globalProvider.Services, globalProvider.System)
	globalProvider.GitService = git.New()
	globalProvider.GitService.SetDependencies([]interfaces.Service{globalProvider.System})
	globalProvider.Services = append(globalProvider.Services, globalProvider.GitService)
	globalProvider.PackerService = packer.New()
	globalProvider.PackerService.SetDependencies([]interfaces.Service{globalProvider.System, globalProvider.GitService})
	globalProvider.Services = append(globalProvider.Services, globalProvider.PackerService)
	globalProvider.VagrantService = vagrant.New()
	globalProvider.VagrantService.SetDependencies([]interfaces.Service{globalProvider.System})
	globalProvider.Services = append(globalProvider.Services, globalProvider.VagrantService)
	globalProvider.ParallelsDesktopService = parallelsdesktop.New()
	globalProvider.ParallelsDesktopService.SetDependencies([]interfaces.Service{})
	globalProvider.Services = append(globalProvider.Services, globalProvider.ParallelsDesktopService)
	ctx := basecontext.NewBaseContext()

	currentUser, err := globalProvider.System.GetCurrentUser(ctx)
	if err != nil {
		panic(err)
	}

	globalProvider.CurrentSystemUser = currentUser
	globalProvider.RunningUser = currentUser

	if globalProvider.RunningUser == "root" {
		dbLocation := "/etc/parallels-api-service"
		err := helpers.CreateDirIfNotExist(dbLocation)
		if err != nil {
			panic(err)
		}

		globalProvider.JsonDatabase = data.NewJsonDatabase(dbLocation + "/data.json")
		globalProvider.JsonDatabase.Connect(ctx)
		globalProvider.Logger.Info("Running as %s, using %s/data.json file", globalProvider.RunningUser, dbLocation)
		globalProvider.JsonDatabase.Save(ctx)
	} else {
		userHome, err := globalProvider.System.GetUserHome(ctx, currentUser)
		if err != nil {
			panic(err)
		}
		dbLocation := userHome + "/.parallels-api-service"
		err = helpers.CreateDirIfNotExist(dbLocation)
		if err != nil {
			panic(err)
		}

		globalProvider.JsonDatabase = data.NewJsonDatabase(dbLocation + "/data.json")
		globalProvider.JsonDatabase.Connect(ctx)
		globalProvider.Logger.Info("Running as %s, using %s/data.json file", globalProvider.RunningUser, dbLocation)
	}

	key := "00000000-0000-0000-0000-000000000000"
	hid := "XXX00000000000000000000000000000000"
	if globalProvider.ParallelsDesktopService.Installed() {
		globalProvider.HardwareInfo, err = globalProvider.ParallelsDesktopService.GetInfo()
		if err != nil {
			globalProvider.Logger.Error("Error getting Parallels info")
		}

		if globalProvider.HardwareInfo == nil {
			common.Logger.Error("Error getting Parallels info")
			panic(errors.New("Error getting Parallels Hardware Info"))
		}

		key = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(globalProvider.HardwareInfo.License.Key, "-", ""), "*", ""))
		hid = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(globalProvider.HardwareInfo.HardwareID, "-", ""), "{", ""), "}", ""))
	}

	globalProvider.HardwareId = hid
	globalProvider.HardwareSecret = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", key, hid)))
}

func Get() *ServiceProvider {
	return globalProvider
}

func (p *ServiceProvider) IsParallelsDesktopAvailable() bool {
	if p.ParallelsDesktopService == nil {
		return false
	}
	if !p.ParallelsDesktopService.Installed() || !p.ParallelsDesktopService.IsLicensed() {
		return false
	}

	return true
}

func (p *ServiceProvider) IsGitAvailable() bool {
	if p.GitService == nil {
		return false
	}
	return p.GitService.Installed()
}

func (p *ServiceProvider) IsPackerAvailable() bool {
	if p.PackerService == nil {
		return false
	}
	return p.PackerService.Installed()
}

func (p *ServiceProvider) IsVagrantAvailable() bool {
	if p.VagrantService == nil {
		return false
	}
	return p.VagrantService.Installed()
}

func (p *ServiceProvider) IsSystemAvailable() bool {
	if p.System == nil {
		return false
	}

	return p.System.Installed()
}

func (p *ServiceProvider) InstallAllTools(asUser string, flags map[string]string) {
	if p.IsParallelsDesktopAvailable() {
		p.ParallelsDesktopService.Install(asUser, "latest", flags)
	}
	if p.IsGitAvailable() {
		p.GitService.Install(asUser, "latest", flags)
	}
	if p.IsPackerAvailable() {
		p.PackerService.Install(asUser, "latest", flags)
	}
	if p.IsVagrantAvailable() {
		p.VagrantService.Install(asUser, "latest", flags)
	}
}

func (p *ServiceProvider) UninstallAllTools(asUser string, uninstallDependencies bool, flags map[string]string) {
	if p.IsParallelsDesktopAvailable() {
		p.ParallelsDesktopService.Uninstall(asUser, uninstallDependencies)
	}
	if p.IsPackerAvailable() {
		p.PackerService.Uninstall(asUser, uninstallDependencies)
	}
	if p.IsVagrantAvailable() {
		p.VagrantService.Uninstall(asUser, uninstallDependencies)
	}
}

func GetService[T *any](name string) (T, error) {
	for _, service := range globalProvider.Services {
		if service.Name() == name {
			return service.(T), nil
		}
	}

	return nil, errors.New("Service not found")
}
