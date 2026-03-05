---
layout: api
title: Roles
default_host: http://localhost
api_prefix: /api
categories:
    - name: Config
      path: config
      endpoints:
        - anchor: _health_probe_get
          method: get
          path: /health/probe
          description: This endpoint returns the API Health Probe
          title: Gets the API Health Probe
        - anchor: _health_system_get
          method: get
          path: /health/system
          description: This endpoint returns the API Health Probe
          title: Gets the API System Health
        - anchor: _logs_get
          method: get
          path: /logs
          description: This endpoint returns the system logs from the disk
          title: Gets the system logs from the disk
        - anchor: _logs_stream_get
          method: get
          path: /logs/stream
          description: This endpoint streams the system logs in real-time via WebSocket
          title: Streams the system logs via WebSocket
        - anchor: _v1_config_hardware_get
          method: get
          path: /v1/config/hardware
          description: This endpoint returns the Hardware Info
          title: Gets the Hardware Info
        - anchor: _v1_config_tools_install_post
          method: post
          path: /v1/config/tools/install
          description: This endpoint installs API requires 3rd party tools
          title: Installs API requires 3rd party tools
        - anchor: _v1_config_tools_restart_post
          method: post
          path: /v1/config/tools/restart
          description: This endpoint restarts the API Service
          title: Restarts the API Service
        - anchor: _v1_config_tools_uninstall_post
          method: post
          path: /v1/config/tools/uninstall
          description: This endpoint uninstalls API requires 3rd party tools
          title: Uninstalls API requires 3rd party tools
        - anchor: _v1_orchestrator_hosts_id_logs_get
          method: get
          path: /v1/orchestrator/hosts/{id}/logs
          description: This endpoint returns the orchestrator host system logs from the disk
          title: Gets the orchestrator host system logs from the disk
        - anchor: _v1_parallels_desktop_key_get
          method: get
          path: /v1/parallels_desktop/key
          description: This endpoint returns Parallels Desktop active license
          title: Gets Parallels Desktop active license
    - name: Api Keys
      path: api_keys
      endpoints:
        - anchor: _v1_auth_api_keys_get
          method: get
          path: /v1/auth/api_keys
          description: This endpoint returns all the api keys
          title: Gets all the api keys
        - anchor: _v1_auth_api_keys_post
          method: post
          path: /v1/auth/api_keys
          description: This endpoint creates an api key
          title: Creates an api key
        - anchor: _v1_auth_api_keys_id_get
          method: get
          path: /v1/auth/api_keys/{id}
          description: This endpoint returns an api key by id or name
          title: Gets an api key by id or name
        - anchor: _v1_auth_api_keys_id_delete
          method: delete
          path: /v1/auth/api_keys/{id}
          description: This endpoint deletes an api key
          title: Deletes an api key
        - anchor: _v1_auth_api_keys_id_revoke_put
          method: put
          path: /v1/auth/api_keys/{id}/revoke
          description: This endpoint revokes an api key
          title: Revoke an api key
    - name: Claims
      path: claims
      endpoints:
        - anchor: _v1_auth_claims_get
          method: get
          path: /v1/auth/claims
          description: This endpoint returns all the claims
          title: Gets all the claims
        - anchor: _v1_auth_claims_post
          method: post
          path: /v1/auth/claims
          description: This endpoint creates a claim
          title: Creates a claim
        - anchor: _v1_auth_claims_id_get
          method: get
          path: /v1/auth/claims/{id}
          description: This endpoint returns a claim
          title: Gets a claim
        - anchor: _v1_auth_claims_id_delete
          method: delete
          path: /v1/auth/claims/{id}
          description: This endpoint Deletes a claim
          title: Delete a claim
    - name: Roles
      path: roles
      endpoints:
        - anchor: _v1_auth_roles_get
          method: get
          path: /v1/auth/roles
          description: This endpoint returns all the roles
          title: Gets all the roles
        - anchor: _v1_auth_roles_post
          method: post
          path: /v1/auth/roles
          description: This endpoint returns a role
          title: Gets a role
        - anchor: _v1_auth_roles_id_get
          method: get
          path: /v1/auth/roles/{id}
          description: This endpoint returns a role
          title: Gets a role
        - anchor: _v1_auth_roles_id_delete
          method: delete
          path: /v1/auth/roles/{id}
          description: This endpoint deletes a role
          title: Delete a role
    - name: Authorization
      path: authorization
      endpoints:
        - anchor: _v1_auth_token_post
          method: post
          path: /v1/auth/token
          description: This endpoint generates a token
          title: Generates a token
        - anchor: _v1_auth_token_validate_post
          method: post
          path: /v1/auth/token/validate
          description: This endpoint validates a token
          title: Validates a token
    - name: Users
      path: users
      endpoints:
        - anchor: _v1_auth_users_get
          method: get
          path: /v1/auth/users
          description: This endpoint returns all the users
          title: Gets all the users
        - anchor: _v1_auth_users_post
          method: post
          path: /v1/auth/users
          description: This endpoint creates a user
          title: Creates a user
        - anchor: _v1_auth_users_id_get
          method: get
          path: /v1/auth/users/{id}
          description: This endpoint returns a user
          title: Gets a user
        - anchor: _v1_auth_users_id_put
          method: put
          path: /v1/auth/users/{id}
          description: This endpoint updates a user
          title: Update a user
        - anchor: _v1_auth_users_id_delete
          method: delete
          path: /v1/auth/users/{id}
          description: This endpoint deletes a user
          title: Deletes a user
        - anchor: _v1_auth_users_id_claims_get
          method: get
          path: /v1/auth/users/{id}/claims
          description: This endpoint returns all the claims for a user
          title: Gets all the claims for a user
        - anchor: _v1_auth_users_id_claims_post
          method: post
          path: /v1/auth/users/{id}/claims
          description: This endpoint adds a claim to a user
          title: Adds a claim to a user
        - anchor: _v1_auth_users_id_claims_claim_id_delete
          method: delete
          path: /v1/auth/users/{id}/claims/{claim_id}
          description: This endpoint removes a claim from a user
          title: Removes a claim from a user
        - anchor: _v1_auth_users_id_roles_get
          method: get
          path: /v1/auth/users/{id}/roles
          description: This endpoint returns all the roles for a user
          title: Gets all the roles for a user
        - anchor: _v1_auth_users_id_roles_post
          method: post
          path: /v1/auth/users/{id}/roles
          description: This endpoint adds a role to a user
          title: Adds a role to a user
        - anchor: _v1_auth_users_id_roles_role_id_delete
          method: delete
          path: /v1/auth/users/{id}/roles/{role_id}
          description: This endpoint removes a role from a user
          title: Removes a role from a user
    - name: Catalogs
      path: catalogs
      endpoints:
        - anchor: _v1_cache_get
          method: get
          path: /v1/cache
          description: This endpoint returns all the remote catalog cache if any
          title: Gets catalog cache
        - anchor: _v1_cache_delete
          method: delete
          path: /v1/cache
          description: This endpoint returns all the remote catalog cache if any
          title: Deletes all catalog cache
        - anchor: _v1_cache_catalogId_delete
          method: delete
          path: /v1/cache/{catalogId}
          description: This endpoint returns all the remote catalog cache if any and all its versions
          title: Deletes catalog cache item and all its versions
        - anchor: _v1_cache_catalogId_version_delete
          method: delete
          path: /v1/cache/{catalogId}/{version}
          description: This endpoint deletes a version of a cache ite,
          title: Deletes catalog cache version item
        - anchor: _v1_catalog_get
          method: get
          path: /v1/catalog
          description: This endpoint returns all the remote catalogs
          title: Gets all the remote catalogs
        - anchor: _v1_catalog_import_put
          method: put
          path: /v1/catalog/import
          description: This endpoint imports a remote catalog manifest metadata into the catalog inventory
          title: Imports a remote catalog manifest metadata into the catalog inventory
        - anchor: _v1_catalog_import-vm_put
          method: put
          path: /v1/catalog/import-vm
          description: This endpoint imports a virtual machine in pvm or macvm format into the catalog inventory generating the metadata for it
          title: Imports a vm into the catalog inventory generating the metadata for it
        - anchor: _v1_catalog_pull_put
          method: put
          path: /v1/catalog/pull
          description: This endpoint pulls a remote catalog manifest
          title: Pull a remote catalog manifest
        - anchor: _v1_catalog_pull_async_put
          method: put
          path: /v1/catalog/pull/async
          description: This endpoint pulls a remote catalog manifest in the background and returns a Job ID to track progress
          title: Pull a remote catalog manifest asynchronously
        - anchor: _v1_catalog_push_post
          method: post
          path: /v1/catalog/push
          description: This endpoint pushes a catalog manifest to the catalog inventory
          title: Pushes a catalog manifest to the catalog inventory
        - anchor: _v1_catalog_catalogId_get
          method: get
          path: /v1/catalog/{catalogId}
          description: This endpoint returns all the remote catalogs
          title: Gets all the remote catalogs
        - anchor: _v1_catalog_catalogId_delete
          method: delete
          path: /v1/catalog/{catalogId}
          description: This endpoint deletes a catalog manifest and all its versions
          title: Deletes a catalog manifest and all its versions
        - anchor: _v1_catalog_catalogId_version_get
          method: get
          path: /v1/catalog/{catalogId}/{version}
          description: This endpoint returns a catalog manifest version
          title: Gets a catalog manifest version
        - anchor: _v1_catalog_catalogId_version_delete
          method: delete
          path: /v1/catalog/{catalogId}/{version}
          description: This endpoint deletes a catalog manifest version
          title: Deletes a catalog manifest version
        - anchor: _v1_catalog_catalogId_version_architecture_get
          method: get
          path: /v1/catalog/{catalogId}/{version}/{architecture}
          description: This endpoint returns a catalog manifest version
          title: Gets a catalog manifest version architecture
        - anchor: _v1_catalog_catalogId_version_architecture_delete
          method: delete
          path: /v1/catalog/{catalogId}/{version}/{architecture}
          description: This endpoint deletes a catalog manifest version
          title: Deletes a catalog manifest version architecture
        - anchor: _v1_catalog_catalogId_version_architecture_claims_delete
          method: delete
          path: /v1/catalog/{catalogId}/{version}/{architecture}/claims
          description: This endpoint removes claims from a catalog manifest version
          title: Removes claims from a catalog manifest version
        - anchor: _v1_catalog_catalogId_version_architecture_claims_patch
          method: patch
          path: /v1/catalog/{catalogId}/{version}/{architecture}/claims
          description: This endpoint adds claims to a catalog manifest version
          title: Updates a catalog
        - anchor: _v1_catalog_catalogId_version_architecture_download_get
          method: get
          path: /v1/catalog/{catalogId}/{version}/{architecture}/download
          description: This endpoint downloads a catalog manifest version
          title: Downloads a catalog manifest version
        - anchor: _v1_catalog_catalogId_version_architecture_revoke_patch
          method: patch
          path: /v1/catalog/{catalogId}/{version}/{architecture}/revoke
          description: This endpoint UnTaints a catalog manifest version
          title: UnTaints a catalog manifest version
        - anchor: _v1_catalog_catalogId_version_architecture_roles_delete
          method: delete
          path: /v1/catalog/{catalogId}/{version}/{architecture}/roles
          description: This endpoint removes roles from a catalog manifest version
          title: Removes roles from a catalog manifest version
        - anchor: _v1_catalog_catalogId_version_architecture_roles_patch
          method: patch
          path: /v1/catalog/{catalogId}/{version}/{architecture}/roles
          description: This endpoint adds roles to a catalog manifest version
          title: Adds roles to a catalog manifest version
        - anchor: _v1_catalog_catalogId_version_architecture_tags_delete
          method: delete
          path: /v1/catalog/{catalogId}/{version}/{architecture}/tags
          description: This endpoint removes tags from a catalog manifest version
          title: Removes tags from a catalog manifest version
        - anchor: _v1_catalog_catalogId_version_architecture_tags_patch
          method: patch
          path: /v1/catalog/{catalogId}/{version}/{architecture}/tags
          description: This endpoint adds tags to a catalog manifest version
          title: Adds tags to a catalog manifest version
        - anchor: _v1_catalog_catalogId_version_architecture_taint_patch
          method: patch
          path: /v1/catalog/{catalogId}/{version}/{architecture}/taint
          description: This endpoint Taints a catalog manifest version
          title: Taints a catalog manifest version
        - anchor: _v1_catalog_catalogId_version_architecture_untaint_patch
          method: patch
          path: /v1/catalog/{catalogId}/{version}/{architecture}/untaint
          description: This endpoint UnTaints a catalog manifest version
          title: UnTaints a catalog manifest version
    - name: CatalogManagers
      path: catalogmanagers
      endpoints:
        - anchor: _v1_catalog-managers_get
          method: get
          path: /v1/catalog-managers
          description: This endpoint returns all the catalog managers
          title: Gets all the catalog managers
        - anchor: _v1_catalog-managers_post
          method: post
          path: /v1/catalog-managers
          description: This endpoint creates a catalog manager
          title: Creates a catalog manager
        - anchor: _v1_catalog-managers_id_get
          method: get
          path: /v1/catalog-managers/{id}
          description: This endpoint returns a catalog manager
          title: Gets a specific catalog manager
        - anchor: _v1_catalog-managers_id_put
          method: put
          path: /v1/catalog-managers/{id}
          description: This endpoint updates a catalog manager
          title: Updates a catalog manager
        - anchor: _v1_catalog-managers_id_delete
          method: delete
          path: /v1/catalog-managers/{id}
          description: This endpoint deletes a catalog manager
          title: Deletes a catalog manager
    - name: Machines
      path: machines
      endpoints:
        - anchor: _v1_machines_get
          method: get
          path: /v1/machines
          description: This endpoint returns all the virtual machines
          title: Gets all the virtual machines
        - anchor: _v1_machines_post
          method: post
          path: /v1/machines
          description: This endpoint creates a virtual machine
          title: Creates a virtual machine
        - anchor: _v1_machines_async_post
          method: post
          path: /v1/machines/async
          description: This endpoint creates a virtual machine in the background and returns a Job ID to track progress
          title: Creates a virtual machine asynchronously
        - anchor: _v1_machines_register_post
          method: post
          path: /v1/machines/register
          description: This endpoint registers a virtual machine
          title: Registers a virtual machine
        - anchor: _v1_machines_id_get
          method: get
          path: /v1/machines/{id}
          description: This endpoint returns a virtual machine
          title: Gets a virtual machine
        - anchor: _v1_machines_id_delete
          method: delete
          path: /v1/machines/{id}
          description: This endpoint deletes a virtual machine
          title: Deletes a virtual machine
        - anchor: _v1_machines_id_clone_put
          method: put
          path: /v1/machines/{id}/clone
          description: This endpoint clones a virtual machine
          title: Clones a virtual machine
        - anchor: _v1_machines_id_execute_put
          method: put
          path: /v1/machines/{id}/execute
          description: This endpoint executes a command on a virtual machine
          title: Executes a command on a virtual machine
        - anchor: _v1_machines_id_pause_get
          method: get
          path: /v1/machines/{id}/pause
          description: This endpoint pauses a virtual machine
          title: Pauses a virtual machine
        - anchor: _v1_machines_id_rename_put
          method: put
          path: /v1/machines/{id}/rename
          description: This endpoint Renames a virtual machine
          title: Renames a virtual machine
        - anchor: _v1_machines_id_reset_get
          method: get
          path: /v1/machines/{id}/reset
          description: This endpoint reset a virtual machine
          title: Reset a virtual machine
        - anchor: _v1_machines_id_restart_get
          method: get
          path: /v1/machines/{id}/restart
          description: This endpoint restarts a virtual machine
          title: Restarts a virtual machine
        - anchor: _v1_machines_id_resume_get
          method: get
          path: /v1/machines/{id}/resume
          description: This endpoint resumes a virtual machine
          title: Resumes a virtual machine
        - anchor: _v1_machines_id_set_put
          method: put
          path: /v1/machines/{id}/set
          description: This endpoint configures a virtual machine
          title: Configures a virtual machine
        - anchor: _v1_machines_id_snapshots_get
          method: get
          path: /v1/machines/{id}/snapshots
          description: This endpoint lists snapshots of a virtual machine
          title: Lists snapshots of a virtual machine
        - anchor: _v1_machines_id_snapshots_post
          method: post
          path: /v1/machines/{id}/snapshots
          description: This endpoint creates a snapshot for a virtual machine
          title: Creates a snapshot for a virtual machine
        - anchor: _v1_machines_id_snapshots_delete
          method: delete
          path: /v1/machines/{id}/snapshots
          description: This endpoint deletes all snapshots of a virtual machine
          title: Deletes all snapshots of a virtual machine
        - anchor: _v1_machines_id_snapshots_snapshot_id_delete
          method: delete
          path: /v1/machines/{id}/snapshots/{snapshot_id}
          description: This endpoint deletes a snapshot of a virtual machine
          title: Deletes a snapshot of a virtual machine
        - anchor: _v1_machines_id_snapshots_snapshot_id_revert_post
          method: post
          path: /v1/machines/{id}/snapshots/{snapshot_id}/revert
          description: This endpoint reverts a virtual machine to a snapshot
          title: Reverts a virtual machine to a snapshot
        - anchor: _v1_machines_id_start_get
          method: get
          path: /v1/machines/{id}/start
          description: This endpoint starts a virtual machine
          title: Starts a virtual machine
        - anchor: _v1_machines_id_status_get
          method: get
          path: /v1/machines/{id}/status
          description: This endpoint returns the current state of a virtual machine
          title: Get the current state of a virtual machine
        - anchor: _v1_machines_id_stop_get
          method: get
          path: /v1/machines/{id}/stop
          description: This endpoint stops a virtual machine
          title: Stops a virtual machine
        - anchor: _v1_machines_id_suspend_get
          method: get
          path: /v1/machines/{id}/suspend
          description: This endpoint suspends a virtual machine
          title: Suspends a virtual machine
        - anchor: _v1_machines_id_unregister_post
          method: post
          path: /v1/machines/{id}/unregister
          description: This endpoint unregister a virtual machine
          title: Unregister a virtual machine
        - anchor: _v1_machines_id_upload_post
          method: post
          path: /v1/machines/{id}/upload
          description: This endpoint executes a command on a virtual machine
          title: Uploads a file to a virtual machine
    - name: Orchestrator
      path: orchestrator
      endpoints:
        - anchor: _v1_orchestrator_hosts_get
          method: get
          path: /v1/orchestrator/hosts
          description: This endpoint returns all hosts from the orchestrator
          title: Gets all hosts from the orchestrator
        - anchor: _v1_orchestrator_hosts_put
          method: put
          path: /v1/orchestrator/hosts
          description: This endpoint updates a host in the orchestrator
          title: Update a Host in the orchestrator
        - anchor: _v1_orchestrator_hosts_post
          method: post
          path: /v1/orchestrator/hosts
          description: This endpoint register a host in the orchestrator
          title: Register a Host in the orchestrator
        - anchor: _v1_orchestrator_hosts_id_get
          method: get
          path: /v1/orchestrator/hosts/{id}
          description: This endpoint returns a host from the orchestrator
          title: Gets a host from the orchestrator
        - anchor: _v1_orchestrator_hosts_id_delete
          method: delete
          path: /v1/orchestrator/hosts/{id}
          description: This endpoint unregister a host from the orchestrator
          title: Unregister a host from the orchestrator
        - anchor: _v1_orchestrator_hosts_id_catalog_cache_get
          method: get
          path: /v1/orchestrator/hosts/{id}/catalog/cache
          description: This endpoint returns orchestrator host catalog cache
          title: Gets orchestrator host catalog cache
        - anchor: _v1_orchestrator_hosts_id_catalog_cache_delete
          method: delete
          path: /v1/orchestrator/hosts/{id}/catalog/cache
          description: This endpoint deletes an orchestrator host cache items
          title: Deletes an orchestrator host cache items
        - anchor: _v1_orchestrator_hosts_id_catalog_cache_catalog_id_delete
          method: delete
          path: /v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}
          description: This endpoint deletes an orchestrator host cache item and all its children
          title: Deletes an orchestrator host cache item and all its children
        - anchor: _v1_orchestrator_hosts_id_catalog_cache_catalog_id_catalog_version_delete
          method: delete
          path: /v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}/{catalog_version}
          description: This endpoint deletes an orchestrator host cache item and all its children
          title: Deletes an orchestrator host cache item and all its children
        - anchor: _v1_orchestrator_hosts_id_disable_put
          method: put
          path: /v1/orchestrator/hosts/{id}/disable
          description: This endpoint will disable an existing host in the orchestrator
          title: Disable a host in the orchestrator
        - anchor: _v1_orchestrator_hosts_id_enable_put
          method: put
          path: /v1/orchestrator/hosts/{id}/enable
          description: This endpoint will enable an existing host in the orchestrator
          title: Enable a host in the orchestrator
        - anchor: _v1_orchestrator_hosts_id_hardware_get
          method: get
          path: /v1/orchestrator/hosts/{id}/hardware
          description: This endpoint returns a host hardware info from the orchestrator
          title: Gets a host hardware info from the orchestrator
        - anchor: _v1_orchestrator_hosts_id_machines_get
          method: get
          path: /v1/orchestrator/hosts/{id}/machines
          description: This endpoint returns orchestrator host virtual machines
          title: Get orchestrator host virtual machines
        - anchor: _v1_orchestrator_hosts_id_machines_post
          method: post
          path: /v1/orchestrator/hosts/{id}/machines
          description: This endpoint creates a orchestrator host virtual machine
          title: Creates a orchestrator host virtual machine
        - anchor: _v1_orchestrator_hosts_id_machines_register_post
          method: post
          path: /v1/orchestrator/hosts/{id}/machines/register
          description: This endpoint registers a virtual machine in a orchestrator host
          title: Register a virtual machine in a orchestrator host
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_get
          method: get
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}
          description: This endpoint returns orchestrator host virtual machine
          title: Get orchestrator host virtual machine
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_delete
          method: delete
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}
          description: This endpoint deletes orchestrator host virtual machine
          title: Deletes orchestrator host virtual machine
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_clone_put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/clone
          description: This endpoint clones orchestrator host virtual machine
          title: Clones orchestrator host virtual machine
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_execute_put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/execute
          description: This endpoint executes a command in a orchestrator host virtual machine
          title: Executes a command in a orchestrator host virtual machine
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_pause_put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/pause
          description: This endpoint pauses orchestrator host virtual machine
          title: Pauses orchestrator host virtual machine
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_rename_put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/rename
          description: This endpoint renames orchestrator host virtual machine
          title: Renames orchestrator host virtual machine
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_reset_put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/reset
          description: This endpoint resets orchestrator host virtual machine
          title: Resets orchestrator host virtual machine
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_restart_put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/restart
          description: This endpoint restarts orchestrator host virtual machine
          title: Restarts orchestrator host virtual machine
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_resume_put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/resume
          description: This endpoint resumes orchestrator host virtual machine
          title: Resumes orchestrator host virtual machine
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_set_put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/set
          description: This endpoint configures orchestrator host virtual machine
          title: Configures orchestrator host virtual machine
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_start_put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/start
          description: This endpoint starts orchestrator host virtual machine
          title: Starts orchestrator host virtual machine
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_status_get
          method: get
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/status
          description: This endpoint returns orchestrator host virtual machine status
          title: Get orchestrator host virtual machine status
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_stop_put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/stop
          description: This endpoint stops orchestrator host virtual machine
          title: Stops orchestrator host virtual machine
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_suspend_put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/suspend
          description: This endpoint suspends orchestrator host virtual machine
          title: Suspends orchestrator host virtual machine
        - anchor: _v1_orchestrator_hosts_id_machines_vmId_unregister_post
          method: post
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/unregister
          description: This endpoint unregister a virtual machine in a orchestrator host
          title: Unregister a virtual machine in a orchestrator host
        - anchor: _v1_orchestrator_hosts_id_reverse-proxy_get
          method: get
          path: /v1/orchestrator/hosts/{id}/reverse-proxy
          description: This endpoint returns orchestrator host reverse proxy configuration
          title: Gets orchestrator host reverse proxy configuration
        - anchor: _v1_orchestrator_hosts_id_reverse-proxy_disable_put
          method: put
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/disable
          description: This endpoint disables orchestrator host reverse proxy
          title: Disables orchestrator host reverse proxy
        - anchor: _v1_orchestrator_hosts_id_reverse-proxy_enable_put
          method: put
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/enable
          description: This endpoint enables orchestrator host reverse proxy
          title: Enables orchestrator host reverse proxy
        - anchor: _v1_orchestrator_hosts_id_reverse-proxy_hosts_get
          method: get
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts
          description: This endpoint returns orchestrator host reverse proxy hosts
          title: Gets orchestrator host reverse proxy hosts
        - anchor: _v1_orchestrator_hosts_id_reverse-proxy_hosts_post
          method: post
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts
          description: This endpoint creates a orchestrator host reverse proxy host
          title: Creates a orchestrator host reverse proxy host
        - anchor: _v1_orchestrator_hosts_id_reverse-proxy_hosts_reverse_proxy_host_id_get
          method: get
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}
          description: This endpoint returns orchestrator host reverse proxy hosts
          title: Gets orchestrator host reverse proxy hosts
        - anchor: _v1_orchestrator_hosts_id_reverse-proxy_hosts_reverse_proxy_host_id_put
          method: put
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}
          description: This endpoint updates an orchestrator host reverse proxy host
          title: Updates an orchestrator host reverse proxy host
        - anchor: _v1_orchestrator_hosts_id_reverse-proxy_hosts_reverse_proxy_host_id_delete
          method: delete
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}
          description: This endpoint deletes an orchestrator host reverse proxy host
          title: Deletes an orchestrator host reverse proxy host
        - anchor: _v1_orchestrator_hosts_id_reverse-proxy_hosts_reverse_proxy_host_id_http_routes_post
          method: post
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes
          description: This endpoint upserts an orchestrator host reverse proxy host http route
          title: Upserts an orchestrator host reverse proxy host http route
        - anchor: _v1_orchestrator_hosts_id_reverse-proxy_hosts_reverse_proxy_host_id_http_routes_route_id_delete
          method: delete
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes/{route_id}
          description: This endpoint deletes an orchestrator host reverse proxy host http route
          title: Deletes an orchestrator host reverse proxy host http route
        - anchor: _v1_orchestrator_hosts_id_reverse-proxy_hosts_reverse_proxy_host_id_tcp_route_post
          method: post
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/tcp_route
          description: This endpoint updates an orchestrator host reverse proxy host tcp route
          title: Update an orchestrator host reverse proxy host tcp route
        - anchor: _v1_orchestrator_hosts_id_reverse-proxy_restart_put
          method: put
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/restart
          description: This endpoint restarts orchestrator host reverse proxy
          title: Restarts orchestrator host reverse proxy
        - anchor: _v1_orchestrator_machines_get
          method: get
          path: /v1/orchestrator/machines
          description: This endpoint returns orchestrator Virtual Machines
          title: Get orchestrator Virtual Machines
        - anchor: _v1_orchestrator_machines_post
          method: post
          path: /v1/orchestrator/machines
          description: This endpoint creates a virtual machine in one of the hosts for the orchestrator
          title: Creates a virtual machine in one of the hosts for the orchestrator
        - anchor: _v1_orchestrator_machines_id_get
          method: get
          path: /v1/orchestrator/machines/{id}
          description: This endpoint returns orchestrator Virtual Machine by its ID
          title: Get orchestrator Virtual Machine
        - anchor: _v1_orchestrator_machines_id_delete
          method: delete
          path: /v1/orchestrator/machines/{id}
          description: This endpoint deletes orchestrator virtual machine
          title: Deletes orchestrator virtual machine
        - anchor: _v1_orchestrator_machines_id_clone_put
          method: put
          path: /v1/orchestrator/machines/{id}/clone
          description: This endpoint clones orchestrator virtual machine
          title: Clones orchestrator virtual machine
        - anchor: _v1_orchestrator_machines_id_execute_put
          method: put
          path: /v1/orchestrator/machines/{id}/execute
          description: This endpoint executes a command in a orchestrator virtual machine
          title: Executes a command in a orchestrator virtual machine
        - anchor: _v1_orchestrator_machines_id_pause_put
          method: put
          path: /v1/orchestrator/machines/{id}/pause
          description: This endpoint pauses orchestrator virtual machine
          title: Pauses orchestrator virtual machine
        - anchor: _v1_orchestrator_machines_id_rename_put
          method: put
          path: /v1/orchestrator/machines/{id}/rename
          description: This endpoint renames orchestrator virtual machine
          title: Renames orchestrator virtual machine
        - anchor: _v1_orchestrator_machines_id_reset_put
          method: put
          path: /v1/orchestrator/machines/{id}/reset
          description: This endpoint resets orchestrator virtual machine
          title: Resets orchestrator virtual machine
        - anchor: _v1_orchestrator_machines_id_restart_put
          method: put
          path: /v1/orchestrator/machines/{id}/restart
          description: This endpoint restarts orchestrator virtual machine
          title: Restarts orchestrator virtual machine
        - anchor: _v1_orchestrator_machines_id_resume_put
          method: put
          path: /v1/orchestrator/machines/{id}/resume
          description: This endpoint resumes orchestrator virtual machine
          title: Resumes orchestrator virtual machine
        - anchor: _v1_orchestrator_machines_id_set_put
          method: put
          path: /v1/orchestrator/machines/{id}/set
          description: This endpoint configures orchestrator virtual machine
          title: Configures orchestrator virtual machine
        - anchor: _v1_orchestrator_machines_id_start_put
          method: put
          path: /v1/orchestrator/machines/{id}/start
          description: This endpoint starts orchestrator virtual machine
          title: Starts orchestrator virtual machine
        - anchor: _v1_orchestrator_machines_id_status_get
          method: get
          path: /v1/orchestrator/machines/{id}/status
          description: This endpoint returns orchestrator virtual machine status
          title: Get orchestrator virtual machine status
        - anchor: _v1_orchestrator_machines_id_stop_put
          method: put
          path: /v1/orchestrator/machines/{id}/stop
          description: This endpoint sops orchestrator virtual machine
          title: Stops orchestrator virtual machine
        - anchor: _v1_orchestrator_machines_id_suspend_put
          method: put
          path: /v1/orchestrator/machines/{id}/suspend
          description: This endpoint suspends orchestrator virtual machine
          title: Suspends orchestrator virtual machine
        - anchor: _v1_orchestrator_overview_resources_get
          method: get
          path: /v1/orchestrator/overview/resources
          description: This endpoint returns orchestrator resource overview
          title: Get orchestrator resource overview
        - anchor: _v1_orchestrator_overview_id_resources_get
          method: get
          path: /v1/orchestrator/overview/{id}/resources
          description: This endpoint returns orchestrator host resources
          title: Get orchestrator host resources
    - name: ReverseProxy
      path: reverseproxy
      endpoints:
        - anchor: _v1_reverse-proxy_get
          method: get
          path: /v1/reverse-proxy
          description: This endpoint returns the reverse proxy configuration
          title: Gets reverse proxy configuration
        - anchor: _v1_reverse-proxy_disable_put
          method: put
          path: /v1/reverse-proxy/disable
          description: This endpoint will disable the reverse proxy
          title: Disable the reverse proxy
        - anchor: _v1_reverse-proxy_enable_put
          method: put
          path: /v1/reverse-proxy/enable
          description: This endpoint will enable the reverse proxy
          title: Enable the reverse proxy
        - anchor: _v1_reverse-proxy_hosts_get
          method: get
          path: /v1/reverse-proxy/hosts
          description: This endpoint returns all the reverse proxy hosts
          title: Gets all the reverse proxy hosts
        - anchor: _v1_reverse-proxy_hosts_post
          method: post
          path: /v1/reverse-proxy/hosts
          description: This endpoint creates a reverse proxy host
          title: Creates a reverse proxy host
        - anchor: _v1_reverse-proxy_hosts_id_get
          method: get
          path: /v1/reverse-proxy/hosts/{id}
          description: This endpoint returns a reverse proxy host
          title: Gets all the reverse proxy host
        - anchor: _v1_reverse-proxy_hosts_id_put
          method: put
          path: /v1/reverse-proxy/hosts/{id}
          description: This endpoint creates a reverse proxy host
          title: Updates a reverse proxy host
        - anchor: _v1_reverse-proxy_hosts_id_delete
          method: delete
          path: /v1/reverse-proxy/hosts/{id}
          description: This endpoint Deletes a reverse proxy host
          title: Delete a a reverse proxy host
        - anchor: _v1_reverse-proxy_hosts_id_http_route_post
          method: post
          path: /v1/reverse-proxy/hosts/{id}/http_route
          description: This endpoint creates or updates a reverse proxy host HTTP route
          title: Creates or updates a reverse proxy host HTTP route
        - anchor: _v1_reverse-proxy_hosts_id_http_routes_post
          method: post
          path: /v1/reverse-proxy/hosts/{id}/http_routes
          description: This endpoint updates a reverse proxy host TCP route
          title: Updates a reverse proxy host TCP route
        - anchor: _v1_reverse-proxy_hosts_id_http_routes_order_put
          method: put
          path: /v1/reverse-proxy/hosts/{id}/http_routes/order
          description: This endpoint reorders HTTP routes for a reverse proxy host
          title: Updates the order of a reverse proxy host HTTP route
        - anchor: _v1_reverse-proxy_hosts_id_http_routes_http_route_id_delete
          method: delete
          path: /v1/reverse-proxy/hosts/{id}/http_routes/{http_route_id}
          description: This endpoint Deletes a reverse proxy host HTTP route
          title: Delete a a reverse proxy host HTTP route
        - anchor: _v1_reverse-proxy_restart_put
          method: put
          path: /v1/reverse-proxy/restart
          description: This endpoint will restart the reverse proxy
          title: Restarts the reverse proxy
    - name: SSH
      path: ssh
      endpoints:
        - anchor: _v1_ssh_execute_post
          method: post
          path: /v1/ssh/execute
          description: Executes a command on a remote host via SSH
          title: Execute SSH Command
    - name: Packer Templates
      path: packer_templates
      endpoints:
        - anchor: _v1_templates_packer_get
          method: get
          path: /v1/templates/packer
          description: This endpoint returns all the packer templates. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
          title: Gets all the packer templates
        - anchor: _v1_templates_packer_post
          method: post
          path: /v1/templates/packer
          description: This endpoint creates a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
          title: Creates a packer template
        - anchor: _v1_templates_packer_id_get
          method: get
          path: /v1/templates/packer/{id}
          description: This endpoint returns a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
          title: Gets a packer template
        - anchor: _v1_templates_packer_id_put
          method: put
          path: /v1/templates/packer/{id}
          description: This endpoint updates a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
          title: Updates a packer template
        - anchor: _v1_templates_packer_id_delete
          method: delete
          path: /v1/templates/packer/{id}
          description: This endpoint deletes a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
          title: Deletes a packer template
    - name: Events
      path: events
      endpoints:
        - anchor: _v1_ws_subscribe_get
          method: get
          path: /v1/ws/subscribe
          description: This endpoint upgrades the HTTP connection to WebSocket and subscribes to event notifications. Authentication is required via Authorization header (Bearer token) or query parameters (access_token or authorization).
          title: Subscribe to event notifications via WebSocket
        - anchor: _v1_ws_unsubscribe_post
          method: post
          path: /v1/ws/unsubscribe
          description: Unsubscribe an active WebSocket client from specific event types without disconnecting. The client must belong to the authenticated user.
          title: Unsubscribe from specific event types
