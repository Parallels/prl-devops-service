package serviceprovider

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/data"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/serviceprovider/git"
	"github.com/Parallels/pd-api-service/serviceprovider/interfaces"
	"github.com/Parallels/pd-api-service/serviceprovider/packer"
	"github.com/Parallels/pd-api-service/serviceprovider/parallelsdesktop"
	"github.com/Parallels/pd-api-service/serviceprovider/system"
	"github.com/Parallels/pd-api-service/serviceprovider/vagrant"
	sql_database "github.com/Parallels/pd-api-service/sql"

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
	SystemHardwareInfo      *models.SystemHardwareInfo
	CpuType                 string
	HardwareId              string
	HardwareSecret          string
	CurrentSystemUser       string
}

var globalProvider *ServiceProvider

func InitCatalogServices(ctx basecontext.ApiContext) {
	cfg := config.Get()
	globalProvider = &ServiceProvider{
		Logger: common.Logger,
	}

	globalProvider.System = system.New(ctx)
	globalProvider.System.SetDependencies([]interfaces.Service{})
	globalProvider.Services = append(globalProvider.Services, globalProvider.System)

	currentUser := "root"
	globalProvider.CurrentSystemUser = currentUser
	globalProvider.RunningUser = currentUser

	if globalProvider.RunningUser == "root" {
		dbLocation := "/etc/parallels-api-service"
		err := helpers.CreateDirIfNotExist(dbLocation)
		if err != nil {
			panic(err)
		}

		if cfg.DatabaseFolder() != "" {
			globalProvider.JsonDatabase = data.NewJsonDatabase(filepath.Join(cfg.DatabaseFolder(), "/data.json"))
		} else {
			globalProvider.JsonDatabase = data.NewJsonDatabase(filepath.Join(dbLocation, "/data.json"))
		}
		_ = globalProvider.JsonDatabase.Connect(ctx)
		ctx.LogInfof("Running as %s, using %s/data.json file", globalProvider.RunningUser, dbLocation)
		_ = globalProvider.JsonDatabase.Save(ctx)
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

		if cfg.DatabaseFolder() != "" {
			globalProvider.JsonDatabase = data.NewJsonDatabase(filepath.Join(cfg.DatabaseFolder(), "/data.json"))
		} else {
			globalProvider.JsonDatabase = data.NewJsonDatabase(filepath.Join(dbLocation, "/data.json"))
		}
		_ = globalProvider.JsonDatabase.Connect(ctx)
		ctx.LogInfof("Running as %s, using %s/data.json file", globalProvider.RunningUser, dbLocation)
	}

	key := "00000000-0000-0000-0000-000000000000"
	hid := "XXX00000000000000000000000000000000"

	if shid, err := globalProvider.System.GetUniqueId(ctx); err == nil {
		ctx.LogInfof("Hardware ID: %s", shid)
		hid = shid
	}

	globalProvider.HardwareId = hid
	globalProvider.HardwareSecret = getHardwareSecret(key, hid)
	if systemHardwareInfo, err := globalProvider.System.GetHardwareInfo(ctx); err == nil {
		globalProvider.SystemHardwareInfo = systemHardwareInfo
	}
}

func InitServices(ctx basecontext.ApiContext) {
	// Create a new Services struct and add the DB service
	cfg := config.Get()
	globalProvider = &ServiceProvider{
		Logger: common.Logger,
	}

	globalProvider.System = system.New(ctx)
	globalProvider.System.SetDependencies([]interfaces.Service{})
	globalProvider.Services = append(globalProvider.Services, globalProvider.System)
	globalProvider.GitService = git.New(ctx)
	globalProvider.GitService.SetDependencies([]interfaces.Service{globalProvider.System})
	globalProvider.Services = append(globalProvider.Services, globalProvider.GitService)
	globalProvider.PackerService = packer.New(ctx)
	globalProvider.PackerService.SetDependencies([]interfaces.Service{globalProvider.System, globalProvider.GitService})
	globalProvider.Services = append(globalProvider.Services, globalProvider.PackerService)
	globalProvider.VagrantService = vagrant.New(ctx)
	globalProvider.VagrantService.SetDependencies([]interfaces.Service{globalProvider.System})
	globalProvider.Services = append(globalProvider.Services, globalProvider.VagrantService)
	globalProvider.ParallelsDesktopService = parallelsdesktop.New(ctx)
	globalProvider.ParallelsDesktopService.SetDependencies([]interfaces.Service{})
	globalProvider.Services = append(globalProvider.Services, globalProvider.ParallelsDesktopService)

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

		if cfg.DatabaseFolder() != "" {
			globalProvider.JsonDatabase = data.NewJsonDatabase(filepath.Join(cfg.DatabaseFolder(), "/data.json"))
		} else {
			globalProvider.JsonDatabase = data.NewJsonDatabase(filepath.Join(dbLocation, "/data.json"))
		}

		_ = globalProvider.JsonDatabase.Connect(ctx)
		ctx.LogInfof("Running as %s, using %s/data.json file", globalProvider.RunningUser, dbLocation)
		_ = globalProvider.JsonDatabase.Save(ctx)
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

		if cfg.DatabaseFolder() != "" {
			globalProvider.JsonDatabase = data.NewJsonDatabase(filepath.Join(cfg.DatabaseFolder(), "/data.json"))
		} else {
			globalProvider.JsonDatabase = data.NewJsonDatabase(filepath.Join(dbLocation, "/data.json"))
		}

		_ = globalProvider.JsonDatabase.Connect(ctx)
		ctx.LogInfof("Running as %s, using %s/data.json file", globalProvider.RunningUser, dbLocation)
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
	globalProvider.HardwareSecret = getHardwareSecret(key, hid)
	if systemHardwareInfo, err := globalProvider.System.GetHardwareInfo(ctx); err == nil {
		globalProvider.SystemHardwareInfo = systemHardwareInfo
	}
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
		_ = p.ParallelsDesktopService.Install(asUser, "latest", flags)
	}
	if p.IsGitAvailable() {
		_ = p.GitService.Install(asUser, "latest", flags)
	}
	if p.IsPackerAvailable() {
		_ = p.PackerService.Install(asUser, "latest", flags)
	}
	if p.IsVagrantAvailable() {
		_ = p.VagrantService.Install(asUser, "latest", flags)
	}
}

func (p *ServiceProvider) UninstallAllTools(asUser string, uninstallDependencies bool, flags map[string]string) {
	if p.IsParallelsDesktopAvailable() {
		_ = p.ParallelsDesktopService.Uninstall(asUser, uninstallDependencies)
	}
	if p.IsPackerAvailable() {
		_ = p.PackerService.Uninstall(asUser, uninstallDependencies)
	}
	if p.IsVagrantAvailable() {
		_ = p.VagrantService.Uninstall(asUser, uninstallDependencies)
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

func getHardwareSecret(key, hid string) string {
	secretKey := strings.ReplaceAll(key, "-", "")
	secretHid := strings.ReplaceAll(hid, "-", "")
	if len(secretKey) > 12 {
		secretKey = secretKey[:12]
	}
	if len(secretHid) > 12 {
		secretHid = secretHid[:12]
	}

	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s%s", secretKey, secretHid)))
}
