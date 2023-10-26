package service_provider

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/data"
	"Parallels/pd-api-service/errors"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/service_provider/git"
	"Parallels/pd-api-service/service_provider/interfaces"
	"Parallels/pd-api-service/service_provider/packer"
	"Parallels/pd-api-service/service_provider/parallels_desktop"
	"Parallels/pd-api-service/service_provider/system"
	"Parallels/pd-api-service/service_provider/vagrant"
	sql_database "Parallels/pd-api-service/sql"
	"encoding/base64"
	"fmt"
	"strings"

	log "github.com/cjlapao/common-go-logger"
)

type ServiceProvider struct {
	RunningUser      string
	Logger           *log.LoggerService
	System           *system.SystemService
	ParallelsService *parallels_desktop.ParallelsService
	GitService       *git.GitService
	PackerService    *packer.PackerService
	VagrantService   *vagrant.VagrantService
	MySqlService     *sql_database.MySQLService
	JsonDatabase     *data.JsonDatabase
	Services         []interfaces.Service
	HardwareInfo     *models.ParallelsDesktopInfo
	HardwareId       string
	HardwareSecret   string
}

var globalProvider *ServiceProvider

func InitServices() {
	// Connect to the SQL server
	// dbService, err := initDatabase()
	// if err != nil {
	// 	panic(err)
	// }

	// Create a new Services struct and add the DB service
	globalProvider = &ServiceProvider{
		Logger: common.Logger,
	}
	stdout, err := helpers.ExecuteWithNoOutput(helpers.Command{
		Command: "whoami",
	})
	if err != nil {
		panic(err)
	}
	globalProvider.RunningUser = strings.ReplaceAll(strings.TrimSpace(stdout), "\n", "")
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
	globalProvider.ParallelsService = parallels_desktop.New()
	globalProvider.ParallelsService.SetDependencies([]interfaces.Service{})
	globalProvider.Services = append(globalProvider.Services, globalProvider.ParallelsService)

	if globalProvider.RunningUser == "root" {
		dbLocation := "/etc/parallels-api-service"
		err := helpers.CreateDirIfNotExist(dbLocation)
		if err != nil {
			panic(err)
		}

		globalProvider.JsonDatabase = data.NewJsonDatabase(dbLocation + "/data.json")
		globalProvider.JsonDatabase.Connect()
		globalProvider.Logger.Info("Running as %s, using %s/data.json file", globalProvider.RunningUser, dbLocation)
	} else {
		dbLocation := "/Users/" + globalProvider.RunningUser + "/.parallels-api-service"
		err := helpers.CreateDirIfNotExist(dbLocation)
		if err != nil {
			panic(err)
		}

		globalProvider.JsonDatabase = data.NewJsonDatabase(dbLocation + "/data.json")
		globalProvider.JsonDatabase.Connect()
		globalProvider.Logger.Info("Running as %s, using %s/data.json file", globalProvider.RunningUser, dbLocation)
	}

	if globalProvider.ParallelsService.Installed() {
		globalProvider.HardwareInfo, err = globalProvider.ParallelsService.GetInfo()
		if err != nil {
			globalProvider.Logger.Error("Error getting Parallels info")
		}

		if globalProvider.HardwareInfo == nil {
			common.Logger.Error("Error getting Parallels info")
			panic(errors.New("Error getting Parallels Hardware Info"))
		}

		key := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(globalProvider.HardwareInfo.License.Key, "-", ""), "*", ""))
		hid := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(globalProvider.HardwareInfo.HardwareID, "-", ""), "{", ""), "}", ""))
		globalProvider.HardwareId = hid
		globalProvider.HardwareSecret = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", key, hid)))
	} else {
		globalProvider.HardwareId = "00000000-0000-0000-0000-000000000000"
		globalProvider.HardwareSecret = "XXX00000000000000000000000000000000"
	}
}

func Get() *ServiceProvider {
	return globalProvider
}

func (p *ServiceProvider) IsParallelsDesktopAvailable() bool {
	if p.ParallelsService == nil {
		return false
	}
	if !p.ParallelsService.Installed() || !p.ParallelsService.IsLicensed() {
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
		p.ParallelsService.Install(asUser, "latest", flags)
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
		p.ParallelsService.Uninstall(asUser, uninstallDependencies)
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

func initDatabase() (*sql_database.MySQLService, error) {
	service := sql_database.MySQLService{}
	_, err := service.Connect()
	if err != nil {
		return nil, err
	}

	return &service, nil
}
