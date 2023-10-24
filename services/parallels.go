package services

import (
	"Parallels/pd-api-service/common"
	data_models "Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/models"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/cjlapao/common-go/commands"
	"github.com/cjlapao/common-go/helper"
)

var globalParallelsService *ParallelsService

type ParallelsVirtualMachineState string

const (
	ParallelsVirtualMachineStateStopped   ParallelsVirtualMachineState = "stopped"
	ParallelsVirtualMachineStateRunning   ParallelsVirtualMachineState = "running"
	ParallelsVirtualMachineStateSuspended ParallelsVirtualMachineState = "suspended"
	ParallelsVirtualMachineStatePaused    ParallelsVirtualMachineState = "paused"
	ParallelsVirtualMachineStateUnknown   ParallelsVirtualMachineState = "unknown"
)

type ParallelsVirtualMachineDesiredState string

const (
	ParallelsVirtualMachineDesiredStateStop    ParallelsVirtualMachineDesiredState = "stop"
	ParallelsVirtualMachineDesiredStateStart   ParallelsVirtualMachineDesiredState = "start"
	ParallelsVirtualMachineDesiredStatePause   ParallelsVirtualMachineDesiredState = "pause"
	ParallelsVirtualMachineDesiredStateSuspend ParallelsVirtualMachineDesiredState = "suspend"
	ParallelsVirtualMachineDesiredStateResume  ParallelsVirtualMachineDesiredState = "resume"
	ParallelsVirtualMachineDesiredStateReset   ParallelsVirtualMachineDesiredState = "reset"
	ParallelsVirtualMachineDesiredStateRestart ParallelsVirtualMachineDesiredState = "restart"
	ParallelsVirtualMachineDesiredStateUnknown ParallelsVirtualMachineDesiredState = "unknown"
)

func (s ParallelsVirtualMachineState) String() string {
	return string(s)
}

func (s ParallelsVirtualMachineDesiredState) String() string {
	return string(s)
}

func ParallelsVirtualMachineDesiredStateFromString(s string) ParallelsVirtualMachineDesiredState {
	switch s {
	case "stop":
		return ParallelsVirtualMachineDesiredStateStop
	case "start":
		return ParallelsVirtualMachineDesiredStateStart
	case "pause":
		return ParallelsVirtualMachineDesiredStatePause
	case "suspend":
		return ParallelsVirtualMachineDesiredStateSuspend
	case "resume":
		return ParallelsVirtualMachineDesiredStateResume
	case "reset":
		return ParallelsVirtualMachineDesiredStateReset
	case "restart":
		return ParallelsVirtualMachineDesiredStateRestart
	default:
		return ParallelsVirtualMachineDesiredStateUnknown
	}
}

type ParallelsService struct {
	executable       string
	serverExecutable string
	Info             *models.ParallelsDesktopInfo
}

func NewParallelsService() *ParallelsService {
	if globalParallelsService != nil {
		return globalParallelsService
	}

	globalParallelsService = &ParallelsService{}

	if globalParallelsService.executable == "" {
		globalServices.Logger.Info("Getting parallels executable")
		out, err := commands.ExecuteWithNoOutput("which", "prlctl")
		path := strings.ReplaceAll(strings.TrimSpace(out), "\n", "")
		if err != nil || path == "" {
			globalServices.Logger.Warn("Parallels executable not found, trying to find it in the default locations")
		}

		if path != "" {
			globalParallelsService.executable = path
			globalParallelsService.serverExecutable = strings.ReplaceAll(path, "prlctl", "prlsrvctl")
		} else {
			if _, err := os.Stat("/usr/bin/prlctl"); err == nil {
				globalParallelsService.executable = "/usr/bin/prlctl"
				globalParallelsService.serverExecutable = "/usr/bin/prlsrvctl"
				os.Setenv("PATH", os.Getenv("PATH")+":/usr/bin")
				// if globalServices.RunningUser == "root" {
				// 	_, err := helpers.ExecuteWithNoOutput(helpers.Command{
				// 		Command: "export",
				// 		Args:    []string{"PATH=$PATH:/usr/bin"},
				// 	})
				// 	if err != nil {
				// 		panic(err)
				// 	}
				// }
			} else if _, err := os.Stat("/usr/local/bin/prlctl"); err == nil {
				globalParallelsService.executable = "/usr/local/bin/prlctl"
				globalParallelsService.serverExecutable = "/usr/local/bin/prlsrvctl"
				os.Setenv("PATH", os.Getenv("PATH")+":/usr/local/bin")
				// if globalServices.RunningUser == "root" {
				// 	_, err := helpers.ExecuteWithNoOutput(helpers.Command{
				// 		Command: "export",
				// 		Args:    []string{"PATH=$PATH:/usr/local/bin"},
				// 	})
				// 	if err != nil {
				// 		panic(err)
				// 	}
				// }
			} else {
				panic(errors.New("Parallels executable not found"))
			}
		}
		globalServices.Logger.Info("Parallels executable: " + globalParallelsService.executable)
	}

	return globalParallelsService
}

