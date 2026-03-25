package constants

// RoleDescriptionMap maps every built-in role ID to a human-readable
// description shown in the UI role picker and role detail pages.
var RoleDescriptionMap = map[string]string{
	USER_ROLE:       "Standard user with read-only and basic self-service access to virtual machines and catalog resources.",
	ADMIN_ROLE:      "Full administrative access to manage users, roles, claims, virtual machines, catalog, and all platform resources.",
	SUPER_USER_ROLE: "Unrestricted system-level access including all administrative, operational, and platform configuration capabilities.",
}

// ClaimDescriptionMap maps every built-in claim ID to a human-readable
// description shown in the UI permissions matrix and claim picker.
var ClaimDescriptionMap = map[string]string{

	// ── Administration › System ────────────────────────────────────────────
	READ_ONLY_CLAIM: "Read-only access across all resources in the system.",
	LIST_CLAIM:      "Broad list access across all resource types.",
	CREATE_CLAIM:    "Broad create access across all resource types.",
	UPDATE_CLAIM:    "Broad update access across all resource types.",
	DELETE_CLAIM:    "Broad delete access across all resource types.",

	// ── Administration › User ─────────────────────────────────────────────
	LIST_USER_CLAIM:   "View all user accounts in the system.",
	CREATE_USER_CLAIM: "Create new user accounts.",
	UPDATE_USER_CLAIM: "Modify existing user account details and credentials.",
	DELETE_USER_CLAIM: "Remove user accounts from the system.",

	// ── Administration › API Key ──────────────────────────────────────────
	LIST_API_KEY_CLAIM:   "View all API keys.",
	CREATE_API_KEY_CLAIM: "Generate new API keys.",
	UPDATE_API_KEY_CLAIM: "Modify existing API key settings.",
	DELETE_API_KEY_CLAIM: "Revoke and remove API keys.",

	// ── Administration › Role ─────────────────────────────────────────────
	LIST_ROLE_CLAIM:   "View all roles.",
	CREATE_ROLE_CLAIM: "Create new roles.",
	UPDATE_ROLE_CLAIM: "Modify roles and their claim assignments.",
	DELETE_ROLE_CLAIM: "Remove roles from the system.",

	// ── Administration › Claim ────────────────────────────────────────────
	LIST_CLAIM_CLAIM:   "View all claims.",
	CREATE_CLAIM_CLAIM: "Create new custom claims.",
	UPDATE_CLAIM_CLAIM: "Modify existing claims.",
	DELETE_CLAIM_CLAIM: "Remove claims from the system.",

	// ── VMs › VM ──────────────────────────────────────────────────────────
	LIST_VM_CLAIM:            "View all virtual machines.",
	CREATE_VM_CLAIM:          "Create and provision new virtual machines.",
	UPDATE_VM_CLAIM:          "Modify virtual machine configuration and settings.",
	UPDATE_VM_STATES_CLAIM:   "Start, stop, pause, and resume virtual machines.",
	DELETE_VM_CLAIM:          "Delete virtual machines permanently.",
	EXECUTE_COMMAND_VM_CLAIM: "Execute commands inside virtual machines.",

	// ── VMs › Snapshot ────────────────────────────────────────────────────
	LIST_SNAPSHOT_VM_CLAIM:        "View all snapshots for any virtual machine.",
	CREATE_SNAPSHOT_VM_CLAIM:      "Create snapshots for any virtual machine.",
	DELETE_SNAPSHOT_VM_CLAIM:      "Delete snapshots from any virtual machine.",
	DELETE_ALL_SNAPSHOTS_VM_CLAIM: "Delete all snapshots from any virtual machine at once.",
	REVERT_SNAPSHOT_VM_CLAIM:      "Revert any virtual machine to a previous snapshot.",

	// ── VMs › Snapshot (Own) ──────────────────────────────────────────────
	LIST_OWN_VM_SNAPSHOT_CLAIM:        "View snapshots for virtual machines you own.",
	CREATE_OWN_VM_SNAPSHOT_CLAIM:      "Create snapshots for virtual machines you own.",
	DELETE_OWN_VM_SNAPSHOT_CLAIM:      "Delete snapshots from virtual machines you own.",
	DELETE_ALL_OWN_VM_SNAPSHOTS_CLAIM: "Delete all snapshots from your own virtual machines at once.",
	REVERT_OWN_VM_SNAPSHOT_CLAIM:      "Revert your own virtual machines to a previous snapshot.",

	// ── VMs › Packer Template ─────────────────────────────────────────────
	LIST_PACKER_TEMPLATE_CLAIM:   "View all Packer build templates.",
	CREATE_PACKER_TEMPLATE_CLAIM: "Create new Packer build templates.",
	UPDATE_PACKER_TEMPLATE_CLAIM: "Modify existing Packer build templates.",
	DELETE_PACKER_TEMPLATE_CLAIM: "Remove Packer build templates.",

	// ── Catalog › Manifest ────────────────────────────────────────────────
	LIST_CATALOG_MANIFEST_CLAIM:   "View all catalog manifests.",
	CREATE_CATALOG_MANIFEST_CLAIM: "Publish new virtual machine manifests to the catalog.",
	UPDATE_CATALOG_MANIFEST_CLAIM: "Modify existing catalog manifests.",
	DELETE_CATALOG_MANIFEST_CLAIM: "Remove manifests from the catalog.",
	PULL_CATALOG_MANIFEST_CLAIM:   "Download virtual machines from the catalog.",
	PUSH_CATALOG_MANIFEST_CLAIM:   "Upload virtual machines to the catalog.",
	IMPORT_CATALOG_MANIFEST_CLAIM: "Import catalog manifests from external sources.",

	// ── Catalog Manager › Manager ─────────────────────────────────────────
	CATALOG_MANAGER_LIST_CLAIM:   "View all registered catalog managers.",
	CATALOG_MANAGER_CREATE_CLAIM: "Register new catalog managers.",
	CATALOG_MANAGER_UPDATE_CLAIM: "Modify catalog manager configuration.",
	CATALOG_MANAGER_DELETE_CLAIM: "Remove catalog managers from the system.",

	// ── Catalog Manager › Manager (Own) ───────────────────────────────────
	CATALOG_MANAGER_LIST_OWN_CLAIM:   "View catalog managers you own.",
	CATALOG_MANAGER_CREATE_OWN_CLAIM: "Register catalog managers under your ownership.",
	CATALOG_MANAGER_UPDATE_OWN_CLAIM: "Modify the catalog managers you own.",
	CATALOG_MANAGER_DELETE_OWN_CLAIM: "Remove your own catalog managers.",

	// ── Catalog Manager › Manifest (Own) ──────────────────────────────────
	CATALOG_MANAGER_LIST_CATALOG_MANIFEST_OWN_CLAIM:   "View manifests on your own catalog managers.",
	CATALOG_MANAGER_CREATE_CATALOG_MANIFEST_OWN_CLAIM: "Publish manifests to your own catalog managers.",
	CATALOG_MANAGER_UPDATE_CATALOG_MANIFEST_OWN_CLAIM: "Modify manifests on your own catalog managers.",
	CATALOG_MANAGER_DELETE_CATALOG_MANIFEST_OWN_CLAIM: "Remove manifests from your own catalog managers.",
	CATALOG_MANAGER_PULL_CATALOG_MANIFEST_OWN_CLAIM:   "Download virtual machines from your own catalog managers.",
	CATALOG_MANAGER_PUSH_CATALOG_MANIFEST_OWN_CLAIM:   "Upload virtual machines to your own catalog managers.",
	CATALOG_MANAGER_IMPORT_CATALOG_MANIFEST_OWN_CLAIM: "Import manifests into your own catalog managers.",

	// ── Reverse Proxy ─────────────────────────────────────────────────────
	CONFIGURE_REVERSE_PROXY_CLAIM: "Manage global reverse proxy configuration and settings.",

	// ── Reverse Proxy › Host ──────────────────────────────────────────────
	LIST_REVERSE_PROXY_HOSTS_CLAIM:  "View all reverse proxy hosts.",
	CREATE_REVERSE_PROXY_HOST_CLAIM: "Add new hosts to the reverse proxy.",
	UPDATE_REVERSE_PROXY_HOST_CLAIM: "Modify reverse proxy host settings.",
	DELETE_REVERSE_PROXY_HOST_CLAIM: "Remove hosts from the reverse proxy.",

	// ── Reverse Proxy › HTTP Route ────────────────────────────────────────
	LIST_REVERSE_PROXY_HOST_HTTP_ROUTES_CLAIM:  "View HTTP routes on reverse proxy hosts.",
	CREATE_REVERSE_PROXY_HOST_HTTP_ROUTE_CLAIM: "Add HTTP routes to reverse proxy hosts.",
	UPDATE_REVERSE_PROXY_HOST_HTTP_ROUTE_CLAIM: "Modify HTTP routes on reverse proxy hosts.",
	DELETE_REVERSE_PROXY_HOST_HTTP_ROUTE_CLAIM: "Remove HTTP routes from reverse proxy hosts.",

	// ── Reverse Proxy › TCP Route ─────────────────────────────────────────
	LIST_REVERSE_PROXY_HOST_TCP_ROUTES_CLAIM:  "View TCP routes on reverse proxy hosts.",
	CREATE_REVERSE_PROXY_HOST_TCP_ROUTE_CLAIM: "Add TCP routes to reverse proxy hosts.",
	UPDATE_REVERSE_PROXY_HOST_TCP_ROUTE_CLAIM: "Modify TCP routes on reverse proxy hosts.",
	DELETE_REVERSE_PROXY_HOST_TCP_ROUTE_CLAIM: "Remove TCP routes from reverse proxy hosts.",

	// ── Cache ─────────────────────────────────────────────────────────────
	LIST_CACHE_CLAIM:        "View items currently stored in the catalog cache.",
	DELETE_CACHE_ITEM_CLAIM: "Remove a specific item from the catalog cache.",
	DELETE_ALL_CACHE_CLAIM:  "Clear all items from the catalog cache.",

	// ── Jobs ──────────────────────────────────────────────────────────────
	JOBS_MANAGER_LIST_CLAIM:   "View all background jobs across all users.",
	JOBS_MANAGER_DELETE_CLAIM: "Cancel and delete background jobs.",
	JOBS_MANAGER_DEBUG_CLAIM:  "Access detailed job debug information and internal logs.",

	// ── Jobs › Own ────────────────────────────────────────────────────────
	JOBS_MANAGER_LIST_OWN_CLAIM: "View your own background jobs.",

	// ── SSH ───────────────────────────────────────────────────────────────
	EXECUTE_SSH_CLAIM: "Open SSH connections to hosts and virtual machines.",
}
