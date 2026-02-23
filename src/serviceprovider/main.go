package serviceprovider

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider/brew"
	"github.com/Parallels/prl-devops-service/serviceprovider/git"
	"github.com/Parallels/prl-devops-service/serviceprovider/interfaces"
	"github.com/Parallels/prl-devops-service/serviceprovider/packer"
	"github.com/Parallels/prl-devops-service/serviceprovider/parallelsdesktop"
	"github.com/Parallels/prl-devops-service/serviceprovider/ssh"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
	"github.com/Parallels/prl-devops-service/serviceprovider/vagrant"
	sql_database "github.com/Parallels/prl-devops-service/sql"

	log "github.com/cjlapao/common-go-logger"
)

type ServiceProvider struct {
	RunningUser             string
	Logger                  *log.LoggerService
	System                  *system.SystemService
	Brew                    *brew.BrewService
	ParallelsDesktopService *parallelsdesktop.ParallelsService
	GitService              *git.GitService
	PackerService           *packer.PackerService
	VagrantService          *vagrant.VagrantService
	SshService              *ssh.SshService
	MySqlService            *sql_database.MySQLService
	JsonDatabase            *data.JsonDatabase
	Services                []interfaces.Service
	HardwareInfo            *models.ParallelsDesktopInfo
	SystemHardwareInfo      *models.SystemHardwareInfo
	CpuType                 string
	HardwareId              string
	HardwareSecret          string
	CurrentSystemUser       string
	License                 string
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
		dbLocation := "/etc/parallels-devops"
		err := helpers.CreateDirIfNotExist(dbLocation)
		if err != nil {
			panic(err)
		}

		if cfg.DatabaseFolder() != "" {
			globalProvider.JsonDatabase = data.NewJsonDatabase(ctx, filepath.Join(cfg.DatabaseFolder(), "/data.json"))
			dbLocation = cfg.DatabaseFolder()
		} else {
			globalProvider.JsonDatabase = data.NewJsonDatabase(ctx, filepath.Join(dbLocation, "/data.json"))
		}

		_ = globalProvider.JsonDatabase.Connect(ctx)
		ctx.LogInfof("Running as %s, using %s/data.json file", globalProvider.RunningUser, dbLocation)
	} else {
		userHome, err := globalProvider.System.GetUserHome(ctx, currentUser)
		if err != nil {
			panic(err)
		}
		dbLocation := userHome + "/.parallels-devops"
		err = helpers.CreateDirIfNotExist(dbLocation)
		if err != nil {
			panic(err)
		}

		if cfg.DatabaseFolder() != "" {
			dbLocation = cfg.DatabaseFolder()
			globalProvider.JsonDatabase = data.NewJsonDatabase(ctx, filepath.Join(cfg.DatabaseFolder(), "/data.json"))
		} else {
			globalProvider.JsonDatabase = data.NewJsonDatabase(ctx, filepath.Join(dbLocation, "/data.json"))
		}
		_ = globalProvider.JsonDatabase.Connect(ctx)
		ctx.LogInfof("Running as %s, using %s/data.json file", globalProvider.RunningUser, dbLocation)
	}

	key := "00000000-0000-0000-0000-000000000000"
	hid := "XXX00000000000000000000000000000000"

	globalProvider.License = key

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