func (s *ParallelsService) findVm(idOrName string) (*models.ParallelsVM, error) {
	vms, err := s.GetVms()
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if strings.EqualFold(vm.Name, idOrName) || strings.EqualFold(vm.ID, idOrName) {
			return &vm, nil
		}
	}

	return nil, nil
}

func (s *ParallelsService) GetVms() ([]models.ParallelsVM, error) {
	var result []models.ParallelsVM
	users, err := GetSystemUsers()
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.New("No users found")
	}

	for _, user := range users {
		globalServices.Logger.Info("Getting VMs for user: " + user.Username)
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
			for _, globalMachine := range result {
				if strings.EqualFold(machine.ID, globalMachine.ID) {
					found = true
					break
				}
			}
			if !found {
				machine.User = user.Username
				result = append(result, machine)
			}
		}
	}

	return result, nil
}

func (s *ParallelsService) GetVm(id string) (*models.ParallelsVM, error) {
	vm, err := s.findVm(id)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func (s *ParallelsService) GetFilteredVm(filter string) ([]models.ParallelsVM, error) {
	filterParts := strings.Split(filter, "=")
	if len(filterParts) != 2 {
		return nil, errors.New("Invalid filter")
	}
	vms, err := s.GetVms()
	if err != nil {
		return nil, err
	}

	if len(vms) == 0 {
		return nil, errors.New("No VMs found")
	}

	if !hasProperty(vms[0], filterParts[0]) {
		return nil, errors.New("Invalid filter property")
	}

	common.Logger.Info("Getting filtered VMs for property %s with value %s", filterParts[0], filterParts[1])
	filteredMachines := make([]models.ParallelsVM, 0)

	for _, vm := range vms {
		value, err := getPropertyAsString(vm, filterParts[0])
		if err != nil {
			continue
		}

		// Match filterParts[1] using a regex expression
		exp, err := regexp.Compile(filterParts[1])
		if err != nil {
			return nil, err
		}

		matched := exp.MatchString(value)
		if matched {
			filteredMachines = append(filteredMachines, vm)
		}
	}

	return filteredMachines, nil
}

func (s *ParallelsService) SetVmState(id string, desiredState ParallelsVirtualMachineDesiredState) error {
	vm, err := s.findVm(id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.New("VM not found")
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

func (s *ParallelsService) StartVm(id string) error {
	return s.SetVmState(id, ParallelsVirtualMachineDesiredStateStart)
}

func (s *ParallelsService) StopVm(id string) error {
	return s.SetVmState(id, ParallelsVirtualMachineDesiredStateStop)
}

func (s *ParallelsService) RestartVm(id string) error {
	return s.SetVmState(id, ParallelsVirtualMachineDesiredStateRestart)
}

func (s *ParallelsService) SuspendVm(id string) error {
	return s.SetVmState(id, ParallelsVirtualMachineDesiredStateSuspend)
}

func (s *ParallelsService) ResumeVm(id string) error {
	return s.SetVmState(id, ParallelsVirtualMachineDesiredStateResume)
}

func (s *ParallelsService) ResetVm(id string) error {
	return s.SetVmState(id, ParallelsVirtualMachineDesiredStateReset)
}

func (s *ParallelsService) PauseVm(id string) error {
	return s.SetVmState(id, ParallelsVirtualMachineDesiredStatePause)
}

func (s *ParallelsService) DeleteVm(id string) error {
	vm, err := s.findVm(id)
	if err != nil {
		return err
	}

	if vm == nil {
		return fmt.Errorf("VM with id %s was not found", id)
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

func (s *ParallelsService) VmStatus(id string) (string, error) {
	vm, err := s.findVm(id)
	if err != nil {
		return "", err
	}
	if vm == nil {
		return "", errors.New("VM not found")
	}

	output, err := commands.ExecuteWithNoOutput("sudo", "-u", vm.User, s.executable, "status", id)

	statusParts := strings.Split(output, " ")
	if len(statusParts) != 4 {
		return "", errors.New("Invalid status output")
	}

	return strings.ReplaceAll(statusParts[3], "\n", ""), nil
}

func (s *ParallelsService) GetInfo() *models.ParallelsDesktopInfo {
	if s.Info != nil {
		return s.Info
	}

	stdout, err := helpers.ExecuteWithNoOutput(helpers.Command{
		Command: s.serverExecutable,
		Args:    []string{"info", "--json"},
	})
	if err != nil {
		return nil
	}

	var info models.ParallelsDesktopInfo
	err = json.Unmarshal([]byte(stdout), &info)
	if err != nil {
		return nil
	}

	s.Info = &info

	return s.Info
}

func (s *ParallelsService) ConfigVmSetRequest(id string, setOperations *models.VirtualMachineSetRequest) error {
	vm, err := s.findVm(id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.New("VM not found")
	}

	for _, op := range setOperations.Operations {
		op.Owner = vm.User
		switch op.Group {
		case "state":
			common.Logger.Info("Setting machine state to %s", op.Operation)
			if err := s.SetVmState(vm.ID, ParallelsVirtualMachineDesiredStateFromString(op.Operation)); err != nil {
				op.Error = err
			}
		case "machine":
			common.Logger.Info("Setting machine property %s to %s", op.Operation, op.Value)
			if err := s.SetVmMachineOperation(vm, op); err != nil {
				op.Error = err
			}
		case "cpu":
			common.Logger.Info("Setting cpu property %s to %s", op.Operation, op.Value)
			if err := s.SetVmCpu(vm, op); err != nil {
				op.Error = err
			}
		case "memory":
			common.Logger.Info("Setting memory property %s to %s", op.Operation, op.Value)
			if err := s.SetVmMemory(vm, op); err != nil {
				op.Error = err
			}
		case "network":
			common.Logger.Info("Setting network property %s to %s", op.Operation, op.Value)
		case "device":
			common.Logger.Info("Setting device property %s to %s", op.Operation, op.Value)
		case "shared_folder":
			common.Logger.Info("Setting shared_folder property %s to %s", op.Operation, op.Value)
		case "rosetta":
			common.Logger.Info("Setting rosetta property %s to %s", op.Operation, op.Value)
			if err := s.SetVmRosettaEmulation(vm, op); err != nil {
				op.Error = err
			}

		default:
			return fmt.Errorf("Invalid group %s", op.Group)
		}
	}

	return nil
}

func (s *ParallelsService) CreateVm(template data_models.VirtualMachineTemplate, desiredState string) (*models.ParallelsVM, error) {
	switch template.Type {
	case data_models.VirtualMachineTemplateTypePacker:
		return s.CreatePackerTemplateVm(template, desiredState)
	}

	// _, err := commands.Execute("prlctl", "set", id, "--bootorder", bootOrder)
	// if err != nil {
	// 	return err
	// }

	return nil, nil
}

func (s *ParallelsService) CreatePackerTemplateVm(template data_models.VirtualMachineTemplate, desiredState string) (*models.ParallelsVM, error) {
	common.Logger.Info("Creating Packer Virtual Machine %s", template.Name)
	existVm, err := s.findVm(template.Name)
	if existVm != nil || err != nil {
		return nil, fmt.Errorf("Machine %v with ID %v already exists and is %s", template.Name, existVm.ID, existVm.State)
	}

	git := globalServices.GitService
	repoPath, err := git.Clone("https://github.com/Parallels/packer-examples", "packer-examples")
	if err != nil {
		common.Logger.Error("Error cloning packer-examples repository: %s", err.Error())
		return nil, err
	}

	common.Logger.Info("Cloned packer-examples repository to %s", repoPath)

	packer := globalServices.PackerService
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
	helper.WriteToFile(overrideFileContent, overrideFilePath)
	common.Logger.Info("Created override file")

	common.Logger.Info("Initializing packer repository")
	if err = packer.Init(scriptPath); err != nil {
		cleanError := helpers.RemoveFolder(repoPath)
		if cleanError != nil {
			common.Logger.Error("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, err
	}
	common.Logger.Info("Initialized packer repository")

	common.Logger.Info("Building packer machine")
	if err = packer.Build(scriptPath, overrideFilePath); err != nil {
		cleanError := helpers.RemoveFolder(repoPath)
		if cleanError != nil {
			common.Logger.Error("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, err
	}

	common.Logger.Info("Built packer machine")

	users, err := GetSystemUsers()
	if err != nil {
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			common.Logger.Error("Error removing folder %s: %s", repoPath, cleanError.Error())
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
		common.Logger.Error("User %s does not exist", template.Owner)
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			common.Logger.Error("Error removing folder %s: %s", repoPath, cleanError.Error())
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
		common.Logger.Error("Error creating user folder %s: %s", userFolder, err.Error())
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			common.Logger.Error("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, err
	}

	common.Logger.Info("Created user folder %s", userFolder)

	destinationFolder := fmt.Sprintf("%s/%s.pvm", userFolder, template.Name)
	sourceFolder := fmt.Sprintf("%s/out/%s.pvm", scriptPath, template.Name)
	err = helpers.MoveFolder(sourceFolder, destinationFolder)
	if err != nil {
		common.Logger.Error("Error moving folder %s to %s: %s", sourceFolder, destinationFolder, err.Error())
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			common.Logger.Error("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		if helper.DirectoryExists(sourceFolder) {
			if cleanError := helpers.RemoveFolder(sourceFolder); cleanError != nil {
				common.Logger.Error("Error removing destination folder %s: %s", repoPath, cleanError.Error())
				return nil, cleanError
			}
		}
		return nil, err
	}

	if template.Owner != "root" {
		_, err = commands.ExecuteWithNoOutput("sudo", "chown", "-R", template.Owner, destinationFolder)
		if err != nil {
			common.Logger.Error("Error changing owner of folder %s to %s: %s", destinationFolder, template.Owner, err.Error())
			if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
				common.Logger.Error("Error removing folder %s: %s", repoPath, cleanError.Error())
				return nil, cleanError
			}
			return nil, err
		}
	}

	common.Logger.Info("Moved folder %s to %s", sourceFolder, destinationFolder)
	_, err = commands.ExecuteWithNoOutput("sudo", "-u", template.Owner, s.executable, "register", destinationFolder)
	if err != nil {
		common.Logger.Error("Error registering VM %s: %s", destinationFolder, err.Error())
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			common.Logger.Error("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, err
	}

	common.Logger.Info("Registered VM %s", destinationFolder)

	existVm, err = s.findVm(template.Name)
	if existVm == nil || err != nil {
		common.Logger.Error("Error finding VM %s: %s", template.Name, err.Error())
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			common.Logger.Error("Error removing folder %s: %s", repoPath, cleanError.Error())
			return nil, cleanError
		}
		return nil, fmt.Errorf("Something went wrong with creating machine %v, it does not exist, err: %v", template.Name, err.Error())
	}

	if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
		common.Logger.Error("Error removing folder %s: %s", repoPath, cleanError.Error())
		return nil, cleanError
	}

	switch desiredState {
	case "running":
		if err := s.StartVm(existVm.ID); err != nil {
			common.Logger.Error("Error starting VM %s: %s", existVm.ID, err.Error())
			return nil, err
		}
	default:
		common.Logger.Info("Desired state is %s, not starting VM %s", desiredState, existVm.ID)
	}

	common.Logger.Info("Created VM %s", existVm.ID)
	return existVm, nil
}

func hasProperty(obj interface{}, propertyName string) bool {
	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return false
	}
	_, ok := value.Type().FieldByName(propertyName)
	return ok
}

func getPropertyAsString(obj interface{}, propertyName string) (string, error) {
	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return "", errors.New("obj is not a struct")
	}
	field := value.FieldByName(propertyName)
	if !field.IsValid() {
		return "", fmt.Errorf("property %s not found", propertyName)
	}
	return fmt.Sprintf("%v", field.Interface()), nil
}

// Config Region

func (s *ParallelsService) SetVmMachineOperation(vm *models.ParallelsVM, op *models.VirtualMachineSetOperation) error {
	args := make([]string, 0)
	switch op.Operation {
	case "clone":
		args = append(args, []string{
			"sudo",
			"-u",
			vm.User,
		}...)
		args = append(args, s.executable, "clone", op.Value, "--name", vm.Name)
	}
	// _, err = commands.ExecuteWithNoOutput("sudo", "-u", vm.User, s.executable, "set", id, "--memSize", size)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (s *ParallelsService) SetVmCpu(vm *models.ParallelsVM, op *models.VirtualMachineSetOperation) error {
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
			return fmt.Errorf("Invalid CPU type %s", op.Value)
		}
		args = append(args, s.executable, "set", vm.ID, "--cpu-type", op.Value)
	default:
		return fmt.Errorf("Invalid operation %s", op.Operation)
	}

	_, err := commands.ExecuteWithNoOutput(cmd, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetVmMemory(vm *models.ParallelsVM, op *models.VirtualMachineSetOperation) error {
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
		args = append(args, s.executable, "set", vm.ID, "--memSize", op.Value)
	default:
		return fmt.Errorf("Invalid operation %s", op.Operation)
	}

	_, err := commands.ExecuteWithNoOutput(cmd, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetVmRosettaEmulation(vm *models.ParallelsVM, op *models.VirtualMachineSetOperation) error {
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
			return fmt.Errorf("Invalid value %s", op.Value)
		}

		if op.Value == "on" || op.Value == "true" {
			args = append(args, s.executable, "set", vm.ID, "--rosetta-linux", "on")
		}
		if op.Value == "off" || op.Value == "false" {
			args = append(args, s.executable, "set", vm.ID, "--rosetta-linux", "off")
		}
	default:
		return fmt.Errorf("Invalid operation %s", op.Operation)
	}

	_, err := commands.ExecuteWithNoOutput(cmd, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) ExecuteCommandOnVm(id string, r *models.VirtualMachineExecuteCommandRequest) (*models.VirtualMachineExecuteCommandResponse, error) {
	response := &models.VirtualMachineExecuteCommandResponse{}
	vm, err := s.findVm(id)
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

	common.Logger.Info("Executing command %s %s", cmd.Command, strings.Join(cmd.Args, " "))
	stdout, stderr, exitCode, cmdError := helpers.ExecuteWithOutput(cmd)
	response.Stdout = stdout
	response.Stderr = stderr
	response.ExitCode = exitCode
	if cmdError != nil {
		response.Error = cmdError.Error()
	}

	common.Logger.Info("Command %s %s executed with exit code %v", cmd.Command, strings.Join(cmd.Args, " "), exitCode)
	return response, nil
}
