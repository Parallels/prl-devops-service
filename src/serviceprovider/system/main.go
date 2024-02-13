package system

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/serviceprovider/interfaces"

	"github.com/cjlapao/common-go/commands"
)

var globalSystemService *SystemService

type SystemService struct {
	ctx            basecontext.ApiContext
	brewExecutable string
	installed      bool
	dependencies   []interfaces.Service
}

func Get() *SystemService {
	if globalSystemService != nil {
		return globalSystemService
	}

	ctx := basecontext.NewBaseContext()

	return New(ctx)
}

func New(ctx basecontext.ApiContext) *SystemService {
	globalSystemService = &SystemService{
		ctx: ctx,
	}

	if globalSystemService.GetOperatingSystem() == "macos" && globalSystemService.FindPath() == "" {
		ctx.LogWarnf("Running without support for brew")
		return globalSystemService
	} else {
		globalSystemService.installed = true
	}

	globalSystemService.SetDependencies([]interfaces.Service{})
	return globalSystemService
}

func (s *SystemService) Name() string {
	return "system"
}

func (s *SystemService) FindPath() string {
	s.ctx.LogInfof("Getting brew executable")
	out, err := commands.ExecuteWithNoOutput("which", "brew")
	path := strings.ReplaceAll(strings.TrimSpace(out), "\n", "")
	if err != nil || path == "" {
		s.ctx.LogWarnf("Brew executable not found, trying to find it in the default locations")
	}

	if path != "" {
		s.brewExecutable = path
		s.ctx.LogInfof("Brew found at: %s", s.brewExecutable)
	} else {
		if _, err := os.Stat("/opt/homebrew/bin/brew"); err == nil {
			s.brewExecutable = "/opt/homebrew/bin/brew"
		} else if _, err := os.Stat("/usr/local/bin/brew"); err == nil {
			s.brewExecutable = "/usr/local/bin/brew"
		} else {
			s.ctx.LogWarnf("Brew executable not found")
			return s.brewExecutable
		}

		s.ctx.LogInfof("Brew found at: %s", s.brewExecutable)
	}

	return s.brewExecutable
}

func (s *SystemService) Version() string {
	cmd := helpers.Command{
		Command: s.brewExecutable,
		Args:    []string{"--version"},
	}

	stdout, _, _, err := helpers.ExecuteWithOutput(cmd)
	if err != nil {
		return "unknown"
	}

	vParts := strings.Split(stdout, " ")
	if len(vParts) > 0 {
		return strings.TrimSpace(vParts[1])
	} else {
		return stdout
	}
}

