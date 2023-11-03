package startup

import (
	"Parallels/pd-api-service/config"
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/controllers"
	"Parallels/pd-api-service/restapi"
	"Parallels/pd-api-service/serviceprovider"
)

var listener *restapi.HttpListener

func InitApi() *restapi.HttpListener {
	listener = restapi.GetHttpListener()
	cfg := config.NewConfig()
	listener.Options.ApiPrefix = cfg.GetApiPrefix()
	listener.Options.HttpPort = cfg.GetApiPort()
	if cfg.TLSEnabled() {
		listener.Options.EnableTLS = true
		listener.Options.TLSCertificate = cfg.GetTlsCertificate()
		listener.Options.TLSPrivateKey = cfg.GetTlsPrivateKey()
		listener.Options.TLSPort = cfg.GetTLSPort()
	}

	listener.AddJsonContent().AddLogger().AddHealthCheck()
	listener.WithPublicUserRegistration()
	RegisterV1Controllers()

	return listener
}

func ResetApi() {
	listener.WaitAndShutdown()
}

func RegisterV1Controllers() {
	provider := serviceprovider.Get()

	// Authorization Controller
	listener.AddController(controllers.GetTokenController(), "/v1/auth/token", "POST")
	listener.AddController(controllers.ValidateTokenController(), "/v1/auth/token/validate", "POST")

	// Users Controller
	listener.AddAuthorizedControllerWithRoles(controllers.GetUsersController(), "/v1/auth/users", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.CreateUserController(), "/v1/auth/users", []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.GetUserController(), "/v1/auth/users/{id}", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.DeleteUserController(), "/v1/auth/users/{id}", []string{constants.SUPER_USER_ROLE}, "DELETE")
	listener.AddAuthorizedControllerWithRoles(controllers.UpdateUserController(), "/v1/auth/users/{id}", []string{constants.SUPER_USER_ROLE}, "PUT")
	listener.AddAuthorizedControllerWithRoles(controllers.GetUserRolesController(), "/v1/auth/users/{id}/roles", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.AddRoleToUserController(), "/v1/auth/users/{id}/roles", []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.RemoveRoleFromUserController(), "/v1/auth/users/{id}/roles/{role_id}", []string{constants.SUPER_USER_ROLE}, "DELETE")
	listener.AddAuthorizedControllerWithRoles(controllers.GetUserClaimsController(), "/v1/auth/users/{id}/claims", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.AddClaimToUserController(), "/v1/auth/users/{id}/claims", []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.RemoveClaimFromUserController(), "/v1/auth/users/{id}/claims/{role_id}", []string{constants.SUPER_USER_ROLE}, "DELETE")

	// Claims Controller
	listener.AddAuthorizedControllerWithRoles(controllers.GetClaimsController(), "/v1/auth/claims", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.CreateClaimController(), "/v1/auth/claims", []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.GetClaimController(), "/v1/auth/claims/{id}", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.DeleteClaimController(), "/auth/claims/{id}", []string{constants.SUPER_USER_ROLE}, "DELETE")

	// Roles Controller
	listener.AddAuthorizedControllerWithRoles(controllers.GetRolesController(), "/v1/auth/roles", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.CreateRoleController(), "/v1/auth/roles", []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.GetRoleController(), "/v1/auth/roles/{id}", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.DeleteRoleController(), "/v1/auth/roles/{id}", []string{constants.SUPER_USER_ROLE}, "DELETE")

	// ApiKey Controller
	listener.AddAuthorizedControllerWithRoles(controllers.GetApiKeysController(), "/v1/auth/api_keys", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.CreateApiKeyController(), "/v1/auth/api_keys", []string{constants.SUPER_USER_ROLE}, "POST")
	listener.AddAuthorizedControllerWithRoles(controllers.GetApiKeyByIdOrNameController(), "/v1/auth/api_keys/{id}", []string{constants.SUPER_USER_ROLE}, "GET")
	listener.AddAuthorizedControllerWithRoles(controllers.DeleteApiKeyController(), "/v1/auth/api_keys/{id}", []string{constants.SUPER_USER_ROLE}, "DELETE")
	listener.AddAuthorizedControllerWithRoles(controllers.RevokeApiKeyController(), "/v1/auth/api_keys/{id}/revoke", []string{constants.SUPER_USER_ROLE}, "PUT")

	// Config Controller
	if provider.System.GetOperatingSystem() == "macos" {
		listener.AddAuthorizedControllerWithRoles(controllers.InstallToolsController(), "/v1/config/tools/install", []string{constants.SUPER_USER_ROLE}, "POST")
		listener.AddAuthorizedControllerWithRoles(controllers.UninstallToolsController(), "/v1/config/tools/uninstall", []string{constants.SUPER_USER_ROLE}, "POST")
	}
	listener.AddAuthorizedControllerWithRoles(controllers.RestartController(), "/v1/config/restart", []string{constants.SUPER_USER_ROLE}, "POST")

	// Virtual Machines Controller
	if provider.IsParallelsDesktopAvailable() {
		listener.AddAuthorizedControllerWithClaims(controllers.GetParallelsDesktopLicenseController(), "/v1/parallels_desktop/key", []string{constants.LIST_VM_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.GetMachinesController(), "/v1/machines", []string{constants.LIST_VM_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.CreateMachine(), "/v1/machines", []string{constants.CREATE_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.RegisterMachineController(), "/v1/machines/register", []string{constants.CREATE_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.DeleteMachineController(), "/v1/machines/{id}", []string{constants.DELETE_VM_CLAIM}, "DELETE")
		listener.AddAuthorizedControllerWithClaims(controllers.GetMachineController(), "/v1/machines/{id}", []string{constants.LIST_VM_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.StartMachineController(), "/v1/machines/{id}/start", []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.StopMachineController(), "/v1/machines/{id}/stop", []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.RestartMachineController(), "/v1/machines/{id}/restart", []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.SuspendMachineController(), "/v1/machines/{id}/suspend", []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.ResetMachineController(), "/v1/machines/{id}/reset", []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.PauseMachineController(), "/v1/machines/{id}/pause", []string{constants.UPDATE_VM_STATES_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.StatusMachineController(), "/v1/machines/{id}/status", []string{constants.LIST_VM_CLAIM}, "GET")
		listener.AddAuthorizedControllerWithClaims(controllers.SetMachineController(), "/v1/machines/{id}/set", []string{constants.UPDATE_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.ExecuteCommandOnMachineController(), "/v1/machines/{id}/execute", []string{constants.EXECUTE_COMMAND_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.RenameMachineController(), "/v1/machines/{id}/rename", []string{constants.UPDATE_VM_CLAIM}, "POST")
		listener.AddAuthorizedControllerWithClaims(controllers.UnregisterMachineController(), "/v1/machines/{id}/unregister", []string{constants.UPDATE_VM_CLAIM}, "POST")
	}

	// Packer Templates Controller
	listener.AddAuthorizedControllerWithClaims(controllers.GetPackerTemplatesController(), "/v1/templates/packer", []string{constants.LIST_PACKER_TEMPLATE_CLAIM}, "GET")
	listener.AddAuthorizedControllerWithClaims(controllers.CreatePackerTemplateController(), "/v1/templates/packer", []string{constants.CREATE_PACKER_TEMPLATE_CLAIM}, "POST")
	listener.AddAuthorizedControllerWithClaims(controllers.UpdatePackerTemplateController(), "/v1/templates/packer/{id}", []string{constants.UPDATE_PACKER_TEMPLATE_CLAIM}, "PUT")
	listener.AddAuthorizedControllerWithClaims(controllers.GetPackerTemplateController(), "/v1/templates/packer/{id}", []string{constants.LIST_VM_CLAIM}, "GET")
	listener.AddAuthorizedControllerWithClaims(controllers.DeletePackerTemplateController(), "/v1/templates/packer/{id}", []string{constants.LIST_VM_CLAIM}, "DELETE")

	// Remote Machines Catalog Controller
	listener.AddAuthorizedControllerWithClaims(controllers.GetCatalogManifestsController(), "/v1/catalog", []string{constants.LIST_CATALOG_MANIFEST_CLAIM}, "GET")
	listener.AddAuthorizedControllerWithClaims(controllers.CreateCatalogManifestController(), "/v1/catalog", []string{constants.CREATE_CATALOG_MANIFEST_CLAIM}, "POST")
	listener.AddAuthorizedControllerWithClaims(controllers.DeleteCatalogManifestController(), "/v1/catalog/{id}", []string{constants.DELETE_CATALOG_MANIFEST_CLAIM}, "DELETE")
	listener.AddAuthorizedControllerWithClaims(controllers.GetCatalogManifestController(), "/v1/catalog/{id}", []string{constants.LIST_CATALOG_MANIFEST_CLAIM}, "GET")
	listener.AddAuthorizedControllerWithClaims(controllers.PushCatalogManifestController(), "/v1/catalog/push", []string{constants.PUSH_CATALOG_MANIFEST_CLAIM}, "POST")
	listener.AddAuthorizedControllerWithClaims(controllers.PullCatalogManifestController(), "/v1/catalog/pull", []string{constants.PULL_CATALOG_MANIFEST_CLAIM}, "PUT")
	listener.AddAuthorizedControllerWithClaims(controllers.ImportCatalogManifestController(), "/v1/catalog/import", []string{constants.PULL_CATALOG_MANIFEST_CLAIM}, "PUT")
}
