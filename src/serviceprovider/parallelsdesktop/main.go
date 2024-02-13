package parallelsdesktop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/data"
	data_models "github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/serviceprovider/git"
	"github.com/Parallels/pd-api-service/serviceprovider/interfaces"
	"github.com/Parallels/pd-api-service/serviceprovider/packer"
	"github.com/Parallels/pd-api-service/serviceprovider/system"

	"github.com/cjlapao/common-go/commands"
	"github.com/cjlapao/common-go/helper"
)

var (
	globalParallelsService *ParallelsService
	logger                 = common.Logger
)

type ParallelsService struct {
	ctx              basecontext.ApiContext
	executable       string
	serverExecutable string
	Info             *models.ParallelsDesktopInfo
	isLicensed       bool
	installed        bool
	dependencies     []interfaces.Service
}

func Get(ctx basecontext.ApiContext) *ParallelsService {
	if globalParallelsService != nil {
		return globalParallelsService
	}
	return New(ctx)
}

func New(ctx basecontext.ApiContext) *ParallelsService {
	globalParallelsService = &ParallelsService{
		ctx: ctx,
	}

	if globalParallelsService.FindPath() == "" {
		ctx.LogWarnf("Running without support for Parallels Desktop")
	} else {
		globalParallelsService.installed = true
	}

	globalParallelsService.SetDependencies([]interfaces.Service{})
	return globalParallelsService
}

func (s *ParallelsService) Name() string {
	return "parallels_desktop"
}

func (s *ParallelsService) FindPath() string {
	s.ctx.LogInfof("Getting prlctl executable")
	out, err := commands.ExecuteWithNoOutput("which", "prlctl")
	path := strings.ReplaceAll(strings.TrimSpace(out), "\n", "")
	if err != nil || path == "" {
		s.ctx.LogWarnf("Parallels Desktop CLI executable not found, trying to find it in the default locations")
	}

	if path != "" {
		s.executable = path
		s.serverExecutable = strings.ReplaceAll(path, "prlctl", "prlsrvctl")
		s.ctx.LogInfof("Parallels Desktop CLI found at: %s", s.executable)
	} else {
		if _, err := os.Stat("/usr/bin/prlctl"); err == nil {
			s.executable = "/usr/bin/prlctl"
			s.serverExecutable = "/usr/bin/prlsrvctl"
			if err := os.Setenv("PATH", os.Getenv("PATH")+":/usr/bin"); err != nil {
				s.ctx.LogWarnf("Error setting PATH environment variable: %v", err)
			}
		} else if _, err := os.Stat("/usr/local/bin/prlctl"); err == nil {
			s.executable = "/usr/local/bin/prlctl"
			s.serverExecutable = "/usr/local/bin/prlsrvctl"
			if err := os.Setenv("PATH", os.Getenv("PATH")+":/usr/local/bin"); err != nil {
				s.ctx.LogWarnf("Error setting PATH environment variable: %v", err)
			}
		} else {
			s.ctx.LogWarnf("Parallels Desktop CLI executable not found, trying to install it")
			return s.executable
		}

		s.ctx.LogInfof("Parallels Desktop CLI found at: %s", s.executable)
	}

	return s.executable
}

func (s *ParallelsService) Version() string {
	cmd := helpers.Command{
		Command: s.executable,
		Args:    []string{"--version"},
	}

	stdout, _, _, err := helpers.ExecuteWithOutput(cmd)
	if err != nil {
		return "unknown"
	}

	v := strings.ReplaceAll(strings.TrimSpace(strings.ReplaceAll(stdout, "prlctl version  ", "")), "\n", "")
	vParts := strings.Split(v, " ")
	if len(vParts) > 0 {
		return vParts[0]
	} else {
		return v
	}
}

func (s *ParallelsService) Install(asUser, version string, flags map[string]string) error {
	if s.installed {
		s.ctx.LogInfof("%s already installed", s.Name())
	} else {

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

		var cmd helpers.Command
		if asUser == "" {
			cmd = helpers.Command{
				Command: "brew",
			}
		} else {
			cmd = helpers.Command{
				Command: "sudo",
				Args:    []string{"-u", asUser, "brew"},
			}
		}

		if version == "" || version == "latest" {
			cmd.Args = append(cmd.Args, "install", "parallels")
		} else {
			cmd.Args = append(cmd.Args, "install", "parallels@"+version)
		}

		s.ctx.LogInfof("Installing %s with command: %v", s.Name(), cmd.String())
		_, err := helpers.ExecuteWithNoOutput(cmd)
		if err != nil {
			return err
		}
		s.installed = true
	}

	license := ""
	username := ""
	password := ""

	for flag, value := range flags {
		switch flag {
		case "license":
			license = value
		case "my_account_username":
			username = value
		case "my_account_password":
			password = value
		}
	}

	if license != "" {
		s.ctx.LogInfof("Activating Parallels Desktop with license %s", license)
		if err := s.InstallLicense(license, username, password); err != nil {
			return err
		}

		if _, err := s.GetInfo(); err != nil {
			return err
		}
	}

	return nil
}

