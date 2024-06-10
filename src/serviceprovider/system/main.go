package system

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider/interfaces"
)

var globalSystemService *SystemService

type SystemServiceCache struct {
	IsCached        bool
	SystemUsers     []models.SystemUser
	CurrentUser     string
	CurrentUserHome string
	UniqueId        string
	HardwareInfo    *models.SystemHardwareInfo
	OperatingSystem string
	Architecture    string
}

type SystemService struct {
	ctx            basecontext.ApiContext
	brewExecutable string
	installed      bool
	dependencies   []interfaces.Service
	cache          SystemServiceCache
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
		cache: SystemServiceCache{
			IsCached:     false,
			SystemUsers:  []models.SystemUser{},
			CurrentUser:  "",
			UniqueId:     "",
			HardwareInfo: nil,
			Architecture: "",
		},
	}

	globalSystemService.SetDependencies([]interfaces.Service{})
	return globalSystemService
}

func (s *SystemService) Name() string {
	return "system"
}

func (s *SystemService) FindPath() string {
	return "system"
}

func (s *SystemService) Version() string {
	return "latest"
}

func (s *SystemService) Install(asUser, version string, flags map[string]string) error {
	return nil
}

func (s *SystemService) Uninstall(asUser string, uninstallDependencies bool) error {
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
	return true
}

func (s *SystemService) GetSystemUsers(ctx basecontext.ApiContext) ([]models.SystemUser, error) {
	if s.cache.SystemUsers != nil && len(s.cache.SystemUsers) > 0 {
		ctx.LogDebugf("Returning cached system users")
		return s.cache.SystemUsers, nil
	}

	response := []models.SystemUser{}
	var err error
	switch s.GetOperatingSystem() {
	case "macos":
		response, err = s.getMacSystemUsers(ctx)
	case "linux":
		response, err = s.getLinuxSystemUsers(ctx)
	case "windows":
		response, err = s.getWindowsSystemUsers(ctx)
	default:
		return nil, errors.New("Not implemented")
	}

	s.cache.SystemUsers = response
	return response, err
}

func (s *SystemService) getMacSystemUsers(ctx basecontext.ApiContext) ([]models.SystemUser, error) {
	result := make([]models.SystemUser, 0)

	cmd := helpers.Command{
		Command: "dscl",
		Args:    []string{".", "list", "/Users"},
	}

	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd)
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

		if user == "root" {
			result = append(result, models.SystemUser{
				Username: user,
				Home:     "/root",
			})
		} else {
			if _, err := os.Stat(userHomeDir); os.IsNotExist(err) {
				continue
			} else {
				result = append(result, models.SystemUser{
					Username: user,
					Home:     userHomeDir,
				})
			}
		}
	}

	return result, nil
}

func (s *SystemService) getLinuxSystemUsers(ctx basecontext.ApiContext) ([]models.SystemUser, error) {
	result := make([]models.SystemUser, 0)

	usersCmd := helpers.Command{
		Command: "/bin/getent",
		Args:    []string{"passwd"},
	}
	usersCmdOut := ""

	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), usersCmd)
	if err != nil {
		catCommand := helpers.Command{
			Command: "cat",
			Args:    []string{"/etc/passwd"},
		}
		out, err := helpers.ExecuteWithNoOutput(ctx.Context(), catCommand)
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

func (s *SystemService) getWindowsSystemUsers(ctx basecontext.ApiContext) ([]models.SystemUser, error) {
	return []models.SystemUser{}, nil
}

func (s *SystemService) GetOperatingSystem() string {
	if s.cache.OperatingSystem != "" {
		return s.cache.OperatingSystem
	}

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

	s.cache.OperatingSystem = runningOs
	return runningOs
}

func (s *SystemService) GetUserHome(ctx basecontext.ApiContext, user string) (string, error) {
	switch s.GetOperatingSystem() {
	case "macos":
		return s.getUserHomeMac(ctx, user)
	case "linux":
		return s.getUserHomeLinux(ctx, user)
	case "windows":
		return s.getUserHomeWindows(ctx, user)
	default:
		return "", errors.New("Not implemented")
	}
}

func (s *SystemService) getUserHomeMac(ctx basecontext.ApiContext, user string) (string, error) {
	cmd := helpers.Command{
		Command: "dscl",
		Args:    []string{".", "read", "/Users/" + user, "NFSHomeDirectory"},
	}
	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.ReplaceAll(out, "NFSHomeDirectory:", "")), nil
}

func (s *SystemService) getUserHomeLinux(ctx basecontext.ApiContext, user string) (string, error) {
	usersCmdOut := ""
	cmd := helpers.Command{
		Command: "/bin/getent",
		Args:    []string{"passwd"},
	}
	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd)
	if err != nil {
		catCmd := helpers.Command{
			Command: "cat",
			Args:    []string{"/etc/passwd"},
		}
		out, err := helpers.ExecuteWithNoOutput(ctx.Context(), catCmd)
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

func (s *SystemService) getUserHomeWindows(ctx basecontext.ApiContext, user string) (string, error) {
	appData, exists := os.LookupEnv("USERNAME")
	if appData != "" && !exists {
		user = "/"
	}
	appData = strings.ReplaceAll(strings.ReplaceAll(appData, "\r\n", ""), "\n", "")

	return appData, nil
}

func (s *SystemService) GetUserId(ctx basecontext.ApiContext, user string) (int, error) {
	switch s.GetOperatingSystem() {
	case "macos":
		return s.getUserIdMac(ctx, user)
	case "linux":
		return s.getUserIdLinux(ctx, user)
	case "windows":
		return s.getUserIdWindows(ctx, user)
	default:
		return -1, errors.New("Not implemented")
	}
}

func (s *SystemService) getUserIdMac(ctx basecontext.ApiContext, user string) (int, error) {
	cmd := helpers.Command{
		Command: "dscl",
		Args:    []string{".", "read", "/Users/" + user, "UniqueID"},
	}
	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd)
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
	cmd := helpers.Command{
		Command: "/bin/id",
		Args:    []string{"-u", user},
	}
	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd)
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

