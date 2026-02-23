package system

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider/interfaces"
	"github.com/cjlapao/common-go/version"
)

var VersionSvc = version.Get()

var globalSystemService *SystemService

type SystemServiceCache struct {
	IsCached                     bool
	SystemUsers                  []models.SystemUser
	CurrentUser                  string
	CurrentUserHome              string
	UniqueId                     string
	HardwareInfo                 *models.SystemHardwareInfo
	OperatingSystem              string
	Architecture                 string
	ExternalIpAddress            string
	LastUpdatedExternalIpAddress int64
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
		ctx.LogDebugf("[SYSTEM] Returning cached system users")
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

	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
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

	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), usersCmd, helpers.ExecutionTimeout)
	if err != nil {
		catCommand := helpers.Command{
			Command: "cat",
			Args:    []string{"/etc/passwd"},
		}
		out, err := helpers.ExecuteWithNoOutput(ctx.Context(), catCommand, helpers.ExecutionTimeout)
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
	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
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
	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		catCmd := helpers.Command{
			Command: "cat",
			Args:    []string{"/etc/passwd"},
		}
		out, err := helpers.ExecuteWithNoOutput(ctx.Context(), catCmd, helpers.ExecutionTimeout)
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
	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
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
	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
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
		ctx.LogDebugf("[SYSTEM] Returning cached current user")
		return s.cache.CurrentUser, nil
	}

	currentUser, err := helpers.GetCurrentSystemUser()
	if err != nil {
		return "", err
	}

	s.cache.CurrentUser = currentUser
	return currentUser, nil
}

func (s *SystemService) GetUniqueId(ctx basecontext.ApiContext) (string, error) {
	if s.cache.UniqueId != "" {
		ctx.LogDebugf("[SYSTEM] Returning cached unique id")
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
	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
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

	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
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

	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
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
	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
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
	_, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *SystemService) GetHardwareInfo(ctx basecontext.ApiContext) (*models.SystemHardwareInfo, error) {
	if s.cache.HardwareInfo != nil {
		ctx.LogDebugf("[SYSTEM] Returning cached hardware info")
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
	cpuBrand, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cpuBrandNameCmd, helpers.ExecutionTimeout)
	if err != nil {
		return nil, err
	}
	cpuType, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cpuTypeCmd, helpers.ExecutionTimeout)
	if err != nil {
		return nil, err
	}
	physicalCpuCount, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), physicalCpuCountCmd, helpers.ExecutionTimeout)
	if err != nil {
		return nil, err
	}
	logicalCpuCount, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), logicalCpuCountCmd, helpers.ExecutionTimeout)
	if err != nil {
		return nil, err
	}
	memorySize, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), memorySizeCmd, helpers.ExecutionTimeout)
	if err != nil {
		return nil, err
	}
	diskAvailable, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), diskAvailableCmd, helpers.ExecutionTimeout)
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
		ctx.LogDebugf("[SYSTEM] Returning cached architecture")
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
	cpuType, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cpuTypeCmd, helpers.ExecutionTimeout)
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
	cpuType, err := helpers.ExecuteWithNoOutput(s.ctx.Context(), cpuTypeCmd, helpers.ExecutionTimeout)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(cpuType, "\n", ""), nil
}

func (s *SystemService) GetHardwareUsage(ctx basecontext.ApiContext) (*models.SystemUsageResponse, error) {
	result := &models.SystemUsageResponse{}

	arch, err := s.GetArchitecture(ctx)
	if err != nil {
		arch = err.Error()
	}

	result.CpuType = arch
	result.DevOpsVersion = VersionSvc.String()

	external_ip, err := s.GetExternalIp(ctx)
	if err == nil {
		result.ExternalIpAddress = external_ip
	}
	os_version, err := s.GetOsVersion(ctx)
	if err == nil {
		result.OsVersion = os_version
	}

	result.OsName = s.GetOSName()

	return result, nil
}

func (s *SystemService) GetExternalIp(ctx basecontext.ApiContext) (string, error) {
	os := s.GetOperatingSystem()
	switch os {
	case "macos":
		return s.getUniversalExternalIp(ctx)
	case "linux":
		return s.getUniversalExternalIp(ctx)
	case "windows":
		return s.getUniversalExternalIp(ctx)
	default:
		return "", fmt.Errorf("operating System %s not implemented yet", os)
	}
}

func (s *SystemService) getUniversalExternalIp(ctx basecontext.ApiContext) (string, error) {
	// First we check if we already have the external ip cached and if the cache is not expired
	if s.cache.ExternalIpAddress != "" && s.cache.LastUpdatedExternalIpAddress > 0 {
		currentTime := time.Now().Unix()
		if currentTime-s.cache.LastUpdatedExternalIpAddress < 2*3600 {
			ctx.LogDebugf("[SYSTEM] Returning cached external IP address")
			return s.cache.ExternalIpAddress, nil
		}
	}

	// first lets try to get the external ip using the ifconfig.me service and curl
	cmd := helpers.Command{
		Command: "curl",
		Args:    []string{"ifconfig.me"},
	}
	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		cmd = helpers.Command{
			Command: "wget",
			Args:    []string{"-qO-", "ifconfig.me"},
		}
		out, err = helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
		if err != nil {
			return "", err
		}
	}

	s.cache.ExternalIpAddress = strings.TrimSpace(out)
	s.cache.LastUpdatedExternalIpAddress = time.Now().Unix()
	return strings.TrimSpace(out), nil
}