func (s *ParallelsService) Uninstall(asUser string, uninstallDependencies bool) error {
	if s.installed {
		s.ctx.LogInfof("Uninstalling %s", s.Name())

		var cmd helpers.Command
		if asUser == "" {
			cmd = helpers.Command{
				Command: "brew",
				Args:    []string{"uninstall", "parallels"},
			}
		} else {
			cmd = helpers.Command{
				Command: "sudo",
				Args:    []string{"-u", asUser, "brew", "uninstall", "parallels"},
			}
		}

		_, err := helpers.ExecuteWithNoOutput(cmd)
		if err != nil {
			return err
		}
	}

	if err := s.DeactivateLicense(); err != nil {
		return err
	}

	if uninstallDependencies {
		// Uninstall service dependency
		if s.dependencies != nil {
			for _, dependency := range s.dependencies {
				if dependency == nil {
					continue
				}
				s.ctx.LogInfof("Uninstalling dependency %s for %s", dependency.Name(), s.Name())
				if err := dependency.Uninstall(asUser, uninstallDependencies); err != nil {
					return err
				}
			}
		}
	}

	s.installed = false
	return nil
}

func (s *ParallelsService) Dependencies() []interfaces.Service {
	if s.dependencies == nil {
		s.dependencies = []interfaces.Service{}
	}
	return s.dependencies
}

func (s *ParallelsService) SetDependencies(dependencies []interfaces.Service) {
	s.dependencies = dependencies
}

func (s *ParallelsService) Installed() bool {
	return s.installed && s.executable != "" && s.serverExecutable != ""
}

func (s *ParallelsService) IsLicensed() bool {
	return s.isLicensed
}

func (s *ParallelsService) GetVms(ctx basecontext.ApiContext, filter string) ([]models.ParallelsVM, error) {
	var systemMachines []models.ParallelsVM
	users, err := system.Get().GetSystemUsers(ctx)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.ErrNoSystemUserFound()
	}

	for _, user := range users {
		ctx.LogInfof("Getting VMs for user: %s", user.Username)
		var userMachines []models.ParallelsVM
		stdout, err := commands.ExecuteWithNoOutput("sudo", "-u", user.Username, s.executable, "list", "-a", "-i", "--json")
		if err != nil {
			continue
		}

		err = json.Unmarshal([]byte(stdout), &userMachines)
		if err != nil {
			continue
		}

		for _, machine := range userMachines {
			found := false
			for _, globalMachine := range systemMachines {
				if strings.EqualFold(machine.ID, globalMachine.ID) {
					found = true
					break
				}
			}
			if !found {
				machine.User = user.Username
				systemMachines = append(systemMachines, machine)
			}
		}
	}

	dbFilter, err := data.ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := data.FilterByProperty(systemMachines, dbFilter)
	if err != nil {
		return nil, err
	}

	return filteredData, nil
}