endpoints:
  - path: /v1/auth/roles
    method: get
    title: Gets all the roles
    description: This endpoint returns all the roles
    requires_authorization: true
    example_blocks:
      - title: cURL
        language: powershell
        code_block: |
          curl --location '{{host}}/v1/auth/roles' \
          --header 'Authorization: ******'
      - title: C#
        language: csharp
        code_block: |
          var client = new HttpClient();
          var request = new HttpRequestMessage(HttpMethod.Get, "{{host}}/v1/auth/roles");
          request.Headers.Add("Authorization", "******");
          
          var response = await client.SendAsync(request);
      - title: Go
        language: go
        code_block: |
          package main
          
          import (
              "fmt"
              "strings"
              "net/http"
              "io/ioutil"
          )
          
          func main() {
              url := "{{host}}/v1/auth/roles"
              method := "GET"
              
              payload := strings.NewReader("")
              
              client := &http.Client {}
              req, err := http.NewRequest(method, url, payload)
              
              if err != nil {
                  fmt.Println(err)
                  return
              }
              req.Header.Add("Authorization", "******")
              
              
              res, err := client.Do(req)
              if err != nil {
                  fmt.Println(err)
                  return
              }
              defer res.Body.Close()
              
              body, err := ioutil.ReadAll(res.Body)
              if err != nil {
                  fmt.Println(err)
                  return
              }
              fmt.Println(string(body))
          }
    response_blocks:
      - code: 200
        code_description: OK
        code_block: |
          [
            {
              "id": "string",
              "name": "string"
            }
          ]
      - code: 400
        code_description: Bad Request
        code_block: |
          {
            "message": "string",
            "timestamp": "2024-01-01T00:00:00Z"
          }
      - code: 401
        code_description: Unauthorized
        code_block: |
          {
            "code": "int",
            "message": "string",
            "stack": [
              {
                "function": "string",
                "file": "string",
                "line": "int"
              }
            ]
          }
  - path: /v1/auth/roles
    method: post
    title: Gets a role
    description: This endpoint returns a role
    requires_authorization: true
    example_blocks:
      - title: cURL
        language: powershell
        code_block: |
          curl --location '{{host}}/v1/auth/roles' \
          --header 'Authorization: ******' \
          --header 'Content-Type: application/json' \
          --data '{\n  \"key\": \"SomeKey\",\n  \"secret\": \"SomeLongSecret\"\n}'
      - title: C#
        language: csharp
        code_block: |
          var client = new HttpClient();
          var request = new HttpRequestMessage(HttpMethod.Post, "{{host}}/v1/auth/roles");
          request.Headers.Add("Authorization", "******");
          request.Content = new StringContent("{\n  \"key\": \"SomeKey\",\n  \"secret\": \"SomeLongSecret\"\n}", Encoding.UTF8, "application/json");
          var response = await client.SendAsync(request);
      - title: Go
        language: go
        code_block: |
          package main
          
          import (
              "fmt"
              "strings"
              "net/http"
              "io/ioutil"
          )
          
          func main() {
              url := "{{host}}/v1/auth/roles"
              method := "POST"
              
              payload := strings.NewReader("{\n  \"key\": \"SomeKey\",\n  \"secret\": \"SomeLongSecret\"\n}")
              
              client := &http.Client {}
              req, err := http.NewRequest(method, url, payload)
              
              if err != nil {
                  fmt.Println(err)
                  return
              }
              req.Header.Add("Authorization", "******")
              req.Header.Add("Content-Type", "application/json")
              
              res, err := client.Do(req)
              if err != nil {
                  fmt.Println(err)
                  return
              }
              defer res.Body.Close()
              
              body, err := ioutil.ReadAll(res.Body)
              if err != nil {
                  fmt.Println(err)
                  return
              }
              fmt.Println(string(body))
          }
    response_blocks:
      - code: 200
        code_description: OK
        code_block: |
          {
            "message": "string",
            "timestamp": "2024-01-01T00:00:00Z"
          }
      - code: 400
        code_description: Bad Request
        code_block: |
          {
            "message": "string",
            "timestamp": "2024-01-01T00:00:00Z"
          }
      - code: 401
        code_description: Unauthorized
        code_block: |
          {
            "code": "int",
            "message": "string",
            "stack": [
              {
                "function": "string",
                "file": "string",
                "line": "int"
              }
            ]
          }
  - path: /v1/auth/roles/{id}
    method: get
    title: Gets a role
    description: This endpoint returns a role
    requires_authorization: true
    example_blocks:
      - title: cURL
        language: powershell
        code_block: |
          curl --location '{{host}}/v1/auth/roles/{id}' \
          --header 'Authorization: ******'
      - title: C#
        language: csharp
        code_block: |
          var client = new HttpClient();
          var request = new HttpRequestMessage(HttpMethod.Get, "{{host}}/v1/auth/roles/{id}");
          request.Headers.Add("Authorization", "******");
          
          var response = await client.SendAsync(request);
      - title: Go
        language: go
        code_block: |
          package main
          
          import (
              "fmt"
              "strings"
              "net/http"
              "io/ioutil"
          )
          
          func main() {
              url := "{{host}}/v1/auth/roles/{id}"
              method := "GET"
              
              payload := strings.NewReader("")
              
              client := &http.Client {}
              req, err := http.NewRequest(method, url, payload)
              
              if err != nil {
                  fmt.Println(err)
                  return
              }
              req.Header.Add("Authorization", "******")
              
              
              res, err := client.Do(req)
              if err != nil {
                  fmt.Println(err)
                  return
              }
              defer res.Body.Close()
              
              body, err := ioutil.ReadAll(res.Body)
              if err != nil {
                  fmt.Println(err)
                  return
              }
              fmt.Println(string(body))
          }
    response_blocks:
      - code: 200
        code_description: OK
        code_block: |
          {
            "message": "string",
            "timestamp": "2024-01-01T00:00:00Z"
          }
      - code: 400
        code_description: Bad Request
        code_block: |
          {
            "message": "string",
            "timestamp": "2024-01-01T00:00:00Z"
          }
      - code: 401
        code_description: Unauthorized
        code_block: |
          {
            "code": "int",
            "message": "string",
            "stack": [
              {
                "function": "string",
                "file": "string",
                "line": "int"
              }
            ]
          }
  - path: /v1/auth/roles/{id}
    method: delete
    title: Delete a role
    description: This endpoint deletes a role
    requires_authorization: true
    example_blocks:
      - title: cURL
        language: powershell
        code_block: |
          curl --location '{{host}}/v1/auth/roles/{id}' \
          --header 'Authorization: ******'
      - title: C#
        language: csharp
        code_block: |
          var client = new HttpClient();
          var request = new HttpRequestMessage(HttpMethod.Delete, "{{host}}/v1/auth/roles/{id}");
          request.Headers.Add("Authorization", "******");
          
          var response = await client.SendAsync(request);
      - title: Go
        language: go
        code_block: |
          package main
          
          import (
              "fmt"
              "strings"
              "net/http"
              "io/ioutil"
          )
          
          func main() {
              url := "{{host}}/v1/auth/roles/{id}"
              method := "DELETE"
              
              payload := strings.NewReader("")
              
              client := &http.Client {}
              req, err := http.NewRequest(method, url, payload)
              
              if err != nil {
                  fmt.Println(err)
                  return
              }
              req.Header.Add("Authorization", "******")
              
              
              res, err := client.Do(req)
              if err != nil {
                  fmt.Println(err)
                  return
              }
              defer res.Body.Close()
              
              body, err := ioutil.ReadAll(res.Body)
              if err != nil {
                  fmt.Println(err)
                  return
              }
              fmt.Println(string(body))
          }
    response_blocks:
      - code: 202
        code_description: Accepted
        code_block: |
          {
            "message": "string",
            "timestamp": "2024-01-01T00:00:00Z"
          }
      - code: 400
        code_description: Bad Request
        code_block: |
          {
            "message": "string",
            "timestamp": "2024-01-01T00:00:00Z"
          }
      - code: 401
        code_description: Unauthorized
        code_block: |
          {
            "code": "int",
            "message": "string",
            "stack": [
              {
                "function": "string",
                "file": "string",
                "line": "int"
              }
            ]
          }
---

# Roles API

This section contains all roles related endpoints.
