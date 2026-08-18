---
layout: api
title: Roles
default_host: http://localhost
api_prefix: /api
is_category_document: true
categories:
    - name: Api Keys
      path: api_keys
      endpoints:
        - anchor: /v1/auth/api_keys-post
          method: post
          path: /v1/auth/api_keys
          description: This endpoint creates an api key
          title: Creates an api key
        - anchor: /v1/auth/api_keys-get
          method: get
          path: /v1/auth/api_keys
          description: This endpoint returns all the api keys
          title: Gets all the api keys
        - anchor: /v1/auth/api_keys/{id}-delete
          method: delete
          path: /v1/auth/api_keys/{id}
          description: This endpoint deletes an api key
          title: Deletes an api key
        - anchor: /v1/auth/api_keys/{id}-get
          method: get
          path: /v1/auth/api_keys/{id}
          description: This endpoint returns an api key by id or name
          title: Gets an api key by id or name
        - anchor: /v1/auth/api_keys/{id}/revoke-put
          method: put
          path: /v1/auth/api_keys/{id}/revoke
          description: This endpoint revokes an api key
          title: Revoke an api key
    - name: Authorization
      path: authorization
      endpoints:
        - anchor: /v1/auth/token-post
          method: post
          path: /v1/auth/token
          description: This endpoint generates a token
          title: Generates a token
        - anchor: /v1/auth/token/validate-post
          method: post
          path: /v1/auth/token/validate
          description: This endpoint validates a token
          title: Validates a token
    - name: Catalogs
      path: catalogs
      endpoints:
        - anchor: /v1/cache-get
          method: get
          path: /v1/cache
          description: This endpoint returns all the remote catalog cache if any
          title: Gets catalog cache
        - anchor: /v1/cache-delete
          method: delete
          path: /v1/cache
          description: This endpoint returns all the remote catalog cache if any
          title: Deletes all catalog cache
        - anchor: /v1/cache/{catalogId}-delete
          method: delete
          path: /v1/cache/{catalogId}
          description: This endpoint returns all the remote catalog cache if any and all its versions
          title: Deletes catalog cache item and all its versions
        - anchor: /v1/cache/{catalogId}/{version}-delete
          method: delete
          path: /v1/cache/{catalogId}/{version}
          description: This endpoint deletes a version of a cache ite,
          title: Deletes catalog cache version item
        - anchor: /v1/catalog-get
          method: get
          path: /v1/catalog
          description: This endpoint returns all the remote catalogs
          title: Gets all the remote catalogs
        - anchor: /v1/catalog/{catalogId}-get
          method: get
          path: /v1/catalog/{catalogId}
          description: This endpoint returns all the remote catalogs
          title: Gets all the remote catalogs
        - anchor: /v1/catalog/{catalogId}/{version}-get
          method: get
          path: /v1/catalog/{catalogId}/{version}
          description: This endpoint returns a catalog manifest version
          title: Gets a catalog manifest version
        - anchor: /v1/catalog/{catalogId}/{version}/{architecture}-get
          method: get
          path: /v1/catalog/{catalogId}/{version}/{architecture}
          description: This endpoint returns a catalog manifest version
          title: Gets a catalog manifest version architecture
        - anchor: /v1/catalog/{catalogId}/{version}/{architecture}/download-get
          method: get
          path: /v1/catalog/{catalogId}/{version}/{architecture}/download
          description: This endpoint downloads a catalog manifest version
          title: Downloads a catalog manifest version
        - anchor: /v1/catalog/{catalogId}/{version}/{architecture}/taint-patch
          method: patch
          path: /v1/catalog/{catalogId}/{version}/{architecture}/taint
          description: This endpoint Taints a catalog manifest version
          title: Taints a catalog manifest version
        - anchor: /v1/catalog/{catalogId}/{version}/{architecture}/untaint-patch
          method: patch
          path: /v1/catalog/{catalogId}/{version}/{architecture}/untaint
          description: This endpoint UnTaints a catalog manifest version
          title: UnTaints a catalog manifest version
        - anchor: /v1/catalog/{catalogId}/{version}/{architecture}/revoke-patch
          method: patch
          path: /v1/catalog/{catalogId}/{version}/{architecture}/revoke
          description: This endpoint UnTaints a catalog manifest version
          title: UnTaints a catalog manifest version
        - anchor: /v1/catalog/{catalogId}/{version}/{architecture}/claims-patch
          method: patch
          path: /v1/catalog/{catalogId}/{version}/{architecture}/claims
          description: This endpoint adds claims to a catalog manifest version
          title: Adds claims to a catalog manifest version
        - anchor: /v1/catalog/{catalogId}/{version}/{architecture}/claims-delete
          method: delete
          path: /v1/catalog/{catalogId}/{version}/{architecture}/claims
          description: This endpoint removes claims from a catalog manifest version
          title: Removes claims from a catalog manifest version
        - anchor: /v1/catalog/{catalogId}/{version}/{architecture}/roles-patch
          method: patch
          path: /v1/catalog/{catalogId}/{version}/{architecture}/roles
          description: This endpoint adds roles to a catalog manifest version
          title: Adds roles to a catalog manifest version
        - anchor: /v1/catalog/{catalogId}/{version}/{architecture}/roles-delete
          method: delete
          path: /v1/catalog/{catalogId}/{version}/{architecture}/roles
          description: This endpoint removes roles from a catalog manifest version
          title: Removes roles from a catalog manifest version
        - anchor: /v1/catalog/{catalogId}/{version}/{architecture}/tags-patch
          method: patch
          path: /v1/catalog/{catalogId}/{version}/{architecture}/tags
          description: This endpoint adds tags to a catalog manifest version
          title: Adds tags to a catalog manifest version
        - anchor: /v1/catalog/{catalogId}/{version}/{architecture}/tags-delete
          method: delete
          path: /v1/catalog/{catalogId}/{version}/{architecture}/tags
          description: This endpoint removes tags from a catalog manifest version
          title: Removes tags from a catalog manifest version
        - anchor: /v1/catalog/{catalogId}-delete
          method: delete
          path: /v1/catalog/{catalogId}
          description: This endpoint deletes a catalog manifest and all its versions
          title: Deletes a catalog manifest and all its versions
        - anchor: /v1/catalog/{catalogId}/{version}-delete
          method: delete
          path: /v1/catalog/{catalogId}/{version}
          description: This endpoint deletes a catalog manifest version
          title: Deletes a catalog manifest version
        - anchor: /v1/catalog/{catalogId}/{version}/{architecture}-delete
          method: delete
          path: /v1/catalog/{catalogId}/{version}/{architecture}
          description: This endpoint deletes a catalog manifest version
          title: Deletes a catalog manifest version architecture
        - anchor: /v1/catalog/push-post
          method: post
          path: /v1/catalog/push
          description: This endpoint pushes a catalog manifest to the catalog inventory
          title: Pushes a catalog manifest to the catalog inventory
        - anchor: /v1/catalog/push/async-post
          method: post
          path: /v1/catalog/push/async
          description: This endpoint pushes a catalog manifest to the catalog inventory in the background and returns a Job ID to track progress
          title: Push a catalog manifest to the catalog inventory asynchronously
        - anchor: /v1/catalog/pull-put
          method: put
          path: /v1/catalog/pull
          description: This endpoint pulls a remote catalog manifest
          title: Pull a remote catalog manifest
        - anchor: /v1/catalog/pull/async-put
          method: put
          path: /v1/catalog/pull/async
          description: This endpoint pulls a remote catalog manifest in the background and returns a Job ID to track progress
          title: Pull a remote catalog manifest asynchronously
        - anchor: /v1/catalog/import-put
          method: put
          path: /v1/catalog/import
          description: This endpoint imports a remote catalog manifest metadata into the catalog inventory
          title: Imports a remote catalog manifest metadata into the catalog inventory
        - anchor: /v1/catalog/import-vm-put
          method: put
          path: /v1/catalog/import-vm
          description: This endpoint imports a virtual machine in pvm or macvm format into the catalog inventory generating the metadata for it
          title: Imports a vm into the catalog inventory generating the metadata for it
        - anchor: /v1/catalog/{catalogId}/{version}/{architecture}/claims-patch
          method: patch
          path: /v1/catalog/{catalogId}/{version}/{architecture}/claims
          description: This endpoint adds claims to a catalog manifest version
          title: Updates a catalog
        - anchor: /v1/catalog/{catalogId}/{version}/{architecture}/metadata-put
          method: put
          path: /v1/catalog/{catalogId}/{version}/{architecture}/metadata
          description: This endpoint atomically updates description, tags, required claims, and required roles for a catalog manifest version. Omit a field to leave it unchanged.
          title: Updates metadata for a catalog manifest version
    - name: CatalogManagers
      path: catalogmanagers
      endpoints:
        - anchor: /v1/catalog-managers-get
          method: get
          path: /v1/catalog-managers
          description: This endpoint returns all the catalog managers
          title: Gets all the catalog managers
        - anchor: /v1/catalog-managers/{id}-get
          method: get
          path: /v1/catalog-managers/{id}
          description: This endpoint returns a catalog manager
          title: Gets a specific catalog manager
        - anchor: /v1/catalog-managers-post
          method: post
          path: /v1/catalog-managers
          description: This endpoint creates a catalog manager
          title: Creates a catalog manager
        - anchor: /v1/catalog-managers/{id}-put
          method: put
          path: /v1/catalog-managers/{id}
          description: This endpoint updates a catalog manager
          title: Updates a catalog manager
        - anchor: /v1/catalog-managers/{id}-delete
          method: delete
          path: /v1/catalog-managers/{id}
          description: This endpoint deletes a catalog manager
          title: Deletes a catalog manager
    - name: Claims
      path: claims
      endpoints:
        - anchor: /v1/auth/claims-get
          method: get
          path: /v1/auth/claims
          description: This endpoint returns all the claims
          title: Gets all the claims
        - anchor: /v1/auth/claims/grouped-get
          method: get
          path: /v1/auth/claims/grouped
          description: This endpoint returns all claims organised by group and resource
          title: Gets all claims grouped for the matrix UI
        - anchor: /v1/auth/claims/{id}-get
          method: get
          path: /v1/auth/claims/{id}
          description: This endpoint returns a claim
          title: Gets a claim
        - anchor: /v1/auth/claims-post
          method: post
          path: /v1/auth/claims
          description: This endpoint creates a claim
          title: Creates a claim
        - anchor: /v1/auth/claims/{id}-delete
          method: delete
          path: /v1/auth/claims/{id}
          description: This endpoint Deletes a claim
          title: Delete a claim
    - name: Config
      path: config
      endpoints:
        - anchor: /v1/parallels_desktop/key-get
          method: get
          path: /v1/parallels_desktop/key
          description: This endpoint returns Parallels Desktop active license
          title: Gets Parallels Desktop active license
        - anchor: /v1/config/tools/install-post
          method: post
          path: /v1/config/tools/install
          description: This endpoint installs API requires 3rd party tools
          title: Installs API requires 3rd party tools
        - anchor: /v1/config/tools/uninstall-post
          method: post
          path: /v1/config/tools/uninstall
          description: This endpoint uninstalls API requires 3rd party tools
          title: Uninstalls API requires 3rd party tools
        - anchor: /v1/config/tools/restart-post
          method: post
          path: /v1/config/tools/restart
          description: This endpoint restarts the API Service
          title: Restarts the API Service
        - anchor: /v1/config/hardware-get
          method: get
          path: /v1/config/hardware
          description: This endpoint returns the Hardware Info
          title: Gets the Hardware Info
        - anchor: /health/system-get
          method: get
          path: /health/system
          description: This endpoint returns the API Health Probe
          title: Gets the API System Health
        - anchor: /logs-get
          method: get
          path: /logs
          description: This endpoint returns the system logs from the disk
          title: Gets the system logs from the disk
        - anchor: /logs/stream-get
          method: get
          path: /logs/stream
          description: This endpoint streams the system logs in real-time via WebSocket
          title: Streams the system logs via WebSocket
        - anchor: /config/diskspace-post
          method: post
          path: /config/diskspace
          description: This endpoint returns the available disk space for the cache folder.
          title: Gets the Parallels disk space information
        - anchor: /v1/orchestrator/hosts/{id}/logs-get
          method: get
          path: /v1/orchestrator/hosts/{id}/logs
          description: This endpoint returns the orchestrator host system logs from the disk
          title: Gets the orchestrator host system logs from the disk
        - anchor: /logs/stream-get
          method: get
          path: /logs/stream
          description: This endpoint streams the system logs in real-time via WebSocket
          title: Streams the system logs via WebSocket
        - anchor: /health/probe-get
          method: get
          path: /health/probe
          description: This endpoint returns the API Health Probe
          title: Gets the API Health Probe
    - name: Events
      path: events
      endpoints:
        - anchor: /v1/ws/subscribe-get
          method: get
          path: /v1/ws/subscribe
          description: This endpoint upgrades the HTTP connection to WebSocket and subscribes to event notifications. Authentication is required via Authorization header (Bearer token) or query parameters (access_token or authorization).
          title: Subscribe to event notifications via WebSocket
        - anchor: /v1/ws/clients-get
          method: get
          path: /v1/ws/clients
          description: Returns all currently connected WebSocket clients with queue depth and ping/pong timestamps. Useful for diagnosing stale or dead clients whose queues are filling up.
          title: List connected WebSocket clients
        - anchor: /v1/ws/stats-get
          method: get
          path: /v1/ws/stats
          description: Returns aggregate statistics including total connected clients, subscription counts per event type, uptime, and per-client details with queue depths.
          title: Get WebSocket event emitter statistics
        - anchor: /v1/ws/unsubscribe-post
          method: post
          path: /v1/ws/unsubscribe
          description: Unsubscribe an active WebSocket client from specific event types without disconnecting. The client must belong to the authenticated user.
          title: Unsubscribe from specific event types
    - name: Jobs
      path: jobs
      endpoints:
        - anchor: /v1/jobs/{id}-delete
          method: delete
          path: /v1/jobs/{id}
          description: This endpoint deletes a single job. Users with JOB_MANAGER_DELETE can delete any job; users with JOB_MANAGER_LIST_OWN can only delete their own.
          title: Deletes a job by ID
    - name: Machines
      path: machines
      endpoints:
        - anchor: /v1/machines-get
          method: get
          path: /v1/machines
          description: This endpoint returns all the virtual machines
          title: Gets all the virtual machines
        - anchor: /v1/machines/{id}-get
          method: get
          path: /v1/machines/{id}
          description: This endpoint returns a virtual machine
          title: Gets a virtual machine
        - anchor: /v1/machines/{id}/start-put
          method: put
          path: /v1/machines/{id}/start
          description: This endpoint starts a virtual machine
          title: Starts a virtual machine
        - anchor: /v1/machines/{id}/stop-put
          method: put
          path: /v1/machines/{id}/stop
          description: This endpoint stops a virtual machine
          title: Stops a virtual machine
        - anchor: /v1/machines/{id}/restart-put
          method: put
          path: /v1/machines/{id}/restart
          description: This endpoint restarts a virtual machine
          title: Restarts a virtual machine
        - anchor: /v1/machines/{id}/suspend-put
          method: put
          path: /v1/machines/{id}/suspend
          description: This endpoint suspends a virtual machine
          title: Suspends a virtual machine
        - anchor: /v1/machines/{id}/resume-put
          method: put
          path: /v1/machines/{id}/resume
          description: This endpoint resumes a virtual machine
          title: Resumes a virtual machine
        - anchor: /v1/machines/{id}/reset-put
          method: put
          path: /v1/machines/{id}/reset
          description: This endpoint reset a virtual machine
          title: Reset a virtual machine
        - anchor: /v1/machines/{id}/pause-put
          method: put
          path: /v1/machines/{id}/pause
          description: This endpoint pauses a virtual machine
          title: Pauses a virtual machine
        - anchor: /v1/machines/{id}-delete
          method: delete
          path: /v1/machines/{id}
          description: This endpoint deletes a virtual machine
          title: Deletes a virtual machine
        - anchor: /v1/machines/{id}/status-get
          method: get
          path: /v1/machines/{id}/status
          description: This endpoint returns the current state of a virtual machine
          title: Get the current state of a virtual machine
        - anchor: /v1/machines/{id}/set-put
          method: put
          path: /v1/machines/{id}/set
          description: This endpoint configures a virtual machine
          title: Configures a virtual machine
        - anchor: /v1/machines/{id}/clone-put
          method: put
          path: /v1/machines/{id}/clone
          description: This endpoint clones a virtual machine
          title: Clones a virtual machine
        - anchor: /v1/machines/{id}/execute-put
          method: put
          path: /v1/machines/{id}/execute
          description: This endpoint executes a command on a virtual machine
          title: Executes a command on a virtual machine
        - anchor: /v1/machines/{id}/upload-post
          method: post
          path: /v1/machines/{id}/upload
          description: This endpoint executes a command on a virtual machine
          title: Uploads a file to a virtual machine
        - anchor: /v1/machines/{id}/rename-put
          method: put
          path: /v1/machines/{id}/rename
          description: This endpoint Renames a virtual machine
          title: Renames a virtual machine
        - anchor: /v1/machines/register-post
          method: post
          path: /v1/machines/register
          description: This endpoint registers a virtual machine
          title: Registers a virtual machine
        - anchor: /v1/machines/{id}/unregister-post
          method: post
          path: /v1/machines/{id}/unregister
          description: This endpoint unregister a virtual machine
          title: Unregister a virtual machine
        - anchor: /v1/machines-post
          method: post
          path: /v1/machines
          description: This endpoint creates a virtual machine
          title: Creates a virtual machine
        - anchor: /v1/machines/async-post
          method: post
          path: /v1/machines/async
          description: This endpoint creates a virtual machine in the background and returns a Job ID to track progress
          title: Creates a virtual machine asynchronously
        - anchor: /v1/machines/{id}/snapshots-post
          method: post
          path: /v1/machines/{id}/snapshots
          description: This endpoint creates a snapshot for a virtual machine
          title: Creates a snapshot for a virtual machine
        - anchor: /v1/machines/{id}/snapshots/{snapshot_id}-delete
          method: delete
          path: /v1/machines/{id}/snapshots/{snapshot_id}
          description: This endpoint deletes a snapshot of a virtual machine
          title: Deletes a snapshot of a virtual machine
        - anchor: /v1/machines/{id}/snapshots-delete
          method: delete
          path: /v1/machines/{id}/snapshots
          description: This endpoint deletes all snapshots of a virtual machine
          title: Deletes all snapshots of a virtual machine
        - anchor: /v1/machines/{id}/snapshots-get
          method: get
          path: /v1/machines/{id}/snapshots
          description: This endpoint lists snapshots of a virtual machine
          title: Lists snapshots of a virtual machine
        - anchor: /v1/machines/{id}/snapshots/{snapshot_id}/revert-post
          method: post
          path: /v1/machines/{id}/snapshots/{snapshot_id}/revert
          description: This endpoint reverts a virtual machine to a snapshot
          title: Reverts a virtual machine to a snapshot
    - name: Orchestrator
      path: orchestrator
      endpoints:
        - anchor: /v1/orchestrator/hosts-get
          method: get
          path: /v1/orchestrator/hosts
          description: This endpoint returns all hosts from the orchestrator
          title: Gets all hosts from the orchestrator
        - anchor: /v1/orchestrator/hosts/{id}-get
          method: get
          path: /v1/orchestrator/hosts/{id}
          description: This endpoint returns a host from the orchestrator
          title: Gets a host from the orchestrator
        - anchor: /v1/orchestrator/hosts/{id}/hardware-get
          method: get
          path: /v1/orchestrator/hosts/{id}/hardware
          description: This endpoint returns a host hardware info from the orchestrator
          title: Gets a host hardware info from the orchestrator
        - anchor: /v1/orchestrator/hosts-post
          method: post
          path: /v1/orchestrator/hosts
          description: This endpoint register a host in the orchestrator
          title: Register a Host in the orchestrator
        - anchor: /v1/orchestrator/hosts/{id}-delete
          method: delete
          path: /v1/orchestrator/hosts/{id}
          description: This endpoint unregister a host from the orchestrator
          title: Unregister a host from the orchestrator
        - anchor: /v1/orchestrator/hosts/{id}/enable-put
          method: put
          path: /v1/orchestrator/hosts/{id}/enable
          description: This endpoint will enable an existing host in the orchestrator
          title: Enable a host in the orchestrator
        - anchor: /v1/orchestrator/hosts/{id}/disable-put
          method: put
          path: /v1/orchestrator/hosts/{id}/disable
          description: This endpoint will disable an existing host in the orchestrator
          title: Disable a host in the orchestrator
        - anchor: /v1/orchestrator/hosts-put
          method: put
          path: /v1/orchestrator/hosts
          description: This endpoint updates a host in the orchestrator
          title: Update a Host in the orchestrator
        - anchor: /v1/orchestrator/overview/resources-get
          method: get
          path: /v1/orchestrator/overview/resources
          description: This endpoint returns orchestrator resource overview
          title: Get orchestrator resource overview
        - anchor: /v1/orchestrator/overview/{id}/resources-get
          method: get
          path: /v1/orchestrator/overview/{id}/resources
          description: This endpoint returns orchestrator host resources
          title: Get orchestrator host resources
        - anchor: /v1/orchestrator/machines-get
          method: get
          path: /v1/orchestrator/machines
          description: This endpoint returns orchestrator Virtual Machines
          title: Get orchestrator Virtual Machines
        - anchor: /v1/orchestrator/machines/{id}-get
          method: get
          path: /v1/orchestrator/machines/{id}
          description: This endpoint returns orchestrator Virtual Machine by its ID
          title: Get orchestrator Virtual Machine
        - anchor: /v1/orchestrator/machines/{id}-delete
          method: delete
          path: /v1/orchestrator/machines/{id}
          description: This endpoint deletes orchestrator virtual machine
          title: Deletes orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/status-get
          method: get
          path: /v1/orchestrator/machines/{id}/status
          description: This endpoint returns orchestrator virtual machine status
          title: Get orchestrator virtual machine status
        - anchor: /v1/orchestrator/machines/{id}/rename-put
          method: put
          path: /v1/orchestrator/machines/{id}/rename
          description: This endpoint renames orchestrator virtual machine
          title: Renames orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/set-put
          method: put
          path: /v1/orchestrator/machines/{id}/set
          description: This endpoint configures orchestrator virtual machine
          title: Configures orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/start-put
          method: put
          path: /v1/orchestrator/machines/{id}/start
          description: This endpoint starts orchestrator virtual machine
          title: Starts orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/stop-put
          method: put
          path: /v1/orchestrator/machines/{id}/stop
          description: This endpoint sops orchestrator virtual machine
          title: Stops orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/restart-put
          method: put
          path: /v1/orchestrator/machines/{id}/restart
          description: This endpoint restarts orchestrator virtual machine
          title: Restarts orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/suspend-put
          method: put
          path: /v1/orchestrator/machines/{id}/suspend
          description: This endpoint suspends orchestrator virtual machine
          title: Suspends orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/resume-put
          method: put
          path: /v1/orchestrator/machines/{id}/resume
          description: This endpoint resumes orchestrator virtual machine
          title: Resumes orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/reset-put
          method: put
          path: /v1/orchestrator/machines/{id}/reset
          description: This endpoint resets orchestrator virtual machine
          title: Resets orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/pause-put
          method: put
          path: /v1/orchestrator/machines/{id}/pause
          description: This endpoint pauses orchestrator virtual machine
          title: Pauses orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/clone-put
          method: put
          path: /v1/orchestrator/machines/{id}/clone
          description: This endpoint clones orchestrator virtual machine
          title: Clones orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/execute-put
          method: put
          path: /v1/orchestrator/machines/{id}/execute
          description: This endpoint executes a command in a orchestrator virtual machine
          title: Executes a command in a orchestrator virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines-get
          method: get
          path: /v1/orchestrator/hosts/{id}/machines
          description: This endpoint returns orchestrator host virtual machines
          title: Get orchestrator host virtual machines
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}-get
          method: get
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}
          description: This endpoint returns orchestrator host virtual machine
          title: Get orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}-delete
          method: delete
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}
          description: This endpoint deletes orchestrator host virtual machine
          title: Deletes orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/status-get
          method: get
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/status
          description: This endpoint returns orchestrator host virtual machine status
          title: Get orchestrator host virtual machine status
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/rename-put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/rename
          description: This endpoint renames orchestrator host virtual machine
          title: Renames orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/set-put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/set
          description: This endpoint configures orchestrator host virtual machine
          title: Configures orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/start-put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/start
          description: This endpoint starts orchestrator host virtual machine
          title: Starts orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/stop-put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/stop
          description: This endpoint stops orchestrator host virtual machine
          title: Stops orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/restart-put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/restart
          description: This endpoint restarts orchestrator host virtual machine
          title: Restarts orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/suspend-put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/suspend
          description: This endpoint suspends orchestrator host virtual machine
          title: Suspends orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/resume-put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/resume
          description: This endpoint resumes orchestrator host virtual machine
          title: Resumes orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/reset-put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/reset
          description: This endpoint resets orchestrator host virtual machine
          title: Resets orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/pause-put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/pause
          description: This endpoint pauses orchestrator host virtual machine
          title: Pauses orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/clone-put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/clone
          description: This endpoint clones orchestrator host virtual machine
          title: Clones orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/execute-put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/execute
          description: This endpoint executes a command in a orchestrator host virtual machine
          title: Executes a command in a orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots-get
          method: get
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots
          description: This endpoint lists snapshots of orchestrator host virtual machine
          title: Lists snapshots of orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots-post
          method: post
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots
          description: This endpoint creates a snapshot for orchestrator host virtual machine
          title: Creates a snapshot for orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots-delete
          method: delete
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots
          description: This endpoint deletes all snapshots of orchestrator host virtual machine
          title: Deletes all snapshots of orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}-delete
          method: delete
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}
          description: This endpoint deletes a snapshot of orchestrator host virtual machine
          title: Deletes a snapshot of orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}/revert-post
          method: post
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}/revert
          description: This endpoint reverts orchestrator host virtual machine to a snapshot
          title: Reverts orchestrator host virtual machine to a snapshot
        - anchor: /v1/orchestrator/machines/{id}/snapshots-get
          method: get
          path: /v1/orchestrator/machines/{id}/snapshots
          description: This endpoint lists snapshots of an orchestrator virtual machine (host resolved automatically)
          title: Lists snapshots of an orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/snapshots-post
          method: post
          path: /v1/orchestrator/machines/{id}/snapshots
          description: This endpoint creates a snapshot for an orchestrator virtual machine (host resolved automatically)
          title: Creates a snapshot for an orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/snapshots-delete
          method: delete
          path: /v1/orchestrator/machines/{id}/snapshots
          description: This endpoint deletes all snapshots of an orchestrator virtual machine (host resolved automatically)
          title: Deletes all snapshots of an orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/snapshots/{snapshot_id}-delete
          method: delete
          path: /v1/orchestrator/machines/{id}/snapshots/{snapshot_id}
          description: This endpoint deletes a snapshot of an orchestrator virtual machine (host resolved automatically)
          title: Deletes a snapshot of an orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{id}/snapshots/{snapshot_id}/revert-post
          method: post
          path: /v1/orchestrator/machines/{id}/snapshots/{snapshot_id}/revert
          description: This endpoint reverts an orchestrator virtual machine to a snapshot (host resolved automatically)
          title: Reverts an orchestrator virtual machine to a snapshot
        - anchor: /v1/orchestrator/hosts/{id}/machines/register-post
          method: post
          path: /v1/orchestrator/hosts/{id}/machines/register
          description: This endpoint registers a virtual machine in a orchestrator host
          title: Register a virtual machine in a orchestrator host
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/unregister-post
          method: post
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/unregister
          description: This endpoint unregister a virtual machine in a orchestrator host
          title: Unregister a virtual machine in a orchestrator host
        - anchor: /v1/orchestrator/hosts/{id}/machines-post
          method: post
          path: /v1/orchestrator/hosts/{id}/machines
          description: This endpoint creates a orchestrator host virtual machine
          title: Creates a orchestrator host virtual machine
        - anchor: /v1/orchestrator/machines-post
          method: post
          path: /v1/orchestrator/machines
          description: This endpoint creates a virtual machine in one of the hosts for the orchestrator
          title: Creates a virtual machine in one of the hosts for the orchestrator
        - anchor: /v1/orchestrator/hosts/{id}/reverse-proxy-get
          method: get
          path: /v1/orchestrator/hosts/{id}/reverse-proxy
          description: This endpoint returns orchestrator host reverse proxy configuration
          title: Gets orchestrator host reverse proxy configuration
        - anchor: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts-get
          method: get
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts
          description: This endpoint returns orchestrator host reverse proxy hosts
          title: Gets orchestrator host reverse proxy hosts
        - anchor: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}-get
          method: get
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}
          description: This endpoint returns orchestrator host reverse proxy hosts
          title: Gets orchestrator host reverse proxy hosts
        - anchor: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts-post
          method: post
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts
          description: This endpoint creates a orchestrator host reverse proxy host
          title: Creates a orchestrator host reverse proxy host
        - anchor: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}-put
          method: put
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}
          description: This endpoint updates an orchestrator host reverse proxy host
          title: Updates an orchestrator host reverse proxy host
        - anchor: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}-delete
          method: delete
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}
          description: This endpoint deletes an orchestrator host reverse proxy host
          title: Deletes an orchestrator host reverse proxy host
        - anchor: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes-post
          method: post
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes
          description: This endpoint upserts an orchestrator host reverse proxy host http route
          title: Upserts an orchestrator host reverse proxy host http route
        - anchor: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes/{route_id}-delete
          method: delete
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes/{route_id}
          description: This endpoint deletes an orchestrator host reverse proxy host http route
          title: Deletes an orchestrator host reverse proxy host http route
        - anchor: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/tcp_route-post
          method: post
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/tcp_route
          description: This endpoint updates an orchestrator host reverse proxy host tcp route
          title: Update an orchestrator host reverse proxy host tcp route
        - anchor: /v1/orchestrator/hosts/{id}/reverse-proxy/restart-put
          method: put
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/restart
          description: This endpoint restarts orchestrator host reverse proxy
          title: Restarts orchestrator host reverse proxy
        - anchor: /v1/orchestrator/hosts/{id}/reverse-proxy/enable-put
          method: put
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/enable
          description: This endpoint enables orchestrator host reverse proxy
          title: Enables orchestrator host reverse proxy
        - anchor: /v1/orchestrator/hosts/{id}/reverse-proxy/disable-put
          method: put
          path: /v1/orchestrator/hosts/{id}/reverse-proxy/disable
          description: This endpoint disables orchestrator host reverse proxy
          title: Disables orchestrator host reverse proxy
        - anchor: /v1/orchestrator/hosts/{id}/catalog/cache-get
          method: get
          path: /v1/orchestrator/hosts/{id}/catalog/cache
          description: This endpoint returns orchestrator host catalog cache
          title: Gets orchestrator host catalog cache
        - anchor: /v1/orchestrator/hosts/{id}/catalog/cache-delete
          method: delete
          path: /v1/orchestrator/hosts/{id}/catalog/cache
          description: This endpoint deletes an orchestrator host cache items
          title: Deletes an orchestrator host cache items
        - anchor: /v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}-delete
          method: delete
          path: /v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}
          description: This endpoint deletes an orchestrator host cache item and all its children
          title: Deletes an orchestrator host cache item and all its children
        - anchor: /v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}/{catalog_version}-delete
          method: delete
          path: /v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}/{catalog_version}
          description: This endpoint deletes an orchestrator host cache item and all its children
          title: Deletes an orchestrator host cache item and all its children
        - anchor: /v1/orchestrator/enrollment-token-post
          method: post
          path: /v1/orchestrator/enrollment-token
          description: Generates a short-lived, single-use token that allows a freshly installed agent to register itself with the orchestrator without requiring a permanent credential.
          title: Create an enrollment token
        - anchor: /v1/orchestrator/enrollment-token/{token}/validate-get
          method: get
          path: /v1/orchestrator/enrollment-token/{token}/validate
          description: Public endpoint that checks whether an enrollment token is valid, unused, and not expired. Used by agents before starting the registration flow.
          title: Validate an enrollment token
        - anchor: /v1/orchestrator/hosts/deploy-post
          method: post
          path: /v1/orchestrator/hosts/deploy
          description: SSHes into a remote host, installs the devops agent, and registers it with this orchestrator. Blocks until the operation completes.
          title: Deploy and register an agent via SSH (synchronous)
        - anchor: /v1/orchestrator/hosts/deploy/async-post
          method: post
          path: /v1/orchestrator/hosts/deploy/async
          description: SSHes into a remote host, installs the devops agent, and registers it with this orchestrator. Returns a job ID immediately; poll /jobs/{id} for status.
          title: Deploy and register an agent via SSH (asynchronous)
        - anchor: /v1/orchestrator/machines/async-post
          method: post
          path: /v1/orchestrator/machines/async
          description: This endpoint creates a virtual machine in one of the orchestrator hosts in the background and returns a Job ID to track progress
          title: Creates a virtual machine in one of the orchestrator hosts asynchronously
        - anchor: /v1/orchestrator/hosts/{id}/machines/async-post
          method: post
          path: /v1/orchestrator/hosts/{id}/machines/async
          description: This endpoint creates a virtual machine in a specific orchestrator host in the background and returns a Job ID to track progress
          title: Creates a virtual machine in a specific orchestrator host asynchronously
    - name: Packer Templates
      path: packer_templates
      endpoints:
        - anchor: /v1/templates/packer-get
          method: get
          path: /v1/templates/packer
          description: This endpoint returns all the packer templates. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
          title: Gets all the packer templates
        - anchor: /v1/templates/packer/{id}-get
          method: get
          path: /v1/templates/packer/{id}
          description: This endpoint returns a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
          title: Gets a packer template
        - anchor: /v1/templates/packer -post
          method: post
          path: '/v1/templates/packer '
          description: This endpoint creates a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
          title: Creates a packer template
        - anchor: /v1/templates/packer/{id} -PUT
          method: PUT
          path: '/v1/templates/packer/{id} '
          description: This endpoint updates a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
          title: Updates a packer template
        - anchor: /v1/templates/packer/{id} -DELETE
          method: DELETE
          path: '/v1/templates/packer/{id} '
          description: This endpoint deletes a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
          title: Deletes a packer template
    - name: ReverseProxy
      path: reverseproxy
      endpoints:
        - anchor: /v1/reverse-proxy-get
          method: get
          path: /v1/reverse-proxy
          description: This endpoint returns the reverse proxy configuration
          title: Gets reverse proxy configuration
        - anchor: /v1/reverse-proxy/hosts-get
          method: get
          path: /v1/reverse-proxy/hosts
          description: This endpoint returns all the reverse proxy hosts
          title: Gets all the reverse proxy hosts
        - anchor: /v1/reverse-proxy/hosts/{id} -get
          method: get
          path: '/v1/reverse-proxy/hosts/{id} '
          description: This endpoint returns a reverse proxy host
          title: Gets all the reverse proxy host
        - anchor: /v1/reverse-proxy/hosts-post
          method: post
          path: /v1/reverse-proxy/hosts
          description: This endpoint creates a reverse proxy host
          title: Creates a reverse proxy host
        - anchor: /v1/reverse-proxy/hosts/{id}-put
          method: put
          path: /v1/reverse-proxy/hosts/{id}
          description: This endpoint creates a reverse proxy host
          title: Updates a reverse proxy host
        - anchor: /v1/reverse-proxy/hosts/{id}-delete
          method: delete
          path: /v1/reverse-proxy/hosts/{id}
          description: This endpoint Deletes a reverse proxy host
          title: Delete a a reverse proxy host
        - anchor: /v1/reverse-proxy/hosts/{id}/http_route-post
          method: post
          path: /v1/reverse-proxy/hosts/{id}/http_route
          description: This endpoint creates or updates a reverse proxy host HTTP route
          title: Creates or updates a reverse proxy host HTTP route
        - anchor: /v1/reverse-proxy/hosts/{id}/http_routes/{http_route_id}-delete
          method: delete
          path: /v1/reverse-proxy/hosts/{id}/http_routes/{http_route_id}
          description: This endpoint Deletes a reverse proxy host HTTP route
          title: Delete a a reverse proxy host HTTP route
        - anchor: /v1/reverse-proxy/hosts/{id}/http_routes/order-put
          method: put
          path: /v1/reverse-proxy/hosts/{id}/http_routes/order
          description: This endpoint reorders HTTP routes for a reverse proxy host
          title: Updates the order of a reverse proxy host HTTP route
        - anchor: /v1/reverse-proxy/hosts/{id}/http_routes-post
          method: post
          path: /v1/reverse-proxy/hosts/{id}/http_routes
          description: This endpoint updates a reverse proxy host TCP route
          title: Updates a reverse proxy host TCP route
        - anchor: /v1/reverse-proxy/restart-put
          method: put
          path: /v1/reverse-proxy/restart
          description: This endpoint will restart the reverse proxy
          title: Restarts the reverse proxy
        - anchor: /v1/reverse-proxy/enable-put
          method: put
          path: /v1/reverse-proxy/enable
          description: This endpoint will enable the reverse proxy
          title: Enable the reverse proxy
        - anchor: /v1/reverse-proxy/disable-put
          method: put
          path: /v1/reverse-proxy/disable
          description: This endpoint will disable the reverse proxy
          title: Disable the reverse proxy
    - name: Roles
      path: roles
      endpoints:
        - anchor: /v1/auth/roles -get
          method: get
          path: '/v1/auth/roles '
          description: This endpoint returns all the roles
          title: Gets all the roles
        - anchor: /v1/auth/roles/{id} -get
          method: get
          path: '/v1/auth/roles/{id} '
          description: This endpoint returns a role
          title: Gets a role
        - anchor: /v1/auth/roles -post
          method: post
          path: '/v1/auth/roles '
          description: This endpoint creates a role
          title: Creates a role
        - anchor: /v1/auth/roles/{id} -delete
          method: delete
          path: '/v1/auth/roles/{id} '
          description: This endpoint deletes a role
          title: Delete a role
        - anchor: /v1/auth/roles/{id}/claims -get
          method: get
          path: '/v1/auth/roles/{id}/claims '
          description: This endpoint returns all claims associated with a role
          title: Gets all claims for a role
        - anchor: /v1/auth/roles/{id}/claims -post
          method: post
          path: '/v1/auth/roles/{id}/claims '
          description: This endpoint adds a claim to a role
          title: Adds a claim to a role
        - anchor: /v1/auth/roles/{id}/claims/{claim_id} -delete
          method: delete
          path: '/v1/auth/roles/{id}/claims/{claim_id} '
          description: This endpoint removes a claim from a role
          title: Removes a claim from a role
    - name: SSH
      path: ssh
      endpoints:
        - anchor: /v1/ssh/execute-post
          method: post
          path: /v1/ssh/execute
          description: Executes a command on a remote host via SSH
          title: Execute SSH Command
    - name: User Configs
      path: user_configs
      endpoints:
        - anchor: /v1/user/configs-get
          method: get
          path: /v1/user/configs
          description: This endpoint returns all configuration entries for the authenticated user
          title: Gets all user configs
        - anchor: /v1/user/configs/{id}-get
          method: get
          path: /v1/user/configs/{id}
          description: This endpoint returns a single configuration entry for the authenticated user
          title: Gets a user config by id or slug
        - anchor: /v1/user/configs-post
          method: post
          path: /v1/user/configs
          description: This endpoint creates a configuration entry for the authenticated user
          title: Creates a user config
        - anchor: /v1/user/configs/{id}-put
          method: put
          path: /v1/user/configs/{id}
          description: This endpoint updates a configuration entry for the authenticated user
          title: Updates a user config
        - anchor: /v1/user/configs/{id}-delete
          method: delete
          path: /v1/user/configs/{id}
          description: This endpoint deletes a configuration entry for the authenticated user
          title: Deletes a user config
    - name: Users
      path: users
      endpoints:
        - anchor: /v1/auth/users -get
          method: get
          path: '/v1/auth/users '
          description: This endpoint returns all the users
          title: Gets all the users
        - anchor: /v1/auth/users/{id} -get
          method: get
          path: '/v1/auth/users/{id} '
          description: This endpoint returns a user
          title: Gets a user
        - anchor: /v1/auth/users -post
          method: post
          path: '/v1/auth/users '
          description: This endpoint creates a user
          title: Creates a user
        - anchor: /v1/auth/users/{id} -delete
          method: delete
          path: '/v1/auth/users/{id} '
          description: This endpoint deletes a user
          title: Deletes a user
        - anchor: /v1/auth/users/{id} -put
          method: put
          path: '/v1/auth/users/{id} '
          description: This endpoint updates a user
          title: Update a user
        - anchor: /v1/auth/users/{id}/roles -get
          method: get
          path: '/v1/auth/users/{id}/roles '
          description: This endpoint returns all the roles for a user
          title: Gets all the roles for a user
        - anchor: /v1/auth/users/{id}/roles -post
          method: post
          path: '/v1/auth/users/{id}/roles '
          description: This endpoint adds a role to a user
          title: Adds a role to a user
        - anchor: /v1/auth/users/{id}/roles/{role_id} -delete
          method: delete
          path: '/v1/auth/users/{id}/roles/{role_id} '
          description: This endpoint removes a role from a user
          title: Removes a role from a user
        - anchor: /v1/auth/users/{id}/claims -get
          method: get
          path: '/v1/auth/users/{id}/claims '
          description: This endpoint returns all the claims for a user
          title: Gets all the claims for a user
        - anchor: /v1/auth/users/{id}/claims -post
          method: post
          path: '/v1/auth/users/{id}/claims '
          description: This endpoint adds a claim to a user
          title: Adds a claim to a user
        - anchor: /v1/auth/users/{id}/claims/{claim_id} -delete
          method: delete
          path: '/v1/auth/users/{id}/claims/{claim_id} '
          description: This endpoint removes a claim from a user
          title: Removes a claim from a user