func (s *SystemService) GetOsVersion(ctx basecontext.ApiContext) (string, error) {
	os := s.GetOperatingSystem()
	switch os {
	case "macos":
		return s.getOsVersionForMac(ctx)
	case "linux":
		return s.getOsVersionForLinux(ctx)
	case "windows":
		return s.getOsVersionForWindows(ctx)
	default:
		return "", fmt.Errorf("operating System %s not implemented yet", os)
	}
}

func (s *SystemService) getOsVersionForMac(ctx basecontext.ApiContext) (string, error) {
	// first lets try to get the external ip using the ifconfig.me service and curl
	cmd := helpers.Command{
		Command: "sw_vers",
		Args:    []string{"-productVersion"},
	}

	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

func (s *SystemService) getOsVersionForLinux(ctx basecontext.ApiContext) (string, error) {
	// Let's first try to parse /etc/os-release
	if content, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(content), "\n")
		var versionId string
		var buildId string
		for _, line := range lines {
			if strings.HasPrefix(line, "VERSION_ID=") {
				versionId = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
			} else if strings.HasPrefix(line, "BUILD_ID=") {
				buildId = strings.Trim(strings.TrimPrefix(line, "BUILD_ID="), "\"")
			}
		}
		if versionId != "" {
			if buildId != "" {
				return fmt.Sprintf("%s (%s)", versionId, buildId), nil
			}
			return versionId, nil
		}
	}

	// Fallback to redhat-release
	if content, err := os.ReadFile("/etc/redhat-release"); err == nil {
		out := string(content)
		if out != "" {
			return strings.TrimSpace(out), nil
		}
	}

	// Ultimate fallback to kernel release
	cmd := helpers.Command{
		Command: "uname",
		Args:    []string{"-r"},
	}
	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

func (s *SystemService) getOsVersionForWindows(ctx basecontext.ApiContext) (string, error) {
	// first lets try to get the external ip using the ifconfig.me service and curl
	cmd := helpers.Command{
		Command: "ver",
	}

	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(out), "\r\n", ""), "\n", ""), nil
}

func (s *SystemService) GetOsTempFolder(ctx basecontext.ApiContext) (string, error) {
	os := s.GetOperatingSystem()
	switch os {
	case "macos":
		return s.getOsTempFolderFoMac(ctx)
	case "linux":
		return s.getOsTempFolderForLinux(ctx)
	case "windows":
		return s.getOsTempFolderForWindows(ctx)
	default:
		return "", fmt.Errorf("operating System %s not implemented yet", os)
	}
}

func (s *SystemService) getOsTempFolderFoMac(ctx basecontext.ApiContext) (string, error) {
	cmd := helpers.Command{
		Command: "echo",
		Args:    []string{"$TMPDIR"},
	}

	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return "", err
	}

	if out == "" {
		out = "/tmp"
	}

	return strings.TrimSpace(out), nil
}

func (s *SystemService) getOsTempFolderForLinux(ctx basecontext.ApiContext) (string, error) {
	cmd := helpers.Command{
		Command: "echo",
		Args:    []string{"$TMPDIR"},
	}

	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return "", err
	}

	if out == "" {
		out = "/tmp"
	}

	return strings.TrimSpace(out), nil
}

func (s *SystemService) getOsTempFolderForWindows(ctx basecontext.ApiContext) (string, error) {
	cmd := helpers.Command{
		Command: "echo",
		Args:    []string{"%TEMP%"},
	}

	out, err := helpers.ExecuteWithNoOutput(ctx.Context(), cmd, helpers.ExecutionTimeout)
	if err != nil {
		return "", err
	}

	if out == "" {
		out = "/tmp"
	}

	return strings.TrimSpace(out), nil
}

func (s *SystemService) GetOSName() string {
	goOs := runtime.GOOS
	switch goOs {
	case "darwin":
		return "macOS"
	case "linux":
		// Let's try to parse /etc/os-release for a pretty name
		if content, err := os.ReadFile("/etc/os-release"); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					name := strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
					if name != "" {
						return name
					}
				} else if strings.HasPrefix(line, "NAME=") {
					name := strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
					if name != "" {
						return name
					}
				}
			}
		}

		// Fallback to redhat-release
		if content, err := os.ReadFile("/etc/redhat-release"); err == nil {
			out := string(content)
			if out != "" {
				parts := strings.Split(strings.TrimSpace(out), " release ")
				if len(parts) > 0 {
					return parts[0]
				}
			}
		}

		return "Linux"
	case "windows":
		return "Windows"
	default:
		return "Unknown"
	}
}