func (s *ParallelsService) GetVm(ctx basecontext.ApiContext, id string) (*models.ParallelsVM, error) {
	vm, err := s.findVm(ctx, id)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func (s *ParallelsService) SetVmState(ctx basecontext.ApiContext, id string, desiredState ParallelsVirtualMachineDesiredState) error {
	vm, err := s.findVm(ctx, id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.ErrNoVirtualMachineFound(id)
	}

	if vm.User == "" {
		vm.User = "root"
	}

	switch desiredState {
	case ParallelsVirtualMachineDesiredStateStart:
		if vm.State == ParallelsVirtualMachineStateRunning.String() {
			return nil
		}
		if vm.State != ParallelsVirtualMachineStateStopped.String() {
			return errors.New("VM is not stopped")
		}
	case ParallelsVirtualMachineDesiredStateStop:
		if vm.State == ParallelsVirtualMachineStateStopped.String() {
			return nil
		}
		if vm.State != ParallelsVirtualMachineStateRunning.String() {
			return errors.New("VM is not running")
		}
	case ParallelsVirtualMachineDesiredStatePause:
		if vm.State == ParallelsVirtualMachineStatePaused.String() {
			return nil
		}
		if vm.State != ParallelsVirtualMachineStateRunning.String() {
			return errors.New("VM is not running")
		}
	case ParallelsVirtualMachineDesiredStateSuspend:
		if vm.State == ParallelsVirtualMachineStateSuspended.String() {
			return nil
		}
		if vm.State != ParallelsVirtualMachineStateRunning.String() {
			return errors.New("VM is not running")
		}
	case ParallelsVirtualMachineDesiredStateResume:
		if vm.State != ParallelsVirtualMachineStatePaused.String() &&
			vm.State != ParallelsVirtualMachineStateSuspended.String() {
			return errors.New("VM is not paused or suspended")
		}
	case ParallelsVirtualMachineDesiredStateReset:
		if vm.State == ParallelsVirtualMachineStateStopped.String() {
			return nil
		}
	case ParallelsVirtualMachineDesiredStateRestart:
		if vm.State == ParallelsVirtualMachineStateStopped.String() {
			return nil
		}
		if vm.State != ParallelsVirtualMachineStateRunning.String() {
			return errors.New("VM is not running")
		}
	default:
		return errors.New("Invalid desired state")
	}

	_, err = commands.ExecuteWithNoOutput("sudo", "-u", vm.User, s.executable, desiredState.String(), id)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) StartVm(ctx basecontext.ApiContext, id string) error {
	return s.SetVmState(ctx, id, ParallelsVirtualMachineDesiredStateStart)
}

func (s *ParallelsService) StopVm(ctx basecontext.ApiContext, id string) error {
	return s.SetVmState(ctx, id, ParallelsVirtualMachineDesiredStateStop)
}

func (s *ParallelsService) RestartVm(ctx basecontext.ApiContext, id string) error {
	return s.SetVmState(ctx, id, ParallelsVirtualMachineDesiredStateRestart)
}

func (s *ParallelsService) SuspendVm(ctx basecontext.ApiContext, id string) error {
	return s.SetVmState(ctx, id, ParallelsVirtualMachineDesiredStateSuspend)
}

func (s *ParallelsService) ResumeVm(ctx basecontext.ApiContext, id string) error {
	return s.SetVmState(ctx, id, ParallelsVirtualMachineDesiredStateResume)
}

func (s *ParallelsService) ResetVm(ctx basecontext.ApiContext, id string) error {
	return s.SetVmState(ctx, id, ParallelsVirtualMachineDesiredStateReset)
}

func (s *ParallelsService) PauseVm(ctx basecontext.ApiContext, id string) error {
	return s.SetVmState(ctx, id, ParallelsVirtualMachineDesiredStatePause)
}

func (s *ParallelsService) DeleteVm(ctx basecontext.ApiContext, id string) error {
	vm, err := s.findVm(ctx, id)
	if err != nil {
		return err
	}

	if vm == nil {
		return errors.Newf("VM with id %s was not found", id)
	}

	if vm.State != "stopped" {
		return errors.New("VM is not stopped")
	}

	_, err = commands.ExecuteWithNoOutput("sudo", "-u", vm.User, s.executable, "delete", id)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) VmStatus(ctx basecontext.ApiContext, id string) (string, error) {
	vm, err := s.findVm(ctx, id)
	if err != nil {
		return "", err
	}
	if vm == nil {
		return "", errors.New("VM not found")
	}

	output, err := commands.ExecuteWithNoOutput("sudo", "-u", vm.User, s.executable, "status", id)
	if err != nil {
		return "", err
	}

	statusParts := strings.Split(output, " ")
	if len(statusParts) != 4 {
		return "", errors.New("Invalid status output")
	}

	return strings.ReplaceAll(statusParts[3], "\n", ""), nil
}

func (s *ParallelsService) RegisterVm(ctx basecontext.ApiContext, r models.RegisterVirtualMachineRequest) error {
	if r.Uuid != "" {
		vm, err := s.findVm(ctx, r.Uuid)
		if err != nil {
			return err
		}
		if vm != nil {
			return errors.Newf("VM with UUID %s already exists", r.Uuid)
		}
	}

	if r.MachineName != "" {
		vm, err := s.findVm(ctx, r.MachineName)
		if err != nil && errors.GetSystemErrorCode(err) != 404 {
			return err
		}
		if vm != nil {
			return errors.Newf("VM with name %s already exists", r.MachineName)
		}
	}

	if err := s.ReplaceMachineNameInConfigPvs(r.Path, r.MachineName); err != nil {
		return err
	}

	cmd := helpers.Command{
		Command: "sudo",
		Args:    make([]string, 0),
	}

	if r.Owner != "" && r.Owner != "root" {
		cmd.Args = append(cmd.Args, "-u", r.Owner)
	}

	cmd.Args = append(cmd.Args, s.executable, "register", r.Path)
	if r.Uuid != "" {
		cmd.Args = append(cmd.Args, "--uuid", r.Uuid)
	}
	if r.RegenerateSourceUuid {
		cmd.Args = append(cmd.Args, "--regenerate-src-uuid")
	}
	if r.Force {
		cmd.Args = append(cmd.Args, "--force")
	}
	if r.DelayApplyingRestrictions {
		cmd.Args = append(cmd.Args, "--delay-applying-restrictions")
	}

	ctx.LogDebugf("Executing command: %s", cmd.String())
	_, err := helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) UnregisterVm(ctx basecontext.ApiContext, r models.UnregisterVirtualMachineRequest) error {
	vm, err := s.findVm(ctx, r.ID)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.ErrNoVirtualMachineFound(r.ID)
	}
	r.Owner = vm.User

	cmd := helpers.Command{
		Command: "sudo",
		Args:    make([]string, 0),
	}

	if r.Owner != "" && r.Owner != "root" {
		cmd.Args = append(cmd.Args, "-u", r.Owner)
	}

	cmd.Args = append(cmd.Args, s.executable, "unregister", r.ID)
	if r.CleanSourceUuid {
		cmd.Args = append(cmd.Args, "--clean-src-uuid")
	}

	ctx.LogInfof(cmd.String())
	_, err = helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		return errors.NewFromErrorf(err, "Error unregistering VM %s", r.ID)
	}

	return nil
}

func (s *ParallelsService) RenameVm(ctx basecontext.ApiContext, r models.RenameVirtualMachineRequest) error {
	vm, err := s.findVm(ctx, r.GetId())
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.New("VM not found")
	}

	cmd := helpers.Command{
		Command: "sudo",
		Args:    make([]string, 0),
	}

	if vm.User != "" && vm.User != "root" {
		cmd.Args = append(cmd.Args, "-u", vm.User)
	}

	cmd.Args = append(cmd.Args, s.executable, "set", r.GetId(), "--name", r.NewName)
	if r.Description != "" {
		cmd.Args = append(cmd.Args, "--description", r.Description)
	}

	ctx.LogInfof(cmd.String())
	_, err = helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) PackVm(ctx basecontext.ApiContext, idOrName string) error {
	vm, err := s.findVm(ctx, idOrName)
	if err != nil {
		return err
	}
	if vm != nil {
		return errors.Newf("VM with ID %s was not found", idOrName)
	}

	cmd := helpers.Command{
		Command: "sudo",
		Args:    make([]string, 0),
	}

	if vm.User != "" && vm.User != "root" {
		cmd.Args = append(cmd.Args, "-u", vm.User)
	}

	cmd.Args = append(cmd.Args, s.executable, "pack", vm.ID)
	_, err = helpers.ExecuteWithNoOutput(cmd)

	return err
}

