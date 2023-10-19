package services

import (
	data_models "Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/models"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cjlapao/common-go/commands"
	"github.com/cjlapao/common-go/helper"
)

var globalParallelsService *ParallelsService

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

func (s *ParallelsService) StartVm(id string) error {
	vm, err := s.findVm(id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.New("VM not found")
	}

	if vm.State == "running" {
		return nil
	}

	if vm.State != "stopped" {
		return errors.New("VM is not stopped")
	}

	_, err = commands.ExecuteWithNoOutput("sudo", "-u", vm.User, s.executable, "start", id)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) StopVm(id string) error {
	vm, err := s.findVm(id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.New("VM not found")
	}

	if vm.State == "stopped" {
		return nil
	}

	if vm.State != "running" {
		return errors.New("VM is not running")
	}

	_, err = commands.ExecuteWithNoOutput("sudo", "-u", vm.User, s.executable, "stop", id)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) RestartVm(id string) error {

	vm, err := s.findVm(id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.New("VM not found")
	}

	if vm.State != "running" {
		return errors.New("VM is not running")
	}

	_, err = commands.ExecuteWithNoOutput("sudo", "-u", vm.User, s.executable, "restart", id)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SuspendVm(id string) error {
	vm, err := s.findVm(id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.New("VM not found")
	}

	if vm.State == "suspended" {
		return nil
	}

	if vm.State != "running" {
		return errors.New("VM is not running")
	}

	_, err = commands.ExecuteWithNoOutput("sudo", "-u", vm.User, s.executable, "suspend", id)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) ResumeVm(id string) error {
	vm, err := s.findVm(id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.New("VM not found")
	}

	if vm.State != "suspended" && vm.State != "paused" {
		return errors.New("VM is not running")
	}

	_, err = commands.ExecuteWithNoOutput("sudo", "-u", vm.User, s.executable, "resume", id)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) ResetVm(id string) error {
	vm, err := s.findVm(id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.New("VM not found")
	}

	if vm.State == "stopped" {
		return nil
	}

	_, err = commands.ExecuteWithNoOutput("sudo", "-u", vm.User, s.executable, "reset", id)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) PauseVm(id string) error {
	vm, err := s.findVm(id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.New("VM not found")
	}

	if vm.State == "paused" {
		return nil
	}

	if vm.State != "running" {
		return errors.New("VM is not running")
	}

	_, err = commands.ExecuteWithNoOutput("sudo", "-u", vm.User, s.executable, "pause", id)
	if err != nil {
		return err
	}

	return nil
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

func (s *ParallelsService) SetCpuSize(id string, size int) error {
	vm, err := s.findVm(id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.New("VM not found")
	}

	var cpuCount string
	if size == 0 {
		cpuCount = "auto"
	} else {
		cpuCount = fmt.Sprintf("%d", size)
	}

	_, err = commands.ExecuteWithNoOutput("sudo", "-u", vm.User, s.executable, "set", id, "--cpus", cpuCount)
	if err != nil {
		return err
	}

	return nil
}

func (s *ParallelsService) SetMemorySize(id string, size int) error {
	vm, err := s.findVm(id)
	if err != nil {
		return err
	}
	if vm == nil {
		return errors.New("VM not found")
	}

	var memSize string
	if size == 0 {
		memSize = "auto"
	} else {
		memSize = fmt.Sprintf("%d", size)
	}
	_, err = commands.ExecuteWithNoOutput("sudo", "-u", vm.User, s.executable, "set", id, "--memSize", memSize)
	if err != nil {
		return err
	}

	return nil
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

func (s *ParallelsService) CreateVirtualMachine(template data_models.VirtualMachineTemplate) (*models.ParallelsVM, error) {
	switch template.Type {
	case data_models.VirtualMachineTemplateTypePacker:
		return s.CreatePackerVirtualMachine(template)
	}

	// _, err := commands.Execute("prlctl", "set", id, "--bootorder", bootOrder)
	// if err != nil {
	// 	return err
	// }

	return nil, nil
}

func (s *ParallelsService) CreatePackerVirtualMachine(template data_models.VirtualMachineTemplate) (*models.ParallelsVM, error) {
	existVm, err := s.findVm(template.Name)
	if existVm != nil || err != nil {
		return nil, fmt.Errorf("Machine %v with ID %v already exists and is %s", template.Name, existVm.ID, existVm.State)
	}

	git := globalServices.GitService
	repoPath, err := git.Clone("https://github.com/Parallels/packer-examples", "packer-examples")
	if err != nil {
		return nil, err
	}

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
	overrideFile["machine_specs"] = map[string]interface{}{
		"memory":    template.Specs["memory"],
		"cpus":      template.Specs["cpu"],
		"disk_size": fmt.Sprintf("%d", template.Specs["disk"]),
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
	if err = packer.Init(scriptPath); err != nil {
		cleanError := helpers.RemoveFolder(repoPath)
		if cleanError != nil {
			return nil, cleanError
		}
		return nil, err
	}

	if err = packer.Build(scriptPath, overrideFilePath); err != nil {
		cleanError := helpers.RemoveFolder(repoPath)
		if cleanError != nil {
			return nil, cleanError
		}
		return nil, err
	}

	users, err := GetSystemUsers()
	if err != nil {
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
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
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			return nil, cleanError
		}
		return nil, errors.New("User does not exist")
	}

	err = helpers.CreateDirIfNotExist(fmt.Sprintf("/Users/%s/Parallels", template.Owner))
	if err != nil {
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			return nil, cleanError
		}
		return nil, err
	}

	destinationFolder := fmt.Sprintf("/Users/%s/Parallels/%s.pvm", template.Owner, template.Name)
	sourceFolder := fmt.Sprintf("%s/out/%s.pvm", scriptPath, template.Name)
	err = helpers.MoveFolder(sourceFolder, destinationFolder)
	if err != nil {
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			return nil, cleanError
		}
		return nil, err
	}

	_, err = commands.ExecuteWithNoOutput("sudo", "-u", template.Owner, s.executable, "register", destinationFolder)
	if err != nil {
		if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
			return nil, cleanError
		}
		return nil, err
	}

	existVm, err = s.findVm(template.Name)
	if existVm == nil || err != nil {
		return nil, fmt.Errorf("Something went wrong with creating machine %v, it does not exist, err: %v", template.Name, err.Error())
	}

	if cleanError := helpers.RemoveFolder(repoPath); cleanError != nil {
		return nil, cleanError
	}

	return existVm, nil
}
