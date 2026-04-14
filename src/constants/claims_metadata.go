package constants

// ── Group name constants ───────────────────────────────────────────────────
// Change these strings here and everywhere in the UI automatically reflects.
const (
	ClaimGroupAdministration = "Administration"
	ClaimGroupVMs            = "VMs"
	ClaimGroupCatalog        = "Catalog"
	ClaimGroupCatalogManager = "Catalog Manager"
	ClaimGroupReverseProxy   = "Reverse Proxy"
	ClaimGroupCache          = "Cache"
	ClaimGroupJobs           = "Jobs"
	ClaimGroupSSH            = "SSH"
	ClaimGroupCustom         = "Custom"
)

// ── Action constants ───────────────────────────────────────────────────────
// Standard CRUD actions map to the four matrix columns.
// Extended actions are rendered as extra columns or an "Other" section by the UI.
const (
	ClaimActionCreate    = "create"
	ClaimActionRead      = "read"
	ClaimActionUpdate    = "update"
	ClaimActionDelete    = "delete"
	ClaimActionExecute   = "execute"
	ClaimActionPull      = "pull"
	ClaimActionPush      = "push"
	ClaimActionImport    = "import"
	ClaimActionRevert    = "revert"
	ClaimActionConfigure = "configure"
)

// ClaimGroupOrder defines the canonical display order for groups in the matrix.
// Groups not listed here are appended alphabetically after the last entry.
var ClaimGroupOrder = []string{
	ClaimGroupAdministration,
	ClaimGroupVMs,
	ClaimGroupCatalog,
	ClaimGroupCatalogManager,
	ClaimGroupReverseProxy,
	ClaimGroupCache,
	ClaimGroupJobs,
	ClaimGroupSSH,
	ClaimGroupCustom,
}

// ── Category metadata ──────────────────────────────────────────────────────

// ClaimCategory holds the display metadata for a single claim.
// Group + Resource determines the matrix row; Action determines the column.
type ClaimCategory struct {
	Group    string
	Resource string
	Action   string
}

