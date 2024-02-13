package vagrant

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/serviceprovider/download"
	"github.com/Parallels/pd-api-service/serviceprovider/interfaces"
	"github.com/Parallels/pd-api-service/serviceprovider/system"

	"github.com/cjlapao/common-go/commands"
	"github.com/cjlapao/common-go/helper"
)

var globalVagrantService *VagrantService

type VagrantService struct {
	ctx          basecontext.ApiContext
	executable   string
	installed    bool
	dependencies []interfaces.Service
}

func Get(ctx basecontext.ApiContext) *VagrantService {
	if globalVagrantService != nil {
		return globalVagrantService
	}
	return New(ctx)
}

func New(ctx basecontext.ApiContext) *VagrantService {
	globalVagrantService = &VagrantService{
		ctx: ctx,
	}
	if globalVagrantService.FindPath() == "" {
		ctx.LogWarnf("Running without support for Vagrant")
	} else {
		globalVagrantService.installed = true
	}

	globalVagrantService.SetDependencies([]interfaces.Service{})
	return globalVagrantService
}

func (s *VagrantService) Name() string {
	return "vagrant"
}

func (s *VagrantService) FindPath() string {
	s.ctx.LogInfof("Getting vagrant executable")
	out, err := commands.ExecuteWithNoOutput("which", "vagrant")
	path := strings.ReplaceAll(strings.TrimSpace(out), "\n", "")
	if err != nil || path == "" {
		s.ctx.LogWarnf("Vagrant executable not found, trying to find it in the default locations")
	}

	if path != "" {
		s.executable = path
		s.ctx.LogInfof("Vagrant found at: %s", s.executable)
	} else {
		if _, err := os.Stat("/usr/local/bin/vagrant"); err == nil {
			s.executable = "/usr/local/bin/vagrant"
		} else if _, err := os.Stat("/opt/homebrew/bin/vagrant"); err == nil {
			s.executable = "/opt/homebrew/bin/vagrant"
		} else {
			s.ctx.LogWarnf("Vagrant executable not found, trying to install it")
			return s.executable
		}

		s.ctx.LogInfof("Vagrant found at: %s", s.executable)
	}

	return s.executable
}

func (s *VagrantService) Version() string {
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"version"},
	}

	stdout, _, _, err := helpers.ExecuteWithOutput(cmd)
	if err != nil {
		return "unknown"
	}

	return strings.ReplaceAll(strings.TrimSpace(strings.ReplaceAll(stdout, "Vagrant ", "")), "\n", "")
}

func (s *VagrantService) Install(asUser, version string, flags map[string]string) error {
	if s.installed {
		s.ctx.LogInfof("%s already installed", s.Name())
		return nil
	}

	// Installing service dependency
	if err := s.installDependencies(asUser, flags); err != nil {
		return err
	}

	var cmd helpers.Command
	switch asUser {
	case "":
		cmd = helpers.Command{
			Command: "brew",
		}
	default:
		cmd = helpers.Command{
			Command: "sudo",
			Args:    []string{"-u", asUser, "brew"},
		}
	}

	if version == "" || version == "latest" {
		cmd.Args = append(cmd.Args, "install", "hashicorp-vagrant")
	} else {
		cmd.Args = append(cmd.Args, "install", "hashicorp-vagrant@"+version)
	}

	s.ctx.LogInfof("Installing %s with command: %v", s.Name(), cmd.String())
	if _, err := helpers.ExecuteWithNoOutput(cmd); err != nil {
		return err
	}

	s.installed = true
	s.ctx.LogInfof("Installing %s plugins", s.Name())
	if err := s.InstallParallelsDesktopPlugin(asUser); err != nil {
		return err
	}

	return nil
}