func (s *ParallelsService) UnpackVm(ctx basecontext.ApiContext, idOrName string) error {
	vm, err := s.findVm(ctx, idOrName)
	if err != nil {
		return err
	}
	if vm != nil {
		return errors.Newf("VM with ID %s was not found", idOrName)
	}

	cmd := helpers.Command{
		Command: "sudo",
		Args:    make([]string, 0),
	}

	if vm.User != "" && vm.User != "root" {
		cmd.Args = append(cmd.Args, "-u", vm.User)
	}

	cmd.Args = append(cmd.Args, s.executable, "unpack", vm.ID)
	_, err = helpers.ExecuteWithNoOutput(cmd)

	return err
}

func (s *ParallelsService) GetInfo() (*models.ParallelsDesktopInfo, error) {
	if s.Info != nil {
		return s.Info, nil
	}

	stdout, err := helpers.ExecuteWithNoOutput(helpers.Command{
		Command: s.serverExecutable,
		Args:    []string{"info", "--json"},
	})
	if err != nil {
		return nil, err
	}

	var info models.ParallelsDesktopInfo
	err = json.Unmarshal([]byte(stdout), &info)
	if err != nil {
		return nil, err
	}

	s.Info = &info
	if info.License.State != "valid" {
		s.ctx.LogErrorf("Parallels license is not active")
	} else {
		s.isLicensed = true
	}

	return s.Info, nil
}

