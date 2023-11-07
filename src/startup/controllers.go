package startup

import (
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/controllers"
	"github.com/Parallels/pd-api-service/restapi"
	"github.com/Parallels/pd-api-service/serviceprovider"
	"github.com/cjlapao/common-go/helper/http_helper"
)

var listener *restapi.HttpListener

func InitApi() *restapi.HttpListener {
	listener = restapi.GetHttpListener()
	cfg := config.NewConfig()
	listener.Options.ApiPrefix = cfg.GetApiPrefix()
	listener.Options.HttpPort = cfg.GetApiPort()
	listener.Options.DefaultApiVersion = "v1"
	if cfg.TLSEnabled() {
		listener.Options.EnableTLS = true
		listener.Options.TLSCertificate = cfg.GetTlsCertificate()
		listener.Options.TLSPrivateKey = cfg.GetTlsPrivateKey()
		listener.Options.TLSPort = cfg.GetTLSPort()
	}

	listener.AddSwagger()
	listener.AddJsonContent().AddLogger().AddHealthCheck()
	listener.WithPublicUserRegistration()
	RegisterDefaultControllers()
	RegisterV1Controllers()

	return listener
}

func ResetApi() {
	listener.WaitAndShutdown()
}

func RegisterDefaultControllers() {
	provider := serviceprovider.Get()

	// Authorization Controller
	listener.AddController(controllers.GetTokenController(), "/auth/token", "POST")
	listener.AddController(controllers.ValidateTokenController(), "/auth/token/validate", "POST")

	// Users Controller
	listener.AddAuthorizedControllerWithRoles(controllers.GetUsersController(), "/auth/users", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.CreateUserController(), "/auth/users", []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.GetUserController(), "/auth/users/{id}", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.DeleteUserController(), "/auth/users/{id}", []string{constants.SUPER_USER_ROLE}, "DELETE")
	listener.AddAuthorizedControllerWithRoles(controllers.UpdateUserController(), "/auth/users/{id}", []string{constants.SUPER_USER_ROLE}, "PUT")
	listener.AddAuthorizedControllerWithRoles(controllers.GetUserRolesController(), "/auth/users/{id}/roles", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.AddRoleToUserController(), "/auth/users/{id}/roles", []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.RemoveRoleFromUserController(), "/auth/users/{id}/roles/{role_id}", []string{constants.SUPER_USER_ROLE}, "DELETE")
	listener.AddAuthorizedControllerWithRoles(controllers.GetUserClaimsController(), "/auth/users/{id}/claims", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.AddClaimToUserController(), "/auth/users/{id}/claims", []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.RemoveClaimFromUserController(), "/auth/users/{id}/claims/{role_id}", []string{constants.SUPER_USER_ROLE}, "DELETE")

	// Claims Controller
	listener.AddAuthorizedControllerWithRoles(controllers.GetClaimsController(), "/auth/claims", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.CreateClaimController(), "/auth/claims", []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.GetClaimController(), "/auth/claims/{id}", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.DeleteClaimController(), "/auth/claims/{id}", []string{constants.SUPER_USER_ROLE}, "DELETE")

	// Roles Controller
	listener.AddAuthorizedControllerWithRoles(controllers.GetRolesController(), "/auth/roles", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.CreateRoleController(), "/auth/roles", []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.GetRoleController(), "/auth/roles/{id}", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.DeleteRoleController(), "/auth/roles/{id}", []string{constants.SUPER_USER_ROLE}, "DELETE")

	// ApiKey Controller
	listener.AddAuthorizedControllerWithRoles(controllers.GetApiKeysController(), "/auth/api_keys", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.CreateApiKeyController(), "/auth/api_keys", []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.GetApiKeyByIdOrNameController(), "/auth/api_keys/{id}", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.DeleteApiKeyController(), "/auth/api_keys/{id}", []string{constants.SUPER_USER_ROLE}, "DELETE")
	listener.AddAuthorizedControllerWithRoles(controllers.RevokeApiKeyController(), "/auth/api_keys/{id}/revoke", []string{constants.SUPER_USER_ROLE}, "PUT")

	// Config Controller
	if provider.System.GetOperatingSystem() == "macos" {
		listener.AddAuthorizedControllerWithRoles(controllers.InstallToolsController(), "/config/tools/install", []string{constants.SUPER_USER_ROLE}, "POST")
		listener.AddAuthorizedControllerWithRoles(controllers.UninstallToolsController(), "/config/tools/uninstall", []string{constants.SUPER_USER_ROLE}, "POST")
	}
	listener.AddAuthorizedControllerWithRoles(controllers.RestartController(), "/config/restart", []string{constants.SUPER_USER_ROLE}, "POST")

	// Virtual Machines Controller
	if provider.IsParallelsDesktopAvailable() {
		listener.AddAuthorizedControllerWithClaims(controllers.GetParallelsDesktopLicenseController(), "/parallels_desktop/key", []string{constants.LIST_VM_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.GetMachinesController(), "/machines", []string{constants.LIST_VM_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.CreateMachine(), "/machines", []string{constants.CREATE_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.RegisterMachineController(), "/machines/register", []string{constants.CREATE_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.DeleteMachineController(), "/machines/{id}", []string{constants.DELETE_VM_CLAIM}, "DELETE")
		listener.AddAuthorizedControllerWithClaims(controllers.GetMachineController(), "/machines/{id}", []string{constants.LIST_VM_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.StartMachineController(), "/machines/{id}/start", []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.StopMachineController(), "/machines/{id}/stop", []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.RestartMachineController(), "/machines/{id}/restart", []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.SuspendMachineController(), "/machines/{id}/suspend", []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.ResetMachineController(), "/machines/{id}/reset", []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.PauseMachineController(), "/machines/{id}/pause", []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.StatusMachineController(), "/machines/{id}/status", []string{constants.LIST_VM_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.SetMachineController(), "/machines/{id}/set", []string{constants.UPDATE_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.ExecuteCommandOnMachineController(), "/machines/{id}/execute", []string{constants.EXECUTE_COMMAND_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.RenameMachineController(), "/machines/{id}/rename", []string{constants.UPDATE_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.UnregisterMachineController(), "/machines/{id}/unregister", []string{constants.UPDATE_VM_CLAIM}, "POST")
	}

	// Packer Templates Controller
	listener.AddAuthorizedControllerWithClaims(controllers.GetPackerTemplatesController(), "/templates/packer", []string{constants.LIST_PACKER_TEMPLATE_CLAIM}, "GET")
	listener.AddAuthorizedControllerWithClaims(controllers.CreatePackerTemplateController(), "/templates/packer", []string{constants.CREATE_PACKER_TEMPLATE_CLAIM}, "POST")
	listener.AddAuthorizedControllerWithClaims(controllers.UpdatePackerTemplateController(), "/templates/packer/{id}", []string{constants.UPDATE_PACKER_TEMPLATE_CLAIM}, "PUT")
	listener.AddAuthorizedControllerWithClaims(controllers.GetPackerTemplateController(), "/templates/packer/{id}", []string{constants.LIST_VM_CLAIM}, "GET")
	listener.AddAuthorizedControllerWithClaims(controllers.DeletePackerTemplateController(), "/templates/packer/{id}", []string{constants.LIST_VM_CLAIM}, "DELETE")

	// Remote Machines Catalog Controller
	listener.AddAuthorizedControllerWithClaims(controllers.GetCatalogManifestsController(), "/catalog", []string{constants.LIST_CATALOG_MANIFEST_CLAIM}, "GET")
	listener.AddAuthorizedControllerWithClaims(controllers.CreateCatalogManifestController(), "/catalog", []string{constants.CREATE_CATALOG_MANIFEST_CLAIM}, "POST")
	listener.AddAuthorizedControllerWithClaims(controllers.DeleteCatalogManifestController(), "/catalog/{id}", []string{constants.DELETE_CATALOG_MANIFEST_CLAIM}, "DELETE")
	listener.AddAuthorizedControllerWithClaims(controllers.GetCatalogManifestController(), "/catalog/{id}", []string{constants.LIST_CATALOG_MANIFEST_CLAIM}, "GET")
	listener.AddAuthorizedControllerWithClaims(controllers.PushCatalogManifestController(), "/catalog/push", []string{constants.PUSH_CATALOG_MANIFEST_CLAIM}, "POST")
	listener.AddAuthorizedControllerWithClaims(controllers.PullCatalogManifestController(), "/catalog/pull", []string{constants.PULL_CATALOG_MANIFEST_CLAIM}, "PUT")
	listener.AddAuthorizedControllerWithClaims(controllers.ImportCatalogManifestController(), "/catalog/import", []string{constants.PULL_CATALOG_MANIFEST_CLAIM}, "PUT")
}

func RegisterV1Controllers() {
	provider := serviceprovider.Get()

	// Authorization Controller
	listener.AddController(controllers.GetTokenController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/token"), "POST")
	listener.AddController(controllers.ValidateTokenController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/token/validate"), "POST")

	// Users Controller
	listener.AddAuthorizedControllerWithRoles(controllers.GetUsersController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/users"), []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.CreateUserController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/users"), []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.GetUserController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/users/{id}"), []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.DeleteUserController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/users/{id}"), []string{constants.SUPER_USER_ROLE}, "DELETE")
	listener.AddAuthorizedControllerWithRoles(controllers.UpdateUserController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/users/{id}"), []string{constants.SUPER_USER_ROLE}, "PUT")
	listener.AddAuthorizedControllerWithRoles(controllers.GetUserRolesController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/users/{id}/roles"), []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.AddRoleToUserController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/users/{id}/roles"), []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.RemoveRoleFromUserController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/users/{id}/roles/{role_id}"), []string{constants.SUPER_USER_ROLE}, "DELETE")
	listener.AddAuthorizedControllerWithRoles(controllers.GetUserClaimsController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/users/{id}/claims"), []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.AddClaimToUserController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/users/{id}/claims"), []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.RemoveClaimFromUserController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/users/{id}/claims/{role_id}"), []string{constants.SUPER_USER_ROLE}, "DELETE")

	// Claims Controller
	listener.AddAuthorizedControllerWithRoles(controllers.GetClaimsController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/claims"), []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.CreateClaimController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/claims"), []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.GetClaimController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/claims/{id}"), []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.DeleteClaimController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/claims/{id}"), []string{constants.SUPER_USER_ROLE}, "DELETE")

	// Roles Controller
	listener.AddAuthorizedControllerWithRoles(controllers.GetRolesController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/roles"), []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.CreateRoleController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/roles"), []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.GetRoleController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/roles/{id}"), []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.DeleteRoleController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/roles/{id}"), []string{constants.SUPER_USER_ROLE}, "DELETE")

	// ApiKey Controller
	listener.AddAuthorizedControllerWithRoles(controllers.GetApiKeysController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/api_keys"), []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.CreateApiKeyController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/api_keys"), []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.GetApiKeyByIdOrNameController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/api_keys/{id}"), []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.DeleteApiKeyController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/api_keys/{id}"), []string{constants.SUPER_USER_ROLE}, "DELETE")
	listener.AddAuthorizedControllerWithRoles(controllers.RevokeApiKeyController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/auth/api_keys/{id}/revoke"), []string{constants.SUPER_USER_ROLE}, "PUT")

	// Config Controller
	if provider.System.GetOperatingSystem() == "macos" {
		listener.AddAuthorizedControllerWithRoles(controllers.InstallToolsController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/config/tools/install"), []string{constants.SUPER_USER_ROLE}, "POST")
		listener.AddAuthorizedControllerWithRoles(controllers.UninstallToolsController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/config/tools/uninstall"), []string{constants.SUPER_USER_ROLE}, "POST")
	}
	listener.AddAuthorizedControllerWithRoles(controllers.RestartController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/config/restart"), []string{constants.SUPER_USER_ROLE}, "POST")

	// Virtual Machines Controller
	if provider.IsParallelsDesktopAvailable() {
		listener.AddAuthorizedControllerWithClaims(controllers.GetParallelsDesktopLicenseController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/parallels_desktop/key"), []string{constants.LIST_VM_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.GetMachinesController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines"), []string{constants.LIST_VM_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.CreateMachine(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines"), []string{constants.CREATE_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.RegisterMachineController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines/register"), []string{constants.CREATE_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.DeleteMachineController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines/{id}"), []string{constants.DELETE_VM_CLAIM}, "DELETE")
		listener.AddAuthorizedControllerWithClaims(controllers.GetMachineController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines/{id}"), []string{constants.LIST_VM_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.StartMachineController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines/{id}/start"), []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.StopMachineController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines/{id}/stop"), []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.RestartMachineController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines/{id}/restart"), []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.SuspendMachineController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines/{id}/suspend"), []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.ResetMachineController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines/{id}/reset"), []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.PauseMachineController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines/{id}/pause"), []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.StatusMachineController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines/{id}/status"), []string{constants.LIST_VM_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.SetMachineController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines/{id}/set"), []string{constants.UPDATE_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.ExecuteCommandOnMachineController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines/{id}/execute"), []string{constants.EXECUTE_COMMAND_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.RenameMachineController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines/{id}/rename"), []string{constants.UPDATE_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.UnregisterMachineController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/machines/{id}/unregister"), []string{constants.UPDATE_VM_CLAIM}, "POST")
	}

	// Packer Templates Controller
	listener.AddAuthorizedControllerWithClaims(controllers.GetPackerTemplatesController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/templates/packer"), []string{constants.LIST_PACKER_TEMPLATE_CLAIM}, "GET")
	listener.AddAuthorizedControllerWithClaims(controllers.CreatePackerTemplateController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/templates/packer"), []string{constants.CREATE_PACKER_TEMPLATE_CLAIM}, "POST")
	listener.AddAuthorizedControllerWithClaims(controllers.UpdatePackerTemplateController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/templates/packer/{id}"), []string{constants.UPDATE_PACKER_TEMPLATE_CLAIM}, "PUT")
	listener.AddAuthorizedControllerWithClaims(controllers.GetPackerTemplateController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/templates/packer/{id}"), []string{constants.LIST_VM_CLAIM}, "GET")
	listener.AddAuthorizedControllerWithClaims(controllers.DeletePackerTemplateController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/templates/packer/{id}"), []string{constants.LIST_VM_CLAIM}, "DELETE")

	// Remote Machines Catalog Controller
	listener.AddAuthorizedControllerWithClaims(controllers.GetCatalogManifestsController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/catalog"), []string{constants.LIST_CATALOG_MANIFEST_CLAIM}, "GET")
	listener.AddAuthorizedControllerWithClaims(controllers.CreateCatalogManifestController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/catalog"), []string{constants.CREATE_CATALOG_MANIFEST_CLAIM}, "POST")
	listener.AddAuthorizedControllerWithClaims(controllers.DeleteCatalogManifestController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/catalog/{id}"), []string{constants.DELETE_CATALOG_MANIFEST_CLAIM}, "DELETE")
	listener.AddAuthorizedControllerWithClaims(controllers.GetCatalogManifestController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/catalog/{id}"), []string{constants.LIST_CATALOG_MANIFEST_CLAIM}, "GET")
	listener.AddAuthorizedControllerWithClaims(controllers.PushCatalogManifestController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/catalog/push"), []string{constants.PUSH_CATALOG_MANIFEST_CLAIM}, "POST")
	listener.AddAuthorizedControllerWithClaims(controllers.PullCatalogManifestController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/catalog/pull"), []string{constants.PULL_CATALOG_MANIFEST_CLAIM}, "PUT")
	listener.AddAuthorizedControllerWithClaims(controllers.ImportCatalogManifestController(), http_helper.JoinUrl(listener.Options.DefaultApiVersion, "/catalog/import"), []string{constants.PULL_CATALOG_MANIFEST_CLAIM}, "PUT")
}