func (s *VagrantService) installDependencies(asUser string, flags map[string]string) error {
	// Installing service dependency
	if s.dependencies != nil {
		for _, dependency := range s.dependencies {
			if dependency == nil {
				return errors.New("Dependency is nil")
			}
			s.ctx.LogInfof("Installing dependency %s for %s", dependency.Name(), s.Name())
			if err := dependency.Install(asUser, "latest", flags); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *VagrantService) Uninstall(asUser string, uninstallDependencies bool) error {
	if s.installed {
		s.ctx.LogInfof("Uninstalling %s", s.Name())
		var cmd helpers.Command
		if asUser == "" {
			cmd = helpers.Command{
				Command: "brew",
				Args:    []string{"uninstall", "hashicorp-vagrant"},
			}
		} else {
			cmd = helpers.Command{
				Command: "sudo",
				Args:    []string{"-u", asUser, "brew", "uninstall", "hashicorp-vagrant"},
			}
		}

		_, err := helpers.ExecuteWithNoOutput(cmd)
		if err != nil {
			return err
		}
	}

	if uninstallDependencies {
		// Uninstall service dependency
		if err := s.uninstallDependencies(asUser); err != nil {
			return err
		}
	}

	s.installed = false
	return nil
}

func (s *VagrantService) uninstallDependencies(asUser string) error {
	// Uninstall service dependency
	if s.dependencies != nil {
		for _, dependency := range s.dependencies {
			if dependency == nil {
				continue
			}
			s.ctx.LogInfof("Uninstalling dependency %s for %s", dependency.Name(), s.Name())
			if err := dependency.Uninstall(asUser, true); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *VagrantService) Dependencies() []interfaces.Service {
	if s.dependencies == nil {
		s.dependencies = []interfaces.Service{}
	}
	return s.dependencies
}

func (s *VagrantService) SetDependencies(dependencies []interfaces.Service) {
	s.dependencies = dependencies
}

func (s *VagrantService) Installed() bool {
	return s.installed && s.executable != ""
}

func (s *VagrantService) InstallParallelsDesktopPlugin(asUser string) error {
	if s.installed {
		s.ctx.LogInfof("Updating Parallels Desktop Plugin %s", s.Name())
		var cmd helpers.Command
		if asUser == "" {
			cmd = helpers.Command{
				Command: "vagrant",
				Args:    []string{"plugin", "install", "vagrant-parallels"},
			}
		} else {
			cmd = helpers.Command{
				Command: "sudo",
				Args:    []string{"-u", asUser, "vagrant", "plugin", "install", "vagrant-parallels"},
			}
		}

		_, err := helpers.ExecuteWithNoOutput(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *VagrantService) UpdatePlugins(asUser string) error {
	if s.installed {
		s.ctx.LogInfof("Updating Parallels Desktop Plugin %s", s.Name())
		var cmd helpers.Command
		if asUser == "" {
			cmd = helpers.Command{
				Command: "vagrant",
				Args:    []string{"plugin", "update"},
			}
		} else {
			cmd = helpers.Command{
				Command: "sudo",
				Args:    []string{"-u", asUser, "vagrant", "plugin", "update"},
			}
		}

		_, err := helpers.ExecuteWithNoOutput(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *VagrantService) updateVagrantFile(ctx basecontext.ApiContext, filePath string, machineName string) error {
	if !helper.FileExists(filePath) {
		return errors.Newf("Vagrant file %v does not exist", filePath)
	}
	if !strings.HasSuffix(filePath, "Vagrantfile") {
		filePath = filepath.Join(filePath, "Vagrantfile")
	}

	vagrantFile, err := LoadVagrantFile(ctx, filePath)
	if err != nil {
		return err
	}

	if err := helper.CopyFile(filePath, filePath+".bak"); err != nil {
		return err
	}

	if err := helper.CopyFile(filePath, filePath+".tmp"); err != nil {
		return err
	}

	blocks := vagrantFile.GetConfigBlock("parallels")
	if len(blocks) == 0 {
		ctx.LogInfof("No parallels block found in vagrant file, adding it")
		parallelsBlock := VagrantConfigBlock{
			Name:         "parallels",
			Type:         "config.vm.provider",
			VariableName: "prl",
		}
		parallelsBlock.SetContentVariable("name", machineName)
		vagrantFile.Root.Children = append(vagrantFile.Root.Children, &parallelsBlock)
	} else {
		block := blocks[len(blocks)-1]
		if block.GetContentVariable("name") != machineName {
			block.SetContentVariable("name", machineName)
		}
	}

	if err := vagrantFile.Save(); err != nil {
		return err
	}

	return nil
}

func (s *VagrantService) getVagrantFolderPath(ctx basecontext.ApiContext, request models.CreateVagrantMachineRequest) (string, error) {
	system := system.Get()
	rootDir, err := system.GetUserHome(ctx, request.Owner)
	if err != nil {
		return "", err
	}
	userId, err := system.GetUserId(ctx, request.Owner)
	if err != nil {
		return "", err
	}

	vagrantFileFolderName := ""
	if request.Name != "" {
		vagrantFileFolderName = helpers.NormalizeString(request.Name)
	} else if request.Box != "" {
		vagrantFileFolderName = helpers.NormalizeString(request.Box)
	} else if request.VagrantFilePath != "" {
		vagrantFileFolderName = helpers.NormalizeString(filepath.Base(request.VagrantFilePath))
	} else {
		return "", errors.NewWithCode("Box or Name must be provided", 500)
	}

	vagrantFileFolder := filepath.Join(rootDir, fmt.Sprintf("vagrant_%s", vagrantFileFolderName))
	if request.VagrantFilePath != "" {
		if !strings.HasPrefix(request.VagrantFilePath, "http://") || !strings.HasPrefix(request.VagrantFilePath, "https://") {
			return filepath.Dir(request.VagrantFilePath), nil
		} else {
			destinationFilePath := filepath.Join(vagrantFileFolder, "Vagrantfile")
			downloadService := download.NewDownloadService()
			if err := downloadService.DownloadFile(request.VagrantFilePath, nil, destinationFilePath); err != nil {
				return "", err
			}

			return filepath.Dir(destinationFilePath), nil
		}
	}

	if err := helpers.CreateDirIfNotExist(vagrantFileFolder); err != nil {
		return "", err
	}

	if err := os.Chown(vagrantFileFolder, userId, -1); err != nil {
		return "", err
	}

	return vagrantFileFolder, nil
}

func (s *VagrantService) getVagrantFilePath(ctx basecontext.ApiContext, request models.CreateVagrantMachineRequest) (string, error) {
	vagrantFileFolder, err := s.getVagrantFolderPath(ctx, request)
	if err != nil {
		return "", err
	}

	vagrantFilePath := filepath.Join(vagrantFileFolder, "Vagrantfile")

	if helper.FileExists(vagrantFilePath) {
		if err := helper.DeleteFile(vagrantFilePath); err != nil {
			return "", err
		}
	}

	return vagrantFilePath, nil
}

func (s *VagrantService) GenerateVagrantFile(ctx basecontext.ApiContext, request models.CreateVagrantMachineRequest) (string, error) {
	vagrantFilePath, err := s.getVagrantFilePath(ctx, request)
	if err != nil {
		return "", err
	}

	file := NewVagrantFile(ctx, vagrantFilePath)
	file.Root.SetContentVariable("vm.box", request.Box)
	if request.Version != "" {
		file.Root.SetContentVariable("vm.box_version", request.Version)
	}

	block := file.Root.NewBlock("config.vm.provider", "parallels", "prl")
	block.SetContentVariable("name", request.Name)
	if request.CustomParallelsConfig != "" {
		lines := strings.Split(request.CustomParallelsConfig, "\n")
		block.Content = append(block.Content, lines...)
	}

	if request.CustomVagrantConfig != "" {
		lines := strings.Split(request.CustomVagrantConfig, "\n")
		file.Root.Content = append(file.Root.Content, lines...)
	}

	file.Refresh()

	if err := file.Save(); err != nil {
		return "", err
	}

	return file.String(), nil
}

func (s *VagrantService) Init(ctx basecontext.ApiContext, request models.CreateVagrantMachineRequest) error {
	vagrantFileFolder, err := s.getVagrantFolderPath(ctx, request)
	if err != nil {
		return err
	}

	if content, err := s.GenerateVagrantFile(ctx, request); err != nil {
		ctx.LogErrorf("Error generating vagrant file: %v", err)
		ctx.LogErrorf("Vagrant file content: %v", content)
		return err
	}

	cmd := helpers.Command{
		Command:          "sudo",
		WorkingDirectory: vagrantFileFolder,
		Args:             make([]string, 0),
	}

	if request.Owner != "" {
		cmd.Args = append(cmd.Args, "-u", request.Owner, s.executable)
	} else {
		cmd.Args = append(cmd.Args, s.executable)
	}

	cmd.Args = append(cmd.Args, "init", request.Box)

	ctx.LogInfof("Initializing vagrant folder with command: %v", cmd.String())
	stdout, _, _, err := helpers.ExecuteAndWatch(cmd)
	if err != nil {
		println(stdout)
		buildError := errors.Newf("There was an error init vagrant folder %v, error: %v", vagrantFileFolder, err.Error())
		return buildError
	}

	return nil
}

func (s *VagrantService) Up(ctx basecontext.ApiContext, request models.CreateVagrantMachineRequest) error {
	vagrantFileFolder, err := s.getVagrantFolderPath(ctx, request)
	if err != nil {
		return err
	}

	cmd := helpers.Command{
		Command:          "sudo",
		WorkingDirectory: vagrantFileFolder,
		Args:             make([]string, 0),
	}

	if request.Owner != "" {
		cmd.Args = append(cmd.Args, "-u", request.Owner, s.executable)
	} else {
		cmd.Args = append(cmd.Args, s.executable)
	}

	if request.VagrantFilePath != "" {
		if err := s.updateVagrantFile(ctx, request.VagrantFilePath, request.Name); err != nil {
			return err
		}
	}

	cmd.Args = append(cmd.Args, "up", "--no-tty", "--machine-readable")
	ctx.LogInfof("Bringing vagrant box %s up with command: %v", request.Box, cmd.String())
	ctx.LogInfof(cmd.String())

	if _, _, _, err = helpers.ExecuteAndWatch(cmd); err != nil {
		buildError := errors.Newf("There was an error bringing the vagrant box up on folder %v, error: %v", vagrantFileFolder, err.Error())
		return buildError
	}

	// Cleaning any backup files we had to create
	if helper.FileExists(filepath.Join(vagrantFileFolder, "Vagrantfile.tmp")) {
		if err := helper.DeleteFile(filepath.Join(vagrantFileFolder, "Vagrantfile")); err != nil {
			return err
		}
		if err := helper.CopyFile(filepath.Join(vagrantFileFolder, "Vagrantfile.tmp"), filepath.Join(vagrantFileFolder, "Vagrantfile")); err != nil {
			return err
		}
		if err := helper.DeleteFile(filepath.Join(vagrantFileFolder, "Vagrantfile.tmp")); err != nil {
			return err
		}
	}

	return nil
}