func (s *ParallelsService) ConfigureVm(ctx basecontext.ApiContext, id string, setOperations *models.VirtualMachineConfigRequest) error {
	vm, err := s.findVm(ctx, id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.ErrNoVirtualMachineFound(id)
	}

	for _, op := range setOperations.Operations {
		op.Owner = vm.User
		switch op.Group {
		case "state":
			ctx.LogInfof("Setting machine state to %s", op.Operation)
			if err := s.SetVmState(ctx, vm.ID, ParallelsVirtualMachineDesiredStateFromString(op.Operation)); err != nil {
				op.Error = err
			}
		case "machine":
			ctx.LogInfof("Setting machine property %s to %s", op.Operation, op.Value)
			if err := s.SetVmMachineOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "cpu":
			ctx.LogInfof("Setting cpu property %s to %s", op.Operation, op.Value)
			if err := s.SetVmCpu(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "memory":
			ctx.LogInfof("Setting memory property %s to %s", op.Operation, op.Value)
			if err := s.SetVmMemory(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "boot-order":
			ctx.LogInfof("Setting boot order property %s to %s", op.Operation, op.Value)
			if err := s.SetVmBootOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "efi-secure-boot":
			ctx.LogInfof("Setting boot efi secure boot property %s to %s", op.Operation, op.Value)
			if err := s.SetVmBootOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "select-boot-device":
			ctx.LogInfof("Setting select boot device property %s to %s", op.Operation, op.Value)
			if err := s.SetVmBootOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "external-boot-device":
			ctx.LogInfof("Setting external boot device property %s to %s", op.Operation, op.Value)
			if err := s.SetVmBootOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "time":
			ctx.LogInfof("Setting time sync property %s to %s", op.Operation, op.Value)
			if err := s.SetTimeSyncOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "network":
			ctx.LogInfof("Setting network property %s to %s", op.Operation, op.Value)
		case "device":
			ctx.LogInfof("Setting device property %s to %s", op.Operation, op.Value)
			if err := s.SetVmDeviceOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "shared-folder":
			ctx.LogInfof("Setting shared_folder property %s to %s", op.Operation, op.Value)
			if err := s.SetVmSharedFolderOperation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "rosetta":
			ctx.LogInfof("Setting rosetta property %s to %s", op.Operation, op.Value)
			if err := s.SetVmRosettaEmulation(ctx, vm, op); err != nil {
				op.Error = err
			}
		case "cmd":
			ctx.LogInfof("Setting custom property %s to %s", op.Operation, op.Value)
			if err := s.RunCustomCommand(ctx, vm, op); err != nil {
				op.Error = err
			}
		default:
			return errors.Newf("Invalid group %s", op.Group)
		}
	}

	return nil
}

func (s *ParallelsService) CreateVm(ctx basecontext.ApiContext, template data_models.PackerTemplate, desiredState string) (*models.ParallelsVM, error) {
	return s.CreatePackerTemplateVm(ctx, template, desiredState)
}

func (s *ParallelsService) CreatePackerTemplateVm(ctx basecontext.ApiContext, template data_models.PackerTemplate, desiredState string) (*models.ParallelsVM, error) {
	ctx.LogInfof("Creating Packer Virtual Machine %s", template.Name)
	existVm, err := s.findVm(ctx, template.Name)
	if existVm != nil || err != nil {
		if errors.GetSystemErrorCode(err) != 404 {
			return nil, errors.Newf("Machine %v with ID %v already exists and is %s", template.Name, existVm.ID, existVm.State)
		}
	}

	git := git.Get(ctx)
	repoPath, err := git.Clone(ctx, "https://github.com/Parallels/packer-examples", template.Owner, "packer-examples")
	if err != nil {
		ctx.LogErrorf("Error cloning packer-examples repository: %s", err.Error())
		return nil, err
	}

	ctx.LogInfof("Cloned packer-examples repository to %s", repoPath)

	packer := packer.Get(ctx)
	scriptPath := fmt.Sprintf("%s/%s", repoPath, template.PackerFolder)
	overrideFilePath := fmt.Sprintf("%s/%s/override.pkrvars.hcl", repoPath, template.PackerFolder)
	overrideFile := make(map[string]interface{})
	if template.Name != "" {
		overrideFile["machine_name"] = template.Name
	}
	if template.Hostname != "" {
		overrideFile["hostname"] = template.Hostname
	}
	overrideFile["create_vagrant_box"] = false
	overrideFile["machine_specs"] = map[string]interface{}{}
	if template.Specs["memory"] != "" {
		memory, err := strconv.Atoi(template.Specs["memory"])
		if err != nil {
			memory = 2048
		}
		overrideFile["machine_specs"].(map[string]interface{})["memory"] = memory
	}
	if template.Specs["cpu"] != "" {
		cpu, err := strconv.Atoi(template.Specs["cpu"])
		if err != nil {
			cpu = 2
		}
		overrideFile["machine_specs"].(map[string]interface{})["cpus"] = cpu
	}
	if template.Specs["disk"] != "" {
		disk, err := strconv.Atoi(template.Specs["disk"])
		if err != nil {
			disk = 40960
		}
		overrideFile["machine_specs"].(map[string]interface{})["disk_size"] = disk
	}

	template.Addons = append(template.Addons, "parallels-tools")
	if len(template.Addons) > 0 {
		overrideFile["addons"] = template.Addons
	}

	for key, value := range template.Variables {
		overrideFile[key] = value
	}

	overrideFileContent := helpers.ToHCL(overrideFile, 0)
	if err := helper.WriteToFile(overrideFileContent, overrideFilePath); err != nil {
		ctx.LogErrorf("Error writing override file %s: %s", overrideFilePath, err.Error())
		return nil, err
	}

	ctx.LogInfof("Created override file")

	ctx.LogInfof("Initializing packer repository")
	if err = packer.Init(ctx, template.Owner, scriptPath); err != nil {
		cleanError := helpers.RemoveFolder(repoPath)
		if cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, err
	}
	ctx.LogInfof("Initialized packer repository")

	ctx.LogInfof("Building packer machine")
	if err = packer.Build(ctx, template.Owner, scriptPath, overrideFilePath); err != nil {
		cleanError := helpers.RemoveFolder(repoPath)
		if cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, err
	}

	ctx.LogInfof("Built packer machine")

	users, err := system.Get().GetSystemUsers(ctx)
	if err != nil {
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, err
	}

	userExists := false
	if template.Owner == "root" {
		userExists = true
	} else {
		for _, user := range users {
			if user.Username == template.Owner {
				userExists = true
				break
			}
		}
	}

	if !userExists {
		ctx.LogErrorf("User %s does not exist", template.Owner)
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, errors.New("User does not exist")
	}

	userFolder := fmt.Sprintf("/Users/%s/Parallels", template.Owner)
	if template.Owner == "root" {
		userFolder = "/var/root"
	}

	err = helpers.CreateDirIfNotExist(userFolder)
	if err != nil {
		ctx.LogErrorf("Error creating user folder %s: %s", userFolder, err.Error())
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, err
	}

	ctx.LogInfof("Created user folder %s", userFolder)

	destinationFolder := fmt.Sprintf("%s/%s.pvm", userFolder, template.Name)
	sourceFolder := fmt.Sprintf("%s/out/%s.pvm", scriptPath, template.Name)
	err = helpers.MoveFolder(sourceFolder, destinationFolder)
	if err != nil {
		ctx.LogErrorf("Error moving folder %s to %s: %s", sourceFolder, destinationFolder, err.Error())
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		if helper.DirectoryExists(sourceFolder) {
			if cleanError := helpers.RemoveFolder(sourceFolder); cleanError != nil {
				ctx.LogErrorf("Error removing destination folder %s: %s", repoPath, cleanError.Error())
				return nil, cleanError
			}
		}
		return nil, err
	}

	if template.Owner != "root" {
		_, err = commands.ExecuteWithNoOutput("sudo", "chown", "-R", template.Owner, destinationFolder)
		if err != nil {
			ctx.LogErrorf("Error changing owner of folder %s to %s: %s", destinationFolder, template.Owner, err.Error())
			if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
				ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
				return nil, cleanError
			}
			return nil, err
		}
	}

	ctx.LogInfof("Moved folder %s to %s", sourceFolder, destinationFolder)
	_, err = commands.ExecuteWithNoOutput("sudo", "-u", template.Owner, s.executable, "register", destinationFolder)
	if err != nil {
		ctx.LogErrorf("Error registering VM %s: %s", destinationFolder, err.Error())
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, err
	}

	ctx.LogInfof("Registered VM %s", destinationFolder)

	existVm, err = s.findVm(ctx, template.Name)
	if existVm == nil || err != nil {
		ctx.LogErrorf("Error finding VM %s: %s", template.Name, err.Error())
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, errors.Newf("Something went wrong with creating machine %v, it does not exist, err: %v", template.Name, err.Error())
	}

	if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
		ctx.LogErrorf("Error removing folder %s: %s", repoPath, cleanError.Error())
		return nil, cleanError
	}

	switch desiredState {
	case "running":
		if err := s.StartVm(ctx, existVm.ID); err != nil {
			ctx.LogErrorf("Error starting VM %s: %s", existVm.ID, err.Error())
			return nil, err
		}
	default:
		ctx.LogInfof("Desired state is %s, not starting VM %s", desiredState, existVm.ID)
	}

	ctx.LogInfof("Created VM %s", existVm.ID)
	return existVm, nil
}