func InitParallelsServices(ctx basecontext.ApiContext) {
	globalProvider = &ServiceProvider{
		Logger: common.Logger,
	}
	globalProvider.System = system.New(ctx)
	globalProvider.System.SetDependencies([]interfaces.Service{})
	globalProvider.Services = append(globalProvider.Services, globalProvider.System)
	globalProvider.ParallelsDesktopService = parallelsdesktop.New(ctx)
	globalProvider.ParallelsDesktopService.SetDependencies([]interfaces.Service{globalProvider.System})
	globalProvider.Services = append(globalProvider.Services, globalProvider.ParallelsDesktopService)

	currentUser, err := globalProvider.System.GetCurrentUser(ctx)
	if err != nil {
		panic(err)
	}

	globalProvider.CurrentSystemUser = currentUser
	globalProvider.RunningUser = currentUser

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

	globalProvider.License = key

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
	globalProvider.Brew = brew.New(ctx)
	globalProvider.Brew.SetDependencies([]interfaces.Service{})
	globalProvider.Services = append(globalProvider.Services, globalProvider.System)
	globalProvider.GitService = git.New(ctx)
	globalProvider.GitService.SetDependencies([]interfaces.Service{globalProvider.System, globalProvider.Brew})
	globalProvider.Services = append(globalProvider.Services, globalProvider.GitService)
	globalProvider.PackerService = packer.New(ctx)
	globalProvider.PackerService.SetDependencies([]interfaces.Service{globalProvider.System, globalProvider.GitService, globalProvider.Brew})
	globalProvider.Services = append(globalProvider.Services, globalProvider.PackerService)
	globalProvider.VagrantService = vagrant.New(ctx)
	globalProvider.VagrantService.SetDependencies([]interfaces.Service{globalProvider.System, globalProvider.Brew})
	globalProvider.Services = append(globalProvider.Services, globalProvider.VagrantService)
	globalProvider.ParallelsDesktopService = parallelsdesktop.New(ctx)
	globalProvider.ParallelsDesktopService.SetDependencies([]interfaces.Service{globalProvider.System, globalProvider.Brew})
	globalProvider.Services = append(globalProvider.Services, globalProvider.ParallelsDesktopService)
	globalProvider.SshService = ssh.New(ctx)
	globalProvider.SshService.SetDependencies([]interfaces.Service{})
	globalProvider.Services = append(globalProvider.Services, globalProvider.SshService)

	currentUser, err := globalProvider.System.GetCurrentUser(ctx)
	if err != nil {
		panic(err)
	}

	globalProvider.CurrentSystemUser = currentUser
	globalProvider.RunningUser = currentUser

	if globalProvider.RunningUser == "root" {
		dbLocation := constants.ServiceDefaultDirectory
		err := helpers.CreateDirIfNotExist(dbLocation)
		if err != nil {
			panic(err)
		}

		if cfg.DatabaseFolder() != "" {
			dbLocation = cfg.DatabaseFolder()
			globalProvider.JsonDatabase = data.NewJsonDatabase(ctx, filepath.Join(cfg.DatabaseFolder(), "/data.json"))
		} else {
			globalProvider.JsonDatabase = data.NewJsonDatabase(ctx, filepath.Join(dbLocation, "/data.json"))
		}

		_ = globalProvider.JsonDatabase.Connect(ctx)
		ctx.LogInfof("Running as %s, using %s/data.json file", globalProvider.RunningUser, dbLocation)
	} else {
		userHome, err := globalProvider.System.GetUserHome(ctx, currentUser)
		if err != nil {
			panic(err)
		}
		dbLocation := userHome + "/.parallels-devops"
		err = helpers.CreateDirIfNotExist(dbLocation)
		if err != nil {
			panic(err)
		}

		if cfg.DatabaseFolder() != "" {
			dbLocation = cfg.DatabaseFolder()
			globalProvider.JsonDatabase = data.NewJsonDatabase(ctx, filepath.Join(cfg.DatabaseFolder(), "/data.json"))
		} else {
			globalProvider.JsonDatabase = data.NewJsonDatabase(ctx, filepath.Join(dbLocation, "/data.json"))
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

	globalProvider.License = key

	globalProvider.HardwareId = hid
	globalProvider.HardwareSecret = getHardwareSecret(key, hid)
	if systemHardwareInfo, err := globalProvider.System.GetHardwareInfo(ctx); err == nil {
		globalProvider.SystemHardwareInfo = systemHardwareInfo
	}
	cacheFolder, err := cfg.CatalogCacheFolder()
	if err != nil {
		ctx.LogErrorf("Error getting cache folder: %v", err)
	} else {
		ctx.LogInfof("Using cache folder: %s", cacheFolder)
	}
	ctx.LogInfof("Using system mode: %s", cfg.Mode())

	if err := TestCacheFolderAccess(ctx); err != nil {
		ctx.LogErrorf("Can't proceed without access to cache folder: %v", err)
		panic(errors.New("Can't proceed without access to cache folder - " + err.Error()))
	} else {
		ctx.LogInfof("Cache folder access test successful")
	}

}

func Get() *ServiceProvider {
	return globalProvider
}

func NewMockProvider() *ServiceProvider {
	globalProvider = &ServiceProvider{
		Logger: common.Logger,
	}

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

func (p *ServiceProvider) IsBrewAvailable() bool {
	if p.Brew == nil {
		return false
	}
	return p.Brew.Installed()
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
	if !p.IsBrewAvailable() {
		_ = p.Brew.Install(asUser, "latest", flags)
	}
	if !p.IsParallelsDesktopAvailable() {
		_ = p.ParallelsDesktopService.Install(asUser, "latest", flags)
	}
	if !p.IsGitAvailable() {
		_ = p.GitService.Install(asUser, "latest", flags)
	}
	if !p.IsPackerAvailable() {
		_ = p.PackerService.Install(asUser, "latest", flags)
	}
	if !p.IsVagrantAvailable() {
		_ = p.VagrantService.Install(asUser, "latest", flags)
	}
}

func (p *ServiceProvider) InstallTool(asUser, tool, version string, flags map[string]string) ToolOperationResult {
	result := ToolOperationResult{
		Type: "install",
	}
	switch tool {
	case "brew":
		result.Name = "brew"
		if !p.IsBrewAvailable() {
			if err := p.Brew.Install(asUser, version, flags); err != nil {
				result.Result = false
				result.Message = err.Error()
			} else {
				result.Result = true
				result.Version = version
			}
		} else {
			result.Result = true
			result.Message = "Brew is already installed"
		}
	case "parallels-desktop", "parallels":
		result.Name = "parallels-desktop"
		if !p.IsParallelsDesktopAvailable() {
			if err := p.ParallelsDesktopService.Install(asUser, version, flags); err != nil {
				result.Result = false
				result.Message = err.Error()
			} else {
				result.Result = true
				result.Version = version
			}
		} else {
			result.Result = true
			result.Message = "Parallels Desktop is already installed"
		}
	case "git":
		result.Name = "git"
		if !p.IsGitAvailable() {
			if err := p.GitService.Install(asUser, version, flags); err != nil {
				result.Result = false
				result.Message = err.Error()
			} else {
				result.Result = true
				result.Version = version
			}
		} else {
			result.Result = true
			result.Message = "Git is already installed"
		}
	case "packer":
		result.Name = "packer"
		if !p.IsPackerAvailable() {
			if err := p.PackerService.Install(asUser, version, flags); err != nil {
				result.Result = false
				result.Message = err.Error()
			} else {
				result.Result = true
				result.Version = version
			}
		} else {
			result.Result = true
			result.Message = "Packer is already installed"
		}
	case "vagrant":
		result.Name = "brew"
		if !p.IsVagrantAvailable() {
			if err := p.VagrantService.Install(asUser, version, flags); err != nil {
				result.Result = false
				result.Message = err.Error()
			} else {
				result.Result = true
				result.Version = version
			}
		} else {
			result.Result = true
			result.Message = "Vagrant is already installed"
		}
	default:
		result.Result = false
		result.Message = fmt.Sprintf("Tool %v not supported", tool)
	}

	return result
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

func (p *ServiceProvider) UninstallTool(asUser, tool string, uninstallDependencies bool, flags map[string]string) ToolOperationResult {
	result := ToolOperationResult{
		Type: "uninstall",
	}

	switch tool {
	case "brew":
		result.Name = "brew"
		if err := p.Brew.Uninstall(asUser, uninstallDependencies); err != nil {
			result.Result = false
			result.Message = err.Error()
		} else {
			result.Result = true
		}
	case "parallels-desktop", "parallels":
		result.Name = "parallels-desktop"
		if err := p.ParallelsDesktopService.Uninstall(asUser, uninstallDependencies); err != nil {
			result.Result = false
			result.Message = err.Error()
		} else {
			result.Result = true
		}
	case "git":
		result.Name = "git"
		if err := p.GitService.Uninstall(asUser, uninstallDependencies); err != nil {
			result.Result = false
			result.Message = err.Error()
		} else {
			result.Result = true
		}
	case "packer":
		result.Name = "packer"
		if err := p.PackerService.Uninstall(asUser, uninstallDependencies); err != nil {
			result.Result = false
			result.Message = err.Error()
		} else {
			result.Result = true
		}
	case "vagrant":
		result.Name = "vagrant"
		if err := p.VagrantService.Uninstall(asUser, uninstallDependencies); err != nil {
			result.Result = false
			result.Message = err.Error()
		} else {
			result.Result = true
		}
	default:
		result.Result = false
		result.Message = fmt.Sprintf("Tool %v not supported", tool)
	}

	return result
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