// ClaimCategoryMap maps every built-in claim ID to its display metadata.
// This is the single source of truth — update here to change the UI matrix.
var ClaimCategoryMap = map[string]ClaimCategory{

	// ── Administration › User ─────────────────────────────────────────────
	LIST_USER_CLAIM:   {ClaimGroupAdministration, "User", ClaimActionRead},
	CREATE_USER_CLAIM: {ClaimGroupAdministration, "User", ClaimActionCreate},
	UPDATE_USER_CLAIM: {ClaimGroupAdministration, "User", ClaimActionUpdate},
	DELETE_USER_CLAIM: {ClaimGroupAdministration, "User", ClaimActionDelete},

	// ── Administration › API Key ──────────────────────────────────────────
	LIST_API_KEY_CLAIM:   {ClaimGroupAdministration, "API Key", ClaimActionRead},
	CREATE_API_KEY_CLAIM: {ClaimGroupAdministration, "API Key", ClaimActionCreate},
	UPDATE_API_KEY_CLAIM: {ClaimGroupAdministration, "API Key", ClaimActionUpdate},
	DELETE_API_KEY_CLAIM: {ClaimGroupAdministration, "API Key", ClaimActionDelete},

	// ── Administration › API Key (Own) ────────────────────────────────────
	LIST_OWN_API_KEY_CLAIM:   {ClaimGroupAdministration, "API Key (Own)", ClaimActionRead},
	CREATE_OWN_API_KEY_CLAIM: {ClaimGroupAdministration, "API Key (Own)", ClaimActionCreate},
	UPDATE_OWN_API_KEY_CLAIM: {ClaimGroupAdministration, "API Key (Own)", ClaimActionUpdate},
	DELETE_OWN_API_KEY_CLAIM: {ClaimGroupAdministration, "API Key (Own)", ClaimActionDelete},

	// ── Administration › Role ─────────────────────────────────────────────
	LIST_ROLE_CLAIM:   {ClaimGroupAdministration, "Role", ClaimActionRead},
	CREATE_ROLE_CLAIM: {ClaimGroupAdministration, "Role", ClaimActionCreate},
	UPDATE_ROLE_CLAIM: {ClaimGroupAdministration, "Role", ClaimActionUpdate},
	DELETE_ROLE_CLAIM: {ClaimGroupAdministration, "Role", ClaimActionDelete},

	// ── Administration › Claim ────────────────────────────────────────────
	LIST_CLAIM_CLAIM:   {ClaimGroupAdministration, "Claim", ClaimActionRead},
	CREATE_CLAIM_CLAIM: {ClaimGroupAdministration, "Claim", ClaimActionCreate},
	UPDATE_CLAIM_CLAIM: {ClaimGroupAdministration, "Claim", ClaimActionUpdate},
	DELETE_CLAIM_CLAIM: {ClaimGroupAdministration, "Claim", ClaimActionDelete},

	// ── Administration › System (broad CRUD grants) ───────────────────────
	READ_ONLY_CLAIM: {ClaimGroupAdministration, "System", ClaimActionRead},
	LIST_CLAIM:      {ClaimGroupAdministration, "System", ClaimActionRead},
	CREATE_CLAIM:    {ClaimGroupAdministration, "System", ClaimActionCreate},
	UPDATE_CLAIM:    {ClaimGroupAdministration, "System", ClaimActionUpdate},
	DELETE_CLAIM:    {ClaimGroupAdministration, "System", ClaimActionDelete},

	// ── VMs › VM ──────────────────────────────────────────────────────────
	LIST_VM_CLAIM:            {ClaimGroupVMs, "VM", ClaimActionRead},
	CREATE_VM_CLAIM:          {ClaimGroupVMs, "VM", ClaimActionCreate},
	UPDATE_VM_CLAIM:          {ClaimGroupVMs, "VM", ClaimActionUpdate},
	UPDATE_VM_STATES_CLAIM:   {ClaimGroupVMs, "VM", ClaimActionUpdate},
	DELETE_VM_CLAIM:          {ClaimGroupVMs, "VM", ClaimActionDelete},
	EXECUTE_COMMAND_VM_CLAIM: {ClaimGroupVMs, "VM", ClaimActionExecute},

	// ── VMs › Snapshot ────────────────────────────────────────────────────
	LIST_SNAPSHOT_VM_CLAIM:        {ClaimGroupVMs, "Snapshot", ClaimActionRead},
	CREATE_SNAPSHOT_VM_CLAIM:      {ClaimGroupVMs, "Snapshot", ClaimActionCreate},
	DELETE_SNAPSHOT_VM_CLAIM:      {ClaimGroupVMs, "Snapshot", ClaimActionDelete},
	DELETE_ALL_SNAPSHOTS_VM_CLAIM: {ClaimGroupVMs, "Snapshot", ClaimActionDelete},
	REVERT_SNAPSHOT_VM_CLAIM:      {ClaimGroupVMs, "Snapshot", ClaimActionRevert},

	// ── VMs › Snapshot (Own) ──────────────────────────────────────────────
	LIST_OWN_VM_SNAPSHOT_CLAIM:        {ClaimGroupVMs, "Snapshot (Own)", ClaimActionRead},
	CREATE_OWN_VM_SNAPSHOT_CLAIM:      {ClaimGroupVMs, "Snapshot (Own)", ClaimActionCreate},
	DELETE_OWN_VM_SNAPSHOT_CLAIM:      {ClaimGroupVMs, "Snapshot (Own)", ClaimActionDelete},
	DELETE_ALL_OWN_VM_SNAPSHOTS_CLAIM: {ClaimGroupVMs, "Snapshot (Own)", ClaimActionDelete},
	REVERT_OWN_VM_SNAPSHOT_CLAIM:      {ClaimGroupVMs, "Snapshot (Own)", ClaimActionRevert},

	// ── VMs › Packer Template ─────────────────────────────────────────────
	LIST_PACKER_TEMPLATE_CLAIM:   {ClaimGroupVMs, "Packer Template", ClaimActionRead},
	CREATE_PACKER_TEMPLATE_CLAIM: {ClaimGroupVMs, "Packer Template", ClaimActionCreate},
	UPDATE_PACKER_TEMPLATE_CLAIM: {ClaimGroupVMs, "Packer Template", ClaimActionUpdate},
	DELETE_PACKER_TEMPLATE_CLAIM: {ClaimGroupVMs, "Packer Template", ClaimActionDelete},

	// ── Catalog › Manifest ────────────────────────────────────────────────
	LIST_CATALOG_MANIFEST_CLAIM:   {ClaimGroupCatalog, "Manifest", ClaimActionRead},
	CREATE_CATALOG_MANIFEST_CLAIM: {ClaimGroupCatalog, "Manifest", ClaimActionCreate},
	UPDATE_CATALOG_MANIFEST_CLAIM: {ClaimGroupCatalog, "Manifest", ClaimActionUpdate},
	DELETE_CATALOG_MANIFEST_CLAIM: {ClaimGroupCatalog, "Manifest", ClaimActionDelete},
	PULL_CATALOG_MANIFEST_CLAIM:   {ClaimGroupCatalog, "Manifest", ClaimActionPull},
	PUSH_CATALOG_MANIFEST_CLAIM:   {ClaimGroupCatalog, "Manifest", ClaimActionPush},
	IMPORT_CATALOG_MANIFEST_CLAIM: {ClaimGroupCatalog, "Manifest", ClaimActionImport},

	// ── Catalog Manager › Manager ─────────────────────────────────────────
	CATALOG_MANAGER_LIST_CLAIM:   {ClaimGroupCatalogManager, "Manager", ClaimActionRead},
	CATALOG_MANAGER_CREATE_CLAIM: {ClaimGroupCatalogManager, "Manager", ClaimActionCreate},
	CATALOG_MANAGER_UPDATE_CLAIM: {ClaimGroupCatalogManager, "Manager", ClaimActionUpdate},
	CATALOG_MANAGER_DELETE_CLAIM: {ClaimGroupCatalogManager, "Manager", ClaimActionDelete},

	// ── Catalog Manager › Manager (Own) ───────────────────────────────────
	CATALOG_MANAGER_LIST_OWN_CLAIM:   {ClaimGroupCatalogManager, "Manager (Own)", ClaimActionRead},
	CATALOG_MANAGER_CREATE_OWN_CLAIM: {ClaimGroupCatalogManager, "Manager (Own)", ClaimActionCreate},
	CATALOG_MANAGER_UPDATE_OWN_CLAIM: {ClaimGroupCatalogManager, "Manager (Own)", ClaimActionUpdate},
	CATALOG_MANAGER_DELETE_OWN_CLAIM: {ClaimGroupCatalogManager, "Manager (Own)", ClaimActionDelete},

	// ── Catalog Manager › Manifest (Own) ──────────────────────────────────
	CATALOG_MANAGER_LIST_CATALOG_MANIFEST_OWN_CLAIM:   {ClaimGroupCatalogManager, "Manifest (Own)", ClaimActionRead},
	CATALOG_MANAGER_CREATE_CATALOG_MANIFEST_OWN_CLAIM: {ClaimGroupCatalogManager, "Manifest (Own)", ClaimActionCreate},
	CATALOG_MANAGER_UPDATE_CATALOG_MANIFEST_OWN_CLAIM: {ClaimGroupCatalogManager, "Manifest (Own)", ClaimActionUpdate},
	CATALOG_MANAGER_DELETE_CATALOG_MANIFEST_OWN_CLAIM: {ClaimGroupCatalogManager, "Manifest (Own)", ClaimActionDelete},
	CATALOG_MANAGER_PULL_CATALOG_MANIFEST_OWN_CLAIM:   {ClaimGroupCatalogManager, "Manifest (Own)", ClaimActionPull},
	CATALOG_MANAGER_PUSH_CATALOG_MANIFEST_OWN_CLAIM:   {ClaimGroupCatalogManager, "Manifest (Own)", ClaimActionPush},
	CATALOG_MANAGER_IMPORT_CATALOG_MANIFEST_OWN_CLAIM: {ClaimGroupCatalogManager, "Manifest (Own)", ClaimActionImport},

	// ── Reverse Proxy › Config ────────────────────────────────────────────
	CONFIGURE_REVERSE_PROXY_CLAIM: {ClaimGroupReverseProxy, "Configuration", ClaimActionConfigure},

	// ── Reverse Proxy › Host ──────────────────────────────────────────────
	LIST_REVERSE_PROXY_HOSTS_CLAIM:  {ClaimGroupReverseProxy, "Host", ClaimActionRead},
	CREATE_REVERSE_PROXY_HOST_CLAIM: {ClaimGroupReverseProxy, "Host", ClaimActionCreate},
	UPDATE_REVERSE_PROXY_HOST_CLAIM: {ClaimGroupReverseProxy, "Host", ClaimActionUpdate},
	DELETE_REVERSE_PROXY_HOST_CLAIM: {ClaimGroupReverseProxy, "Host", ClaimActionDelete},

	// ── Reverse Proxy › HTTP Route ────────────────────────────────────────
	LIST_REVERSE_PROXY_HOST_HTTP_ROUTES_CLAIM:  {ClaimGroupReverseProxy, "HTTP Route", ClaimActionRead},
	CREATE_REVERSE_PROXY_HOST_HTTP_ROUTE_CLAIM: {ClaimGroupReverseProxy, "HTTP Route", ClaimActionCreate},
	UPDATE_REVERSE_PROXY_HOST_HTTP_ROUTE_CLAIM: {ClaimGroupReverseProxy, "HTTP Route", ClaimActionUpdate},
	DELETE_REVERSE_PROXY_HOST_HTTP_ROUTE_CLAIM: {ClaimGroupReverseProxy, "HTTP Route", ClaimActionDelete},

	// ── Reverse Proxy › TCP Route ─────────────────────────────────────────
	LIST_REVERSE_PROXY_HOST_TCP_ROUTES_CLAIM:  {ClaimGroupReverseProxy, "TCP Route", ClaimActionRead},
	CREATE_REVERSE_PROXY_HOST_TCP_ROUTE_CLAIM: {ClaimGroupReverseProxy, "TCP Route", ClaimActionCreate},
	UPDATE_REVERSE_PROXY_HOST_TCP_ROUTE_CLAIM: {ClaimGroupReverseProxy, "TCP Route", ClaimActionUpdate},
	DELETE_REVERSE_PROXY_HOST_TCP_ROUTE_CLAIM: {ClaimGroupReverseProxy, "TCP Route", ClaimActionDelete},

	// ── Cache ─────────────────────────────────────────────────────────────
	LIST_CACHE_CLAIM:        {ClaimGroupCache, "Cache", ClaimActionRead},
	DELETE_CACHE_ITEM_CLAIM: {ClaimGroupCache, "Cache", ClaimActionDelete},
	DELETE_ALL_CACHE_CLAIM:  {ClaimGroupCache, "Cache", ClaimActionDelete},

	// ── Jobs ──────────────────────────────────────────────────────────────
	JOBS_MANAGER_LIST_CLAIM:   {ClaimGroupJobs, "Job", ClaimActionRead},
	JOBS_MANAGER_DELETE_CLAIM: {ClaimGroupJobs, "Job", ClaimActionDelete},
	JOBS_MANAGER_DEBUG_CLAIM:  {ClaimGroupJobs, "Job", ClaimActionExecute},

	// ── Jobs › Own ────────────────────────────────────────────────────────
	JOBS_MANAGER_LIST_OWN_CLAIM: {ClaimGroupJobs, "Job (Own)", ClaimActionRead},

	// ── SSH ───────────────────────────────────────────────────────────────
	EXECUTE_SSH_CLAIM: {ClaimGroupSSH, "SSH", ClaimActionExecute},
}