// Config Region
func (s *ParallelsService) SetVmMachineOperation(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	cmd := helpers.Command{
		Command: "sudo",
		Args:    make([]string, 0),
	}
	cmd.Args = append(cmd.Args, "-u", vm.User)

	switch op.Operation {
	case "clone":
		cmd.Args = append(cmd.Args, s.executable, "clone", vm.ID, "--name", op.Value)
		cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
	case "archive":
		cmd.Args = append(cmd.Args, s.executable, "archive", vm.ID)
	case "unarchive":
		cmd.Args = append(cmd.Args, s.executable, "unarchive", vm.ID)
	case "pack":
		cmd.Args = append(cmd.Args, s.executable, "pack", vm.ID)
	case "unpack":
		cmd.Args = append(cmd.Args, s.executable, "unpack", vm.ID)
	case "encrypt":
		cmd.Args = append(cmd.Args, s.executable, "encrypt", vm.ID)
		cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
	case "decrypt":
		cmd.Args = append(cmd.Args, s.executable, "decrypt", vm.ID)
		cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
	case "reset-uptime":
		cmd.Args = append(cmd.Args, s.executable, "reset-uptime", vm.ID)
	case "install-tools":
		cmd.Args = append(cmd.Args, s.executable, "install-tools", vm.ID)
	case "rename":
		cmd.Args = append(cmd.Args, s.executable, "set", vm.ID, "--name", op.Value)
	default:
		return errors.ErrConfigInvalidOperation(op.Operation)
	}

	ctx.LogInfof(cmd.String())
	_, err := helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetVmBootOperation(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	cmd := helpers.Command{
		Command: "sudo",
		Args:    make([]string, 0),
	}
	cmd.Args = append(cmd.Args, "-u", vm.User, s.executable, "set", vm.ID)

	switch op.Operation {
	case "boot-order":
		cmd.Args = append(cmd.Args, "--device-bootorder", op.Value)
	case "bios-type":
		if op.Value != "legacy" && op.Value != "efi32" && op.Value != "efi64" && op.Value != "efi-arm64" {
			return errors.ErrConfigInvalidBiosType(op.Value)
		}
		cmd.Args = append(cmd.Args, "--device-bootorder", op.Value)
	case "efi-secure-boot":
		if op.Value == "on" || op.Value == "true" {
			cmd.Args = append(cmd.Args, "--efi-secure-boot", "on")
		} else {
			cmd.Args = append(cmd.Args, "--efi-secure-boot", "off")
		}
	case "select-boot-device":
		if op.Value == "on" || op.Value == "true" {
			cmd.Args = append(cmd.Args, "--select-boot-device", "on")
		} else {
			cmd.Args = append(cmd.Args, "--select-boot-device", "off")
		}
	case "external-boot-device":
		cmd.Args = append(cmd.Args, "--external-boot-device", op.Value)
	default:
		return errors.ErrConfigInvalidOperation(op.Operation)
	}

	ctx.LogInfof(cmd.String())
	_, err := helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetVmSharedFolderOperation(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	cmd := helpers.Command{
		Command: "sudo",
		Args:    make([]string, 0),
	}
	cmd.Args = append(cmd.Args, "-u", vm.User, s.executable, "set", vm.ID)

	switch op.Operation {
	case "add":
		if op.GetOption("path").Value == "" {
			return errors.ErrConfigMissingSharedFolderPath()
		}
		cmd.Args = append(cmd.Args, "--shf-host-add", op.Value)
		cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
	case "set":
		cmd.Args = append(cmd.Args, "--shf-host-set", op.Value)
		cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
	case "delete":
		cmd.Args = append(cmd.Args, "--shf-host-del", op.Value)
	default:
		return errors.ErrConfigInvalidOperation(op.Operation)
	}

	ctx.LogInfof(cmd.String())
	_, err := helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetVmDeviceOperation(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	cmd := helpers.Command{
		Command: "sudo",
		Args:    make([]string, 0),
	}
	cmd.Args = append(cmd.Args, "-u", vm.User, s.executable, "set", vm.ID)

	switch op.Operation {
	case "add":
		switch op.Value {
		case "cdrom":
			cmd.Args = append(cmd.Args, "--device-add", "cdrom")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		case "fdd":
			cmd.Args = append(cmd.Args, "--device-add", "fdd")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		case "hdd":
			cmd.Args = append(cmd.Args, "--device-add", "hdd")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		case "net":
			cmd.Args = append(cmd.Args, "--device-add", "net")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		case "serial":
			cmd.Args = append(cmd.Args, "--device-add", "serial")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		case "parallel":
			cmd.Args = append(cmd.Args, "--device-add", "parallel")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		case "usb":
			cmd.Args = append(cmd.Args, "--device-add", "usb")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		case "sound":
			cmd.Args = append(cmd.Args, "--device-add", "sound")
			cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
		default:
			return errors.ErrConfigInvalidOperation(op.Value)
		}
	case "set":
		cmd.Args = append(cmd.Args, "--device-set", op.Value)
		cmd.Args = append(cmd.Args, op.GetCmdArgs()...)
	case "connect":
		cmd.Args = append(cmd.Args, "--device-connect", op.Value)
	case "disconnect":
		cmd.Args = append(cmd.Args, "--device-disconnect", op.Value)
	default:
		return errors.ErrConfigInvalidOperation(op.Operation)
	}

	ctx.LogInfof(cmd.String())
	_, err := helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetVmCpu(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	if vm.State != "stopped" {
		return errors.New("VM is not stopped")
	}
	cmd := "sudo"
	args := make([]string, 0)
	// Setting the owner in the command
	if op.Owner != "root" {
		args = append(args, "-u", op.Owner)
	}
	switch op.Operation {
	case "set":
		if op.Value != "auto" {
			_, err := strconv.Atoi(op.Value)
			if err != nil {
				return err
			}
		}
		args = append(args, s.executable, "set", vm.ID, "--cpus", op.Value)
	case "set_type":
		if op.Value != "x86" && op.Value != "arm" {
			return errors.Newf("Invalid CPU type %s", op.Value)
		}
		args = append(args, s.executable, "set", vm.ID, "--cpu-type", op.Value)
	default:
		return errors.Newf("Invalid operation %s", op.Operation)
	}

	_, err := commands.ExecuteWithNoOutput(cmd, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetVmMemory(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	if vm.State != "stopped" {
		return errors.New("VM is not stopped")
	}
	cmd := "sudo"
	args := make([]string, 0)
	// Setting the owner in the command
	if op.Owner != "root" {
		args = append(args, "-u", op.Owner)
	}

	switch op.Operation {
	case "set":
		if op.Value != "auto" {
			_, err := strconv.Atoi(op.Value)
			if err != nil {
				return err
			}
		}
		args = append(args, s.executable, "set", vm.ID, "--memsize", op.Value)
	default:
		return errors.Newf("Invalid operation %s", op.Operation)
	}

	_, err := commands.ExecuteWithNoOutput(cmd, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetVmRosettaEmulation(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	if vm.State != "stopped" {
		return errors.New("VM is not stopped")
	}
	cmd := "sudo"
	args := make([]string, 0)
	// Setting the owner in the command
	if op.Owner != "root" {
		args = append(args, "-u", op.Owner)
	}

	switch op.Operation {
	case "set":
		if op.Value != "on" && op.Value != "off" && op.Value != "true" && op.Value != "false" {
			return errors.Newf("Invalid value %s", op.Value)
		}

		if op.Value == "on" || op.Value == "true" {
			args = append(args, s.executable, "set", vm.ID, "--rosetta-linux", "on")
		}
		if op.Value == "off" || op.Value == "false" {
			args = append(args, s.executable, "set", vm.ID, "--rosetta-linux", "off")
		}
	default:
		return errors.Newf("Invalid operation %s", op.Operation)
	}

	_, err := commands.ExecuteWithNoOutput(cmd, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetTimeSyncOperation(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	cmd := helpers.Command{
		Command: "sudo",
		Args:    make([]string, 0),
	}
	cmd.Args = append(cmd.Args, "-u", vm.User, s.executable, "set", vm.ID)

	switch op.Operation {
	case "time-sync":
		if op.Value == "on" || op.Value == "true" {
			cmd.Args = append(cmd.Args, "--time-sync", "on")
		} else {
			cmd.Args = append(cmd.Args, "--time-sync", "off")
		}
	case "time-sync-smart-mode":
		if op.Value == "on" || op.Value == "true" {
			cmd.Args = append(cmd.Args, "--time-sync-smart-mode", "on")
		} else {
			cmd.Args = append(cmd.Args, "--time-sync-smart-mode", "off")
		}
	case "disable-timezone-synct":
		if op.Value == "on" || op.Value == "true" {
			cmd.Args = append(cmd.Args, "--disable-timezone-sync", "on")
		} else {
			cmd.Args = append(cmd.Args, "--disable-timezone-sync", "off")
		}
	case "time-sync-interval":
		cmd.Args = append(cmd.Args, "--time-sync-interval", op.Value)
	default:
		return errors.ErrConfigInvalidOperation(op.Operation)
	}

	ctx.LogInfof(cmd.String())
	_, err := helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) ExecuteCommandOnVm(ctx basecontext.ApiContext, id string, r *models.VirtualMachineExecuteCommandRequest) (*models.VirtualMachineExecuteCommandResponse, error) {
	response := &models.VirtualMachineExecuteCommandResponse{}
	vm, err := s.findVm(ctx, id)
	if err != nil {
		return nil, err
	}
	if vm == nil {
		return nil, errors.New("VM not found")
	}

	if vm.State != "running" {
		return nil, errors.New("VM is not running")
	}

	cmd := helpers.Command{
		Command: "sudo",
	}
	args := make([]string, 0)
	// Setting the owner in the command
	if vm.User != "root" {
		args = append(args, "-u", vm.User)
	}
	args = append(args, s.executable, "exec", vm.ID, r.Command)
	cmd.Args = args

	ctx.LogInfof("Executing command %s %s", cmd.Command, strings.Join(cmd.Args, " "))
	stdout, stderr, exitCode, cmdError := helpers.ExecuteWithOutput(cmd)
	response.Stdout = stdout
	response.Stderr = stderr
	response.ExitCode = exitCode
	if cmdError != nil {
		response.Error = cmdError.Error()
	}

	ctx.LogInfof("Command %s %s executed with exit code %v", cmd.Command, strings.Join(cmd.Args, " "), exitCode)
	return response, nil
}

func (s *ParallelsService) ReplaceMachineNameInConfigPvs(path string, newName string) error {
	configPath := filepath.Join(path, "config.pvs")
	if !helper.FileExists(configPath) {
		return errors.Newf("Config file %s not found", configPath)
	}

	// fileInfo, err := os.Stat("filename")
	// if err != nil {
	// 	return err
	// }

	// fileMode := fileInfo.Mode()
	// // Get the file owner
	// sys := fileInfo.Sys().(*syscall.Stat_t)
	// uid := sys.Uid
	// gid := int(sys.Gid)

	file, err := os.Open(filepath.Clean(configPath))
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := helper.ReadFromFile(configPath)
	if err != nil {
		return err
	}

	pattern := regexp.MustCompile(`<[Vv]m[Nn]ame>[^<]*</[Vv]m[Nn]ame>`)

	newContent := pattern.ReplaceAllString(string(content), fmt.Sprintf("<VmName>%s</VmName>", newName))

	err = helper.WriteToFile(newContent, configPath)
	if err != nil {
		return err
	}

	// // Set the mode and owner of another file to be the same as configPath
	// err = os.Chmod(configPath, fileMode)
	// if err != nil {
	// 	return err
	// }
	// err = os.Chown(configPath, uid, gid)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (s *ParallelsService) RunCustomCommand(ctx basecontext.ApiContext, vm *models.ParallelsVM, op *models.VirtualMachineConfigRequestOperation) error {
	if vm.State != "stopped" {
		return errors.New("VM is not stopped")
	}
	cmd := "sudo"
	args := make([]string, 0)
	// Setting the owner in the command
	if op.Owner != "root" {
		args = append(args, "-u", op.Owner)
	}

	args = append(args, s.executable, op.Operation, vm.ID)
	args = append(args, op.GetCmdArgs()...)

	_, err := commands.ExecuteWithNoOutput(cmd, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) GetHardwareUsage(ctx basecontext.ApiContext) (*models.SystemUsageResponse, error) {
	result := &models.SystemUsageResponse{}

	vms, err := s.GetVms(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if vm.State == "running" {
			result.TotalInUse.LogicalCpuCount += vm.Hardware.CPU.Cpus
			memorySizeInt, err := helpers.GetSizeByteFromString(vm.Hardware.Memory.Size)
			if err != nil {
				return nil, err
			}
			result.TotalInUse.MemorySize += helpers.ConvertByteToMegabyte(memorySizeInt)
			if vm.Hardware.Hdd0.Size != "" {
				hddSizeInt, err := helpers.GetSizeByteFromString(vm.Hardware.Hdd0.Size)
				if err != nil {
					return nil, err
				}
				result.TotalInUse.DiskSize += helpers.ConvertByteToMegabyte(hddSizeInt)
			}
		} else {
			result.TotalReserved.LogicalCpuCount += vm.Hardware.CPU.Cpus
			memorySizeInt, err := helpers.GetSizeByteFromString(vm.Hardware.Memory.Size)
			if err != nil {
				return nil, err
			}
			result.TotalReserved.MemorySize += helpers.ConvertByteToMegabyte(memorySizeInt)
			if vm.Hardware.Hdd0.Size != "" {
				hddSizeInt, err := helpers.GetSizeByteFromString(vm.Hardware.Hdd0.Size)
				if err != nil {
					return nil, err
				}
				result.TotalReserved.DiskSize += helpers.ConvertByteToMegabyte(hddSizeInt)
			}
		}
	}

	systemSrv := system.Get()
	systemInfo, err := systemSrv.GetHardwareInfo(ctx)
	if err != nil {
		return nil, err
	}
	result.CpuType = systemInfo.CpuType
	result.CpuBrand = systemInfo.CpuBrand
	result.Total.LogicalCpuCount = int64(systemInfo.LogicalCpuCount)
	result.Total.MemorySize = systemInfo.MemorySize
	result.Total.DiskSize = systemInfo.DiskSize - systemInfo.FreeDiskSize

	result.TotalAvailable.DiskSize = result.Total.DiskSize - result.TotalReserved.DiskSize - result.TotalInUse.DiskSize
	result.TotalAvailable.MemorySize = result.Total.MemorySize - result.TotalInUse.MemorySize
	result.TotalAvailable.LogicalCpuCount = result.Total.LogicalCpuCount - result.TotalInUse.LogicalCpuCount

	return result, nil
}