func (s *SystemService) Install(asUser, version string, flags map[string]string) error {
	if s.installed {
		s.ctx.LogInfof("%s already installed", s.Name())
		return nil
	}

	// Installing service dependency
	if s.dependencies != nil {
		for _, dependency := range s.dependencies {
			if dependency == nil {
				return errors.New("Dependency is nil")
			}
			s.ctx.LogInfof("Installing dependency %s", dependency.Name())
			if err := dependency.Install(asUser, "latest", flags); err != nil {
				return err
			}
		}
	}

	cmd := helpers.Command{
		Command: "/bin/bash",
		Args:    []string{"-c", "\"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""},
	}

	s.ctx.LogInfof("Installing %s with command: %v", s.Name(), cmd.String())
	_, err := helpers.ExecuteWithNoOutput(cmd)
	if err != nil {
		return err
	}

	s.installed = true
	return nil
}

func (s *SystemService) Uninstall(asUser string, uninstallDependencies bool) error {
	if s.installed {
		s.ctx.LogInfof("Uninstalling %s", s.Name())

		cmd := helpers.Command{
			Command: "/bin/bash",
			Args:    []string{"-c", "\"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/uninstall.sh)\""},
		}

		_, err := helpers.ExecuteWithNoOutput(cmd)
		if err != nil {
			return err
		}
	}

	if !uninstallDependencies {
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

func (s *SystemService) Dependencies() []interfaces.Service {
	if s.dependencies == nil {
		s.dependencies = []interfaces.Service{}
	}
	return s.dependencies
}

func (s *SystemService) SetDependencies(dependencies []interfaces.Service) {
	s.dependencies = dependencies
}

func (s *SystemService) Installed() bool {
	return s.installed && s.brewExecutable != ""
}

func (s *SystemService) GetSystemUsers(ctx basecontext.ApiContext) ([]models.SystemUser, error) {
	switch s.GetOperatingSystem() {
	case "macos":
		return s.getMacSystemUsers(ctx)
	case "linux":
		return s.getLinuxSystemUsers(ctx)
	default:
		return nil, errors.New("Not implemented")
	}
}

func (s *SystemService) getMacSystemUsers(ctx basecontext.ApiContext) ([]models.SystemUser, error) {
	result := make([]models.SystemUser, 0)

	out, err := commands.ExecuteWithNoOutput("dscl", ".", "list", "/Users")
	if err != nil {
		return nil, err
	}

	users := strings.Split(out, "\n")
	for _, user := range users {
		user = strings.TrimSpace(user)
		if user == "" {
			continue
		}
		if strings.HasPrefix(user, "_") {
			continue
		}
		if strings.HasPrefix(user, "daemon") {
			continue
		}
		if strings.HasPrefix(user, "nobody") {
			continue
		}

		userHomeDir, err := s.GetUserHome(ctx, user)
		if err != nil {
			continue
		}

		if _, err := os.Stat(userHomeDir); os.IsNotExist(err) {
			continue
		} else {
			result = append(result, models.SystemUser{
				Username: user,
				Home:     userHomeDir,
			})
		}
	}

	return result, nil
}

func (s *SystemService) getLinuxSystemUsers(ctx basecontext.ApiContext) ([]models.SystemUser, error) {
	result := make([]models.SystemUser, 0)

	usersCmdOut := ""
	out, err := commands.ExecuteWithNoOutput("/bin/getent", "passwd")
	if err != nil {
		out, err := commands.ExecuteWithNoOutput("cat", "/etc/passwd")
		if err != nil {
			return nil, err
		} else {
			usersCmdOut = out
		}
	} else {
		usersCmdOut = out
	}

	users := strings.Split(usersCmdOut, "\n")
	for _, user := range users {
		user = strings.TrimSpace(user)
		if user == "" {
			continue
		}
		parts := strings.Split(user, ":")
		if len(parts) < 6 {
			continue
		}
		userHomeDir := parts[5]
		if _, err := os.Stat(userHomeDir); os.IsNotExist(err) {
			continue
		} else {
			result = append(result, models.SystemUser{
				Username: parts[0],
				Home:     userHomeDir,
			})
		}
	}

	return result, nil
}

func (s *SystemService) GetOperatingSystem() string {
	runningOs := ""
	switch os := runtime.GOOS; os {
	case "darwin":
		runningOs = "macos"
	case "linux":
		runningOs = "linux"
	case "windows":
		runningOs = "windows"
	default:
		runningOs = "unknown"
	}

	return runningOs
}

func (s *SystemService) GetUserHome(ctx basecontext.ApiContext, user string) (string, error) {
	switch s.GetOperatingSystem() {
	case "macos":
		return s.getUserHomeMac(ctx, user)
	case "linux":
		return s.getUserHomeLinux(ctx, user)
	default:
		return "", errors.New("Not implemented")
	}
}

func (s *SystemService) getUserHomeMac(ctx basecontext.ApiContext, user string) (string, error) {
	out, err := commands.ExecuteWithNoOutput("dscl", ".", "read", "/Users/"+user, "NFSHomeDirectory")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.ReplaceAll(out, "NFSHomeDirectory:", "")), nil
}

func (s *SystemService) getUserHomeLinux(ctx basecontext.ApiContext, user string) (string, error) {
	usersCmdOut := ""
	out, err := commands.ExecuteWithNoOutput("/bin/getent", "passwd")
	if err != nil {
		out, err := commands.ExecuteWithNoOutput("cat", "/etc/passwd")
		if err != nil {
			return "", err
		} else {
			usersCmdOut = out
		}
	} else {
		usersCmdOut = out
	}

	parts := strings.Split(usersCmdOut, ":")
	if len(parts) < 6 {
		return "", errors.New("Invalid passwd file")
	}

	return parts[5], nil
}

func (s *SystemService) GetUserId(ctx basecontext.ApiContext, user string) (int, error) {
	switch s.GetOperatingSystem() {
	case "macos":
		return s.getUserIdMac(ctx, user)
	case "linux":
		return s.getUserIdLinux(ctx, user)
	default:
		return -1, errors.New("Not implemented")
	}
}

func (s *SystemService) getUserIdMac(ctx basecontext.ApiContext, user string) (int, error) {
	out, err := commands.ExecuteWithNoOutput("dscl", ".", "read", "/Users/"+user, "UniqueID")
	if err != nil {
		return -1, err
	}

	strId := strings.TrimSpace(strings.ReplaceAll(out, "UniqueID:", ""))
	if strId == "" {
		return -1, errors.New("User not found")
	}

	id, err := strconv.Atoi(strId)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (s *SystemService) getUserIdLinux(ctx basecontext.ApiContext, user string) (int, error) {
	out, err := commands.ExecuteWithNoOutput("/bin/id", "-u", user)
	if err != nil {
		return -1, err
	}

	strId := strings.TrimSpace(out)
	if strId == "" {
		return -1, errors.New("User not found")
	}

	id, err := strconv.Atoi(strId)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (s *SystemService) GetCurrentUser(ctx basecontext.ApiContext) (string, error) {
	switch s.GetOperatingSystem() {
	case "macos":
		return s.getMacCurrentUser(ctx)
	case "linux":
		return s.getLinuxCurrentUser(ctx)
	default:
		return "", errors.New("Not implemented")
	}
}

func (s *SystemService) getMacCurrentUser(ctx basecontext.ApiContext) (string, error) {
	out, err := commands.ExecuteWithNoOutput("whoami")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

func (s *SystemService) getLinuxCurrentUser(ctx basecontext.ApiContext) (string, error) {
	user, exists := os.LookupEnv("USER")
	if user != "" && !exists {
		user = "root"
	}

	return user, nil
}

func (s *SystemService) GetUniqueId(ctx basecontext.ApiContext) (string, error) {
	switch s.GetOperatingSystem() {
	case "macos":
		return s.getUniqueIdMac(ctx)
	case "linux":
		return s.getUniqueIdLinux(ctx)
	default:
		return "", errors.New("Not implemented")
	}
}

func (s *SystemService) getUniqueIdMac(ctx basecontext.ApiContext) (string, error) {
	out, err := commands.ExecuteWithNoOutput("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	if err != nil {
		return "", err
	}

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "IOPlatformUUID") {
			parts := strings.Split(line, "=")
			if len(parts) < 2 {
				return "", errors.New("Invalid IOPlatformUUID")
			}

			return strings.TrimSpace(parts[1]), nil
		}
	}

	return "", errors.New("IOPlatformUUID not found")
}

func (s *SystemService) getUniqueIdLinux(ctx basecontext.ApiContext) (string, error) {
	out, err := commands.ExecuteWithNoOutput("cat", "/etc/machine-id")
	if err != nil {
		return "", err
	}

	return strings.Trim(strings.TrimSpace(out), "\n"), nil
}

func (s *SystemService) ChangeFileUserOwner(ctx basecontext.ApiContext, userName string, filePath string) error {
	ctx.LogDebugf("Changing file %s owner to %s", filePath, userName)
	switch s.GetOperatingSystem() {
	case "macos":
		return s.changeMacFileUserOwner(userName, filePath)
	case "linux":
		return s.changeLinuxFileUserOwner(userName, filePath)
	default:
		return errors.New("Not implemented")
	}
}

func (s *SystemService) changeMacFileUserOwner(userName string, filePath string) error {
	_, err := commands.ExecuteWithNoOutput("chown", "-R", userName, filePath)
	if err != nil {
		return err
	}

	return nil
}

func (s *SystemService) changeLinuxFileUserOwner(userName string, filePath string) error {
	_, err := commands.ExecuteWithNoOutput("chown", "-R", userName, filePath)
	if err != nil {
		return err
	}

	return nil
}

func (s *SystemService) GetHardwareInfo(ctx basecontext.ApiContext) (*models.SystemHardwareInfo, error) {
	switch s.GetOperatingSystem() {
	case "macos":
		return s.getMacSystemHardwareInfo(ctx)
	case "linux":
		return nil, errors.New("Not implemented")
	default:
		return nil, errors.New("Not implemented")
	}
}

func (s *SystemService) getMacSystemHardwareInfo(ctx basecontext.ApiContext) (*models.SystemHardwareInfo, error) {
	result := models.SystemHardwareInfo{}
	cpuBrandNameCmd := helpers.Command{
		Command: "sysctl",
		Args:    []string{"-n", "machdep.cpu.brand_string"},
	}
	cpuTypeCmd := helpers.Command{
		Command: "uname",
		Args:    []string{"-m"},
	}
	physicalCpuCountCmd := helpers.Command{
		Command: "sysctl",
		Args:    []string{"-n", "hw.physicalcpu"},
	}
	logicalCpuCountCmd := helpers.Command{
		Command: "sysctl",
		Args:    []string{"-n", "hw.logicalcpu"},
	}
	memorySizeCmd := helpers.Command{
		Command: "sysctl",
		Args:    []string{"-n", "hw.memsize"},
	}
	diskAvailableCmd := helpers.Command{
		Command: "df",
		Args:    []string{"-h", "/"},
	}
	cpuBrand, err := helpers.ExecuteWithNoOutput(cpuBrandNameCmd)
	if err != nil {
		return nil, err
	}
	cpuType, err := helpers.ExecuteWithNoOutput(cpuTypeCmd)
	if err != nil {
		return nil, err
	}
	physicalCpuCount, err := helpers.ExecuteWithNoOutput(physicalCpuCountCmd)
	if err != nil {
		return nil, err
	}
	logicalCpuCount, err := helpers.ExecuteWithNoOutput(logicalCpuCountCmd)
	if err != nil {
		return nil, err
	}
	memorySize, err := helpers.ExecuteWithNoOutput(memorySizeCmd)
	if err != nil {
		return nil, err
	}
	diskAvailable, err := helpers.ExecuteWithNoOutput(diskAvailableCmd)
	if err != nil {
		return nil, err
	}
	result.CpuType = strings.ReplaceAll(cpuType, "\n", "")
	result.CpuBrand = strings.ReplaceAll(cpuBrand, "\n", "")
	physicalCpuCountInt, err := strconv.Atoi(helpers.CleanOutputString(physicalCpuCount))
	if err != nil {
		return nil, err
	}
	result.PhysicalCpuCount = physicalCpuCountInt
	logicalCpuCountInt, err := strconv.Atoi(helpers.CleanOutputString(logicalCpuCount))
	if err != nil {
		return nil, err
	}
	result.LogicalCpuCount = logicalCpuCountInt
	memorySizeInt, err := strconv.ParseFloat(helpers.CleanOutputString(memorySize), 64)
	if err != nil {
		return nil, err
	}
	result.MemorySize = helpers.ConvertByteToMegabyte(memorySizeInt)
	totalDiskInt, diskAvailableInt, err := s.parseDfCommand(diskAvailable)
	if err != nil {
		return nil, err
	}
	result.DiskSize = helpers.ConvertByteToMegabyte(totalDiskInt)
	result.FreeDiskSize = helpers.ConvertByteToMegabyte(diskAvailableInt)

	return &result, nil
}

func (s *SystemService) parseDfCommand(output string) (totalDisk float64, freeDisk float64, err error) {
	lines := strings.Split(output, "\n")
	if len(lines) > 1 {
		fields := strings.Fields(lines[1])
		if len(fields) > 2 {
			totalDisk, err = helpers.GetSizeByteFromString(helpers.CleanOutputString(fields[1]))
			if err != nil {
				return -1, -1, err
			}
			freeDisk, err = helpers.GetSizeByteFromString(helpers.CleanOutputString(fields[2]))
			if err != nil {
				return -1, -1, err
			}
		}
	}

	return totalDisk, freeDisk, nil
}

func (s *SystemService) GetArchitecture(ctx basecontext.ApiContext) (string, error) {
	switch s.GetOperatingSystem() {
	case "macos":
		return s.getMacArchitecture(ctx)
	case "linux":
		return s.getLinuxArchitecture(ctx)
	default:
		return "", errors.New("Not implemented")
	}
}

func (s *SystemService) getMacArchitecture(ctx basecontext.ApiContext) (string, error) {
	cpuTypeCmd := helpers.Command{
		Command: "uname",
		Args:    []string{"-m"},
	}
	cpuType, err := helpers.ExecuteWithNoOutput(cpuTypeCmd)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(cpuType, "\n", ""), nil
}

func (s *SystemService) getLinuxArchitecture(ctx basecontext.ApiContext) (string, error) {
	cpuTypeCmd := helpers.Command{
		Command: "uname",
		Args:    []string{"-m"},
	}
	cpuType, err := helpers.ExecuteWithNoOutput(cpuTypeCmd)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(cpuType, "\n", ""), nil
}