func (s *SystemService) getUserIdWindows(ctx basecontext.ApiContext, user string) (int, error) {
	return 100, nil
}

func (s *SystemService) GetCurrentUser(ctx basecontext.ApiContext) (string, error) {
	if s.cache.CurrentUser != "" {
		ctx.LogDebugf("Returning cached current user")
		return s.cache.CurrentUser, nil
	}

	currentUser := ""
	var err error

	switch s.GetOperatingSystem() {
	case "macos":
		currentUser, err = s.getMacCurrentUser(ctx)
	case "linux":
		currentUser, err = s.getLinuxCurrentUser(ctx)
	case "windows":
		currentUser, err = s.getWindowsCurrentUser(ctx)
	default:
		return "", errors.New("Not implemented")
	}

	s.cache.CurrentUser = currentUser
	return currentUser, err
}

func (s *SystemService) getMacCurrentUser(ctx basecontext.ApiContext) (string, error) {
	cmd := helpers.Command{
		Command: "whoami",
	}
	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd)
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

func (s *SystemService) getWindowsCurrentUser(ctx basecontext.ApiContext) (string, error) {
	user, exists := os.LookupEnv("USERNAME")
	if user != "" && !exists {
		user = "root"
	}

	return user, nil
}

func (s *SystemService) GetUniqueId(ctx basecontext.ApiContext) (string, error) {
	if s.cache.UniqueId != "" {
		ctx.LogDebugf("Returning cached unique id")
		return s.cache.UniqueId, nil
	}

	uniqueId := ""
	var err error

	switch s.GetOperatingSystem() {
	case "macos":
		uniqueId, err = s.getUniqueIdMac(ctx)
	case "linux":
		uniqueId, err = s.getUniqueIdLinux(ctx)
	default:
		return "", errors.New("Not implemented")
	}

	s.cache.UniqueId = uniqueId
	return uniqueId, err
}

func (s *SystemService) getUniqueIdMac(ctx basecontext.ApiContext) (string, error) {
	cmd := helpers.Command{
		Command: "ioreg",
		Args:    []string{"-rd1", "-c", "IOPlatformExpertDevice"},
	}
	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd)
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
	cmd := helpers.Command{
		Command: "cat",
		Args:    []string{"/etc/machine-id"},
	}

	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd)
	if err != nil {
		return "", err
	}

	return strings.Trim(strings.TrimSpace(out), "\n"), nil
}

func (s *SystemService) getUniqueIdWindows(ctx basecontext.ApiContext) (string, error) {
	cmd := helpers.Command{
		Command: "wmic",
		Args:    []string{"path", "win32_computersystemproduct", "get", "UUID"},
	}

	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd)
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
	cmd := helpers.Command{
		Command: "chown",
		Args:    []string{"-R", userName, filePath},
	}
	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd)
	if err != nil {
		return err
	}

	return nil
}

func (s *SystemService) changeLinuxFileUserOwner(userName string, filePath string) error {
	cmd := helpers.Command{
		Command: "chown",
		Args:    []string{"-R", userName, filePath},
	}
	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd)
	if err != nil {
		return err
	}

	return nil
}

func (s *SystemService) GetHardwareInfo(ctx basecontext.ApiContext) (*models.SystemHardwareInfo, error) {
	if s.cache.HardwareInfo != nil {
		ctx.LogDebugf("Returning cached hardware info")
		return s.cache.HardwareInfo, nil
	}

	var response *models.SystemHardwareInfo
	var err error

	switch s.GetOperatingSystem() {
	case "macos":
		response, err = s.getMacSystemHardwareInfo(ctx)
	case "linux":
		return nil, errors.New("Not implemented")
	default:
		return nil, errors.New("Not implemented")
	}

	s.cache.HardwareInfo = response
	return response, err
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
	cpuBrand, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cpuBrandNameCmd)
	if err != nil {
		return nil, err
	}
	cpuType, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cpuTypeCmd)
	if err != nil {
		return nil, err
	}
	physicalCpuCount, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), physicalCpuCountCmd)
	if err != nil {
		return nil, err
	}
	logicalCpuCount, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), logicalCpuCountCmd)
	if err != nil {
		return nil, err
	}
	memorySize, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), memorySizeCmd)
	if err != nil {
		return nil, err
	}
	diskAvailable, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), diskAvailableCmd)
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
	if s.cache.Architecture != "" {
		ctx.LogDebugf("Returning cached architecture")
		return s.cache.Architecture, nil
	}

	response := ""
	var err error

	switch s.GetOperatingSystem() {
	case "macos":
		response, err = s.getMacArchitecture(ctx)
	case "linux":
		response, err = s.getLinuxArchitecture(ctx)
	default:
		return "", errors.New("Not implemented")
	}

	s.cache.Architecture = response
	return response, err
}

func (s *SystemService) getMacArchitecture(ctx basecontext.ApiContext) (string, error) {
	cpuTypeCmd := helpers.Command{
		Command: "uname",
		Args:    []string{"-m"},
	}
	cpuType, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cpuTypeCmd)
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
	cpuType, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cpuTypeCmd)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(cpuType, "\n", ""), nil
}