endpoints:
    - title: Gets all the roles
      description: This endpoint returns all the roles
      requires_authorization: true
      category: Roles
      category_path: roles
      path: '/v1/auth/roles '
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.RoleResponse'
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/auth/roles ' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/auth/roles ");
            request.Headers.Add("Authorization", "••••••");
            var response = await client.SendAsync(request);
            response.EnsureSuccessStatusCode();
            var responseString = await response.Content.ReadAsStringAsync();
          title: C#
          language: csharp
        - code_block: |
            package main

            import (
              "fmt"
              "net/http"
              "strings"
              "io"
            )

            func main() {
              url := "http://localhost/api/v1/auth/roles "
              method := "get"
              client := &http.Client{}
              req, err := http.NewRequest(method, url, payload)
              if err != nil {
                fmt.Println(err)
                return
              }
              req.Header.Add("Content-Type", "application/json")

              req.Header.Add("Authorization", "••••••")
              res, err := client.Do(req)
              if err != nil {
                fmt.Println(err)
                return
              }
              defer res.Body.Close()
              body, err := io.ReadAll(res.Body)
              if err != nil {
                fmt.Println(err)
                return
              }
              fmt.Println(string(body))
            }
          title: Go
          language: go
    - title: Gets a role
      description: This endpoint returns a role
      requires_authorization: true
      category: Roles
      category_path: roles
      path: '/v1/auth/roles/{id} '
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Role ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.RoleResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/auth/roles/{id} ' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/auth/roles/{id} ");
            request.Headers.Add("Authorization", "••••••");
            var response = await client.SendAsync(request);
            response.EnsureSuccessStatusCode();
            var responseString = await response.Content.ReadAsStringAsync();
          title: C#
          language: csharp
        - code_block: |
            package main

            import (
              "fmt"
              "net/http"
              "strings"
              "io"
            )

            func main() {
              url := "http://localhost/api/v1/auth/roles/{id} "
              method := "get"
              client := &http.Client{}
              req, err := http.NewRequest(method, url, payload)
              if err != nil {
                fmt.Println(err)
                return
              }
              req.Header.Add("Content-Type", "application/json")

              req.Header.Add("Authorization", "••••••")
              res, err := client.Do(req)
              if err != nil {
                fmt.Println(err)
                return
              }
              defer res.Body.Close()
              body, err := io.ReadAll(res.Body)
              if err != nil {
                fmt.Println(err)
                return
              }
              fmt.Println(string(body))
            }
          title: Go
          language: go
    - title: Creates a role
      description: This endpoint creates a role
      requires_authorization: true
      category: Roles
      category_path: roles
      path: '/v1/auth/roles '
      method: post
      parameters:
        - name: roleRequest
          required: false
          type: body
          value_type: object
          description: Role Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "201"
          code_description: Created
          title: models.RoleResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/auth/roles ' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/auth/roles ");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{ object }");
            request.Content = content;
            var response = await client.SendAsync(request);
            response.EnsureSuccessStatusCode();
            var responseString = await response.Content.ReadAsStringAsync();
          title: C#
          language: csharp
        - code_block: |
            package main

            import (
              "fmt"
              "net/http"
              "strings"
              "io"
            )

            func main() {
              url := "http://localhost/api/v1/auth/roles "
              method := "post"
              payload := strings.NewReader(`{ object }`)
              client := &http.Client{}
              req, err := http.NewRequest(method, url, payload)
              if err != nil {
                fmt.Println(err)
                return
              }
              req.Header.Add("Content-Type", "application/json")

              req.Header.Add("Authorization", "••••••")
              res, err := client.Do(req)
              if err != nil {
                fmt.Println(err)
                return
              }
              defer res.Body.Close()
              body, err := io.ReadAll(res.Body)
              if err != nil {
                fmt.Println(err)
                return
              }
              fmt.Println(string(body))
            }
          title: Go
          language: go
    - title: Delete a role
      description: This endpoint deletes a role
      requires_authorization: true
      category: Roles
      category_path: roles
      path: '/v1/auth/roles/{id} '
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Role ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/auth/roles/{id} ' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/auth/roles/{id} ");
            request.Headers.Add("Authorization", "••••••");
            var response = await client.SendAsync(request);
            response.EnsureSuccessStatusCode();
            var responseString = await response.Content.ReadAsStringAsync();
          title: C#
          language: csharp
        - code_block: |
            package main

            import (
              "fmt"
              "net/http"
              "strings"
              "io"
            )

            func main() {
              url := "http://localhost/api/v1/auth/roles/{id} "
              method := "delete"
              client := &http.Client{}
              req, err := http.NewRequest(method, url, payload)
              if err != nil {
                fmt.Println(err)
                return
              }
              req.Header.Add("Content-Type", "application/json")

              req.Header.Add("Authorization", "••••••")
              res, err := client.Do(req)
              if err != nil {
                fmt.Println(err)
                return
              }
              defer res.Body.Close()
              body, err := io.ReadAll(res.Body)
              if err != nil {
                fmt.Println(err)
                return
              }
              fmt.Println(string(body))
            }
          title: Go
          language: go
    - title: Gets all claims for a role
      description: This endpoint returns all claims associated with a role
      requires_authorization: true
      category: Roles
      category_path: roles
      path: '/v1/auth/roles/{id}/claims '
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Role ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.ClaimResponse'
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/auth/roles/{id}/claims ' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/auth/roles/{id}/claims ");
            request.Headers.Add("Authorization", "••••••");
            var response = await client.SendAsync(request);
            response.EnsureSuccessStatusCode();
            var responseString = await response.Content.ReadAsStringAsync();
          title: C#
          language: csharp
        - code_block: |
            package main

            import (
              "fmt"
              "net/http"
              "strings"
              "io"
            )

            func main() {
              url := "http://localhost/api/v1/auth/roles/{id}/claims "
              method := "get"
              client := &http.Client{}
              req, err := http.NewRequest(method, url, payload)
              if err != nil {
                fmt.Println(err)
                return
              }
              req.Header.Add("Content-Type", "application/json")

              req.Header.Add("Authorization", "••••••")
              res, err := client.Do(req)
              if err != nil {
                fmt.Println(err)
                return
              }
              defer res.Body.Close()
              body, err := io.ReadAll(res.Body)
              if err != nil {
                fmt.Println(err)
                return
              }
              fmt.Println(string(body))
            }
          title: Go
          language: go
    - title: Adds a claim to a role
      description: This endpoint adds a claim to a role
      requires_authorization: true
      category: Roles
      category_path: roles
      path: '/v1/auth/roles/{id}/claims '
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Role ID
        - name: body
          required: false
          type: body
          value_type: object
          description: Claim Name
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "201"
          code_description: Created
          title: models.ClaimResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/auth/roles/{id}/claims ' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/auth/roles/{id}/claims ");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{ object }");
            request.Content = content;
            var response = await client.SendAsync(request);
            response.EnsureSuccessStatusCode();
            var responseString = await response.Content.ReadAsStringAsync();
          title: C#
          language: csharp
        - code_block: |
            package main

            import (
              "fmt"
              "net/http"
              "strings"
              "io"
            )

            func main() {
              url := "http://localhost/api/v1/auth/roles/{id}/claims "
              method := "post"
              payload := strings.NewReader(`{ object }`)
              client := &http.Client{}
              req, err := http.NewRequest(method, url, payload)
              if err != nil {
                fmt.Println(err)
                return
              }
              req.Header.Add("Content-Type", "application/json")

              req.Header.Add("Authorization", "••••••")
              res, err := client.Do(req)
              if err != nil {
                fmt.Println(err)
                return
              }
              defer res.Body.Close()
              body, err := io.ReadAll(res.Body)
              if err != nil {
                fmt.Println(err)
                return
              }
              fmt.Println(string(body))
            }
          title: Go
          language: go
    - title: Removes a claim from a role
      description: This endpoint removes a claim from a role
      requires_authorization: true
      category: Roles
      category_path: roles
      path: '/v1/auth/roles/{id}/claims/{claim_id} '
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Role ID
        - name: claim_id
          required: true
          type: path
          value_type: string
          description: Claim ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/auth/roles/{id}/claims/{claim_id} ' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/auth/roles/{id}/claims/{claim_id} ");
            request.Headers.Add("Authorization", "••••••");
            var response = await client.SendAsync(request);
            response.EnsureSuccessStatusCode();
            var responseString = await response.Content.ReadAsStringAsync();
          title: C#
          language: csharp
        - code_block: |
            package main

            import (
              "fmt"
              "net/http"
              "strings"
              "io"
            )

            func main() {
              url := "http://localhost/api/v1/auth/roles/{id}/claims/{claim_id} "
              method := "delete"
              client := &http.Client{}
              req, err := http.NewRequest(method, url, payload)
              if err != nil {
                fmt.Println(err)
                return
              }
              req.Header.Add("Content-Type", "application/json")

              req.Header.Add("Authorization", "••••••")
              res, err := client.Do(req)
              if err != nil {
                fmt.Println(err)
                return
              }
              defer res.Body.Close()
              body, err := io.ReadAll(res.Body)
              if err != nil {
                fmt.Println(err)
                return
              }
              fmt.Println(string(body))
            }
          title: Go
          language: go

---
# Roles endpoints 

 This document contains the endpoints for the Roles category.


