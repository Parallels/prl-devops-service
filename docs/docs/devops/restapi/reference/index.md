---
layout: api
title: API Documentation
default_host: http://localhost
api_prefix: /api
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
    - title: Creates an api key
      description: This endpoint creates an api key
      requires_authorization: true
      category: Api Keys
      category_path: api_keys
      path: /v1/auth/api_keys
      method: post
      headers:
        - name: x-filter
          required: false
          type: header
          value_type: string
          description: Filter entities
      parameters:
        - name: apiKey
          required: false
          type: body
          value_type: object
          description: Body
          body: '{ object }'
      default_required_roles:
        - SUPER_USER
      default_required_claims:
        - CREATE_API_KEY
        - LIST
      content_markdown: |-
        # This endpoint will create an api key in the system
        API Keys are used to authenticate with the system from external applications
        ## How are they different from a user?
        A user normally has a password and is used to authenticate with the system
        An api key is used to authenticate with the system from an external application
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ApiKeyResponse
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/api_keys' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"key\": \"Some Key\",\n  \"secret\": \"SomeLongSecret\"\n  }\n'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/auth/api_keys");
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
              url := "http://localhost/api/v1/auth/api_keys"
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
    - title: Gets all the api keys
      description: This endpoint returns all the api keys
      requires_authorization: true
      category: Api Keys
      category_path: api_keys
      path: /v1/auth/api_keys
      method: get
      default_required_claims:
        - LIST_API_KEY
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.ApiKeyResponse'
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/api_keys' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/auth/api_keys");
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
              url := "http://localhost/api/v1/auth/api_keys"
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
    - title: Deletes an api key
      description: This endpoint deletes an api key
      requires_authorization: true
      category: Api Keys
      category_path: api_keys
      path: /v1/auth/api_keys/{id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Api Key ID
      default_required_claims:
        - DELETE_API_KEY
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/api_keys/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/auth/api_keys/{id}");
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
              url := "http://localhost/api/v1/auth/api_keys/{id}"
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
    - title: Gets an api key by id or name
      description: This endpoint returns an api key by id or name
      requires_authorization: true
      category: Api Keys
      category_path: api_keys
      path: /v1/auth/api_keys/{id}
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Api Key ID
      default_required_claims:
        - LIST_API_KEY
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ApiKeyResponse
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/api_keys/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/auth/api_keys/{id}");
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
              url := "http://localhost/api/v1/auth/api_keys/{id}"
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
    - title: Revoke an api key
      description: This endpoint revokes an api key
      requires_authorization: true
      category: Api Keys
      category_path: api_keys
      path: /v1/auth/api_keys/{id}/revoke
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Api Key ID
      default_required_roles:
        - SUPER_USER
      default_required_claims:
        - LIST_API_KEY
        - DELETE_API_KEY
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/api_keys/{id}/revoke' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/auth/api_keys/{id}/revoke");
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
              url := "http://localhost/api/v1/auth/api_keys/{id}/revoke"
              method := "put"
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
    - title: Generates a token
      description: This endpoint generates a token
      requires_authorization: true
      category: Authorization
      category_path: authorization
      path: /v1/auth/token
      method: post
      parameters:
        - name: login
          required: false
          type: body
          value_type: object
          description: Body
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.LoginResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.ApiErrorDiagnosticsResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/auth/token' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/auth/token");
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
              url := "http://localhost/api/v1/auth/token"
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
    - title: Validates a token
      description: This endpoint validates a token
      requires_authorization: true
      category: Authorization
      category_path: authorization
      path: /v1/auth/token/validate
      method: post
      parameters:
        - name: tokenRequest
          required: false
          type: body
          value_type: object
          description: Body
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ValidateTokenResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.ApiErrorDiagnosticsResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/auth/token/validate' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/auth/token/validate");
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
              url := "http://localhost/api/v1/auth/token/validate"
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
    - title: Gets catalog cache
      description: This endpoint returns all the remote catalog cache if any
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/cache
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.CatalogManifest'
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
        - code_block: "curl --location 'http://localhost/api/v1/cache' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/cache");
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
              url := "http://localhost/api/v1/cache"
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
    - title: Deletes all catalog cache
      description: This endpoint returns all the remote catalog cache if any
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/cache
      method: delete
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
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
        - code_block: "curl --location 'http://localhost/api/v1/cache' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/cache");
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
              url := "http://localhost/api/v1/cache"
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
    - title: Deletes catalog cache item and all its versions
      description: This endpoint returns all the remote catalog cache if any and all its versions
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/cache/{catalogId}
      method: delete
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
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
        - code_block: "curl --location 'http://localhost/api/v1/cache/{catalogId}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/cache/{catalogId}");
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
              url := "http://localhost/api/v1/cache/{catalogId}"
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
    - title: Deletes catalog cache version item
      description: This endpoint deletes a version of a cache ite,
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/cache/{catalogId}/{version}
      method: delete
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
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
        - code_block: "curl --location 'http://localhost/api/v1/cache/{catalogId}/{version}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/cache/{catalogId}/{version}");
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
              url := "http://localhost/api/v1/cache/{catalogId}/{version}"
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
    - title: Gets all the remote catalogs
      description: This endpoint returns all the remote catalogs
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]map[string][]models.CatalogManifest'
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/catalog");
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
              url := "http://localhost/api/v1/catalog"
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
    - title: Gets all the remote catalogs
      description: This endpoint returns all the remote catalogs
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}
      method: get
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.CatalogManifest'
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/catalog/{catalogId}");
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
              url := "http://localhost/api/v1/catalog/{catalogId}"
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
    - title: Gets a catalog manifest version
      description: This endpoint returns a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}
      method: get
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/catalog/{catalogId}/{version}");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}"
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
    - title: Gets a catalog manifest version architecture
      description: This endpoint returns a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}/{architecture}
      method: get
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
        - name: architecture
          required: true
          type: path
          value_type: string
          description: Architecture
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}"
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
    - title: Downloads a catalog manifest version
      description: This endpoint downloads a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}/{architecture}/download
      method: get
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
        - name: architecture
          required: true
          type: path
          value_type: string
          description: Architecture
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/download' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/download");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/download"
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
    - title: Taints a catalog manifest version
      description: This endpoint Taints a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}/{architecture}/taint
      method: patch
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
        - name: architecture
          required: true
          type: path
          value_type: string
          description: Architecture
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/taint' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Patch, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/taint");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/taint"
              method := "patch"
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
    - title: UnTaints a catalog manifest version
      description: This endpoint UnTaints a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}/{architecture}/untaint
      method: patch
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
        - name: architecture
          required: true
          type: path
          value_type: string
          description: Architecture
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/untaint' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Patch, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/untaint");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/untaint"
              method := "patch"
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
    - title: UnTaints a catalog manifest version
      description: This endpoint UnTaints a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}/{architecture}/revoke
      method: patch
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
        - name: architecture
          required: true
          type: path
          value_type: string
          description: Architecture
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/revoke' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Patch, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/revoke");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/revoke"
              method := "patch"
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
    - title: Adds claims to a catalog manifest version
      description: This endpoint adds claims to a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}/{architecture}/claims
      method: patch
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
        - name: architecture
          required: true
          type: path
          value_type: string
          description: Architecture
        - name: request
          required: false
          type: body
          value_type: object
          description: Body
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Patch, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims"
              method := "patch"
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
    - title: Removes claims from a catalog manifest version
      description: This endpoint removes claims from a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}/{architecture}/claims
      method: delete
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
        - name: architecture
          required: true
          type: path
          value_type: string
          description: Architecture
        - name: request
          required: false
          type: body
          value_type: object
          description: Body
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims"
              method := "delete"
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
    - title: Adds roles to a catalog manifest version
      description: This endpoint adds roles to a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}/{architecture}/roles
      method: patch
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
        - name: architecture
          required: true
          type: path
          value_type: string
          description: Architecture
        - name: request
          required: false
          type: body
          value_type: object
          description: Body
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/roles' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Patch, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/roles");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/roles"
              method := "patch"
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
    - title: Removes roles from a catalog manifest version
      description: This endpoint removes roles from a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}/{architecture}/roles
      method: delete
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
        - name: architecture
          required: true
          type: path
          value_type: string
          description: Architecture
        - name: request
          required: false
          type: body
          value_type: object
          description: Body
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/roles' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/roles");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/roles"
              method := "delete"
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
    - title: Adds tags to a catalog manifest version
      description: This endpoint adds tags to a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}/{architecture}/tags
      method: patch
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
        - name: architecture
          required: true
          type: path
          value_type: string
          description: Architecture
        - name: request
          required: false
          type: body
          value_type: object
          description: Body
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/tags' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Patch, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/tags");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/tags"
              method := "patch"
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
    - title: Removes tags from a catalog manifest version
      description: This endpoint removes tags from a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}/{architecture}/tags
      method: delete
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
        - name: architecture
          required: true
          type: path
          value_type: string
          description: Architecture
        - name: request
          required: false
          type: body
          value_type: object
          description: Body
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/tags' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/tags");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/tags"
              method := "delete"
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
    - title: Deletes a catalog manifest and all its versions
      description: This endpoint deletes a catalog manifest and all its versions
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}
      method: delete
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/catalog/{catalogId}");
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
              url := "http://localhost/api/v1/catalog/{catalogId}"
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
    - title: Deletes a catalog manifest version
      description: This endpoint deletes a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}
      method: delete
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/catalog/{catalogId}/{version}");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}"
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
    - title: Deletes a catalog manifest version architecture
      description: This endpoint deletes a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}/{architecture}
      method: delete
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
        - name: architecture
          required: true
          type: path
          value_type: string
          description: Architecture
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}"
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
    - title: Pushes a catalog manifest to the catalog inventory
      description: This endpoint pushes a catalog manifest to the catalog inventory
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/push
      method: post
      parameters:
        - name: pushRequest
          required: false
          type: body
          value_type: object
          description: Push request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/push' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/catalog/push");
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
              url := "http://localhost/api/v1/catalog/push"
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
    - title: Push a catalog manifest to the catalog inventory asynchronously
      description: This endpoint pushes a catalog manifest to the catalog inventory in the background and returns a Job ID to track progress
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/push/async
      method: post
      parameters:
        - name: pushRequest
          required: false
          type: body
          value_type: object
          description: Push request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "202"
          code_description: Accepted
          title: models.JobResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/push/async' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/catalog/push/async");
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
              url := "http://localhost/api/v1/catalog/push/async"
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
    - title: Pull a remote catalog manifest
      description: This endpoint pulls a remote catalog manifest
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/pull
      method: put
      parameters:
        - name: pullRequest
          required: false
          type: body
          value_type: object
          description: Pull request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.PullCatalogManifestResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/pull' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/catalog/pull");
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
              url := "http://localhost/api/v1/catalog/pull"
              method := "put"
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
    - title: Pull a remote catalog manifest asynchronously
      description: This endpoint pulls a remote catalog manifest in the background and returns a Job ID to track progress
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/pull/async
      method: put
      parameters:
        - name: pullRequest
          required: false
          type: body
          value_type: object
          description: Pull request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "202"
          code_description: Accepted
          title: models.JobResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/pull/async' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/catalog/pull/async");
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
              url := "http://localhost/api/v1/catalog/pull/async"
              method := "put"
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
    - title: Imports a remote catalog manifest metadata into the catalog inventory
      description: This endpoint imports a remote catalog manifest metadata into the catalog inventory
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/import
      method: put
      parameters:
        - name: importRequest
          required: false
          type: body
          value_type: object
          description: Pull request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ImportCatalogManifestResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/import' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/catalog/import");
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
              url := "http://localhost/api/v1/catalog/import"
              method := "put"
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
    - title: Imports a vm into the catalog inventory generating the metadata for it
      description: This endpoint imports a virtual machine in pvm or macvm format into the catalog inventory generating the metadata for it
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/import-vm
      method: put
      parameters:
        - name: importRequest
          required: false
          type: body
          value_type: object
          description: Vm Impoty request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ImportVmResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/import-vm' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/catalog/import-vm");
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
              url := "http://localhost/api/v1/catalog/import-vm"
              method := "put"
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
    - title: Updates a catalog
      description: This endpoint adds claims to a catalog manifest version
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}/{architecture}/claims
      method: patch
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: request
          required: false
          type: body
          value_type: object
          description: Body
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Patch, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims"
              method := "patch"
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
    - title: Updates metadata for a catalog manifest version
      description: This endpoint atomically updates description, tags, required claims, and required roles for a catalog manifest version. Omit a field to leave it unchanged.
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/{catalogId}/{version}/{architecture}/metadata
      method: put
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: version
          required: true
          type: path
          value_type: string
          description: Version
        - name: architecture
          required: true
          type: path
          value_type: string
          description: Architecture
        - name: request
          required: false
          type: body
          value_type: object
          description: Body
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/metadata' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/metadata");
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/metadata"
              method := "put"
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
    - title: Gets all the catalog managers
      description: This endpoint returns all the catalog managers
      requires_authorization: true
      category: CatalogManagers
      category_path: catalogmanagers
      path: /v1/catalog-managers
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.CatalogManager'
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog-managers' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/catalog-managers");
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
              url := "http://localhost/api/v1/catalog-managers"
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
    - title: Gets a specific catalog manager
      description: This endpoint returns a catalog manager
      requires_authorization: true
      category: CatalogManagers
      category_path: catalogmanagers
      path: /v1/catalog-managers/{id}
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Manager ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManager
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog-managers/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/catalog-managers/{id}");
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
              url := "http://localhost/api/v1/catalog-managers/{id}"
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
    - title: Creates a catalog manager
      description: This endpoint creates a catalog manager
      requires_authorization: true
      category: CatalogManagers
      category_path: catalogmanagers
      path: /v1/catalog-managers
      method: post
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManager
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog-managers' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/catalog-managers");
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
              url := "http://localhost/api/v1/catalog-managers"
              method := "post"
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
    - title: Updates a catalog manager
      description: This endpoint updates a catalog manager
      requires_authorization: true
      category: CatalogManagers
      category_path: catalogmanagers
      path: /v1/catalog-managers/{id}
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Manager ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CatalogManager
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog-managers/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/catalog-managers/{id}");
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
              url := "http://localhost/api/v1/catalog-managers/{id}"
              method := "put"
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
    - title: Deletes a catalog manager
      description: This endpoint deletes a catalog manager
      requires_authorization: true
      category: CatalogManagers
      category_path: catalogmanagers
      path: /v1/catalog-managers/{id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Manager ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ApiCommonResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog-managers/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/catalog-managers/{id}");
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
              url := "http://localhost/api/v1/catalog-managers/{id}"
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
    - title: Gets all the claims
      description: This endpoint returns all the claims
      requires_authorization: true
      category: Claims
      category_path: claims
      path: /v1/auth/claims
      method: get
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/claims' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/auth/claims");
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
              url := "http://localhost/api/v1/auth/claims"
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
    - title: Gets all claims grouped for the matrix UI
      description: This endpoint returns all claims organised by group and resource
      requires_authorization: true
      category: Claims
      category_path: claims
      path: /v1/auth/claims/grouped
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.ClaimGroupResponse'
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/claims/grouped' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/auth/claims/grouped");
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
              url := "http://localhost/api/v1/auth/claims/grouped"
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
    - title: Gets a claim
      description: This endpoint returns a claim
      requires_authorization: true
      category: Claims
      category_path: claims
      path: /v1/auth/claims/{id}
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Claim ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/claims/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/auth/claims/{id}");
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
              url := "http://localhost/api/v1/auth/claims/{id}"
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
    - title: Creates a claim
      description: This endpoint creates a claim
      requires_authorization: true
      category: Claims
      category_path: claims
      path: /v1/auth/claims
      method: post
      parameters:
        - name: claimRequest
          required: false
          type: body
          value_type: object
          description: Claim Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/claims' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/auth/claims");
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
              url := "http://localhost/api/v1/auth/claims"
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
    - title: Delete a claim
      description: This endpoint Deletes a claim
      requires_authorization: true
      category: Claims
      category_path: claims
      path: /v1/auth/claims/{id}
      method: delete
      parameters:
        - name: id
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/claims/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/auth/claims/{id}");
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
              url := "http://localhost/api/v1/auth/claims/{id}"
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
    - title: Gets Parallels Desktop active license
      description: This endpoint returns Parallels Desktop active license
      requires_authorization: true
      category: Config
      category_path: config
      path: /v1/parallels_desktop/key
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ParallelsDesktopLicense
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/parallels_desktop/key' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/parallels_desktop/key");
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
              url := "http://localhost/api/v1/parallels_desktop/key"
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
    - title: Installs API requires 3rd party tools
      description: This endpoint installs API requires 3rd party tools
      requires_authorization: true
      category: Config
      category_path: config
      path: /v1/config/tools/install
      method: post
      parameters:
        - name: installToolsRequest
          required: false
          type: body
          value_type: object
          description: Install Tools Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.InstallToolsResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/config/tools/install' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/config/tools/install");
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
              url := "http://localhost/api/v1/config/tools/install"
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
    - title: Uninstalls API requires 3rd party tools
      description: This endpoint uninstalls API requires 3rd party tools
      requires_authorization: true
      category: Config
      category_path: config
      path: /v1/config/tools/uninstall
      method: post
      parameters:
        - name: uninstallToolsRequest
          required: false
          type: body
          value_type: object
          description: Uninstall Tools Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.InstallToolsResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/config/tools/uninstall' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/config/tools/uninstall");
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
              url := "http://localhost/api/v1/config/tools/uninstall"
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
    - title: Restarts the API Service
      description: This endpoint restarts the API Service
      requires_authorization: true
      category: Config
      category_path: config
      path: /v1/config/tools/restart
      method: post
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/config/tools/restart' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/config/tools/restart");
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
              url := "http://localhost/api/v1/config/tools/restart"
              method := "post"
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
    - title: Gets the Hardware Info
      description: This endpoint returns the Hardware Info
      requires_authorization: true
      category: Config
      category_path: config
      path: /v1/config/hardware
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.SystemUsageResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/config/hardware' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/config/hardware");
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
              url := "http://localhost/api/v1/config/hardware"
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
    - title: Gets the API System Health
      description: This endpoint returns the API Health Probe
      requires_authorization: true
      category: Config
      category_path: config
      path: /health/system
      method: get
      parameters:
        - name: full
          required: false
          type: query
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ServiceHealthCheck
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/health/system' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/health/system");
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
              url := "http://localhost/api/health/system"
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
    - title: Gets the system logs from the disk
      description: This endpoint returns the system logs from the disk
      requires_authorization: true
      category: Config
      category_path: config
      path: /logs
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/logs' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/logs");
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
              url := "http://localhost/api/logs"
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
    - title: Streams the system logs via WebSocket
      description: This endpoint streams the system logs in real-time via WebSocket
      requires_authorization: true
      category: Config
      category_path: config
      path: /logs/stream
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/logs/stream' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/logs/stream");
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
              url := "http://localhost/api/logs/stream"
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
    - title: Gets the Parallels disk space information
      description: This endpoint returns the available disk space for the cache folder.
      requires_authorization: true
      category: Config
      category_path: config
      path: /config/diskspace
      method: post
      parameters:
        - name: createRequest
          required: false
          type: body
          value_type: object
          description: Disk Space Available Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.DiskSpaceAvailable
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/config/diskspace' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/config/diskspace");
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
              url := "http://localhost/api/config/diskspace"
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
    - title: Subscribe to event notifications via WebSocket
      description: This endpoint upgrades the HTTP connection to WebSocket and subscribes to event notifications. Authentication is required via Authorization header (Bearer token) or query parameters (access_token or authorization).
      requires_authorization: true
      category: Events
      category_path: events
      path: /v1/ws/subscribe
      method: get
      parameters:
        - name: event_types
          required: false
          type: query
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "409"
          code_description: Conflict
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "503"
          code_description: Service Unavailable
          title: models.ApiErrorDiagnosticsResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/ws/subscribe' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/ws/subscribe");
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
              url := "http://localhost/api/v1/ws/subscribe"
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
    - title: List connected WebSocket clients
      description: Returns all currently connected WebSocket clients with queue depth and ping/pong timestamps. Useful for diagnosing stale or dead clients whose queues are filling up.
      requires_authorization: true
      category: Events
      category_path: events
      path: /v1/ws/clients
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "503"
          code_description: Service Unavailable
          title: models.ApiErrorDiagnosticsResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/ws/clients' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/ws/clients");
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
              url := "http://localhost/api/v1/ws/clients"
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
    - title: Get WebSocket event emitter statistics
      description: Returns aggregate statistics including total connected clients, subscription counts per event type, uptime, and per-client details with queue depths.
      requires_authorization: true
      category: Events
      category_path: events
      path: /v1/ws/stats
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.EventEmitterStats
          language: json
        - code_block: '{ object }'
          code: "503"
          code_description: Service Unavailable
          title: models.ApiErrorDiagnosticsResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/ws/stats' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/ws/stats");
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
              url := "http://localhost/api/v1/ws/stats"
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
    - title: Unsubscribe from specific event types
      description: Unsubscribe an active WebSocket client from specific event types without disconnecting. The client must belong to the authenticated user.
      requires_authorization: true
      category: Events
      category_path: events
      path: /v1/ws/unsubscribe
      method: post
      parameters:
        - name: body
          required: false
          type: body
          value_type: object
          description: Unsubscribe request with client ID and event types
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "503"
          code_description: Service Unavailable
          title: models.ApiErrorDiagnosticsResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/ws/unsubscribe' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/ws/unsubscribe");
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
              url := "http://localhost/api/v1/ws/unsubscribe"
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
    - title: Deletes a job by ID
      description: This endpoint deletes a single job. Users with JOB_MANAGER_DELETE can delete any job; users with JOB_MANAGER_LIST_OWN can only delete their own.
      requires_authorization: true
      category: Jobs
      category_path: jobs
      path: /v1/jobs/{id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Job ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
        - code_block: '{ object }'
          code: "403"
          code_description: Forbidden
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "404"
          code_description: Not Found
          title: models.ApiErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/jobs/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/jobs/{id}");
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
              url := "http://localhost/api/v1/jobs/{id}"
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
    - title: Gets all the virtual machines
      description: This endpoint returns all the virtual machines
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines
      method: get
      parameters:
        - name: filter
          required: false
          type: header
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.ParallelsVM'
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/machines");
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
              url := "http://localhost/api/v1/machines"
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
    - title: Gets a virtual machine
      description: This endpoint returns a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ParallelsVM
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/machines/{id}");
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
              url := "http://localhost/api/v1/machines/{id}"
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
    - title: Starts a virtual machine
      description: This endpoint starts a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/start
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/start' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/machines/{id}/start");
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
              url := "http://localhost/api/v1/machines/{id}/start"
              method := "put"
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
    - title: Stops a virtual machine
      description: This endpoint stops a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/stop
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
        - name: force
          required: false
          type: query
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/stop' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/machines/{id}/stop");
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
              url := "http://localhost/api/v1/machines/{id}/stop"
              method := "put"
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
    - title: Restarts a virtual machine
      description: This endpoint restarts a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/restart
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/restart' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/machines/{id}/restart");
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
              url := "http://localhost/api/v1/machines/{id}/restart"
              method := "put"
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
    - title: Suspends a virtual machine
      description: This endpoint suspends a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/suspend
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/suspend' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/machines/{id}/suspend");
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
              url := "http://localhost/api/v1/machines/{id}/suspend"
              method := "put"
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
    - title: Resumes a virtual machine
      description: This endpoint resumes a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/resume
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/resume' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/machines/{id}/resume");
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
              url := "http://localhost/api/v1/machines/{id}/resume"
              method := "put"
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
    - title: Reset a virtual machine
      description: This endpoint reset a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/reset
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/reset' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/machines/{id}/reset");
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
              url := "http://localhost/api/v1/machines/{id}/reset"
              method := "put"
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
    - title: Pauses a virtual machine
      description: This endpoint pauses a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/pause
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/pause' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/machines/{id}/pause");
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
              url := "http://localhost/api/v1/machines/{id}/pause"
              method := "put"
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
    - title: Deletes a virtual machine
      description: This endpoint deletes a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/machines/{id}");
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
              url := "http://localhost/api/v1/machines/{id}"
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
    - title: Get the current state of a virtual machine
      description: This endpoint returns the current state of a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/status
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineStatusResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/status' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/machines/{id}/status");
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
              url := "http://localhost/api/v1/machines/{id}/status"
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
    - title: Configures a virtual machine
      description: This endpoint configures a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/set
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
        - name: configRequest
          required: false
          type: body
          value_type: object
          description: Machine Set Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineConfigResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/set' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/machines/{id}/set");
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
              url := "http://localhost/api/v1/machines/{id}/set"
              method := "put"
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
    - title: Clones a virtual machine
      description: This endpoint clones a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/clone
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
        - name: configRequest
          required: false
          type: body
          value_type: object
          description: Machine Clone Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineCloneCommandResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/clone' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/machines/{id}/clone");
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
              url := "http://localhost/api/v1/machines/{id}/clone"
              method := "put"
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
    - title: Executes a command on a virtual machine
      description: This endpoint executes a command on a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/execute
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
        - name: executeRequest
          required: false
          type: body
          value_type: object
          description: Machine Execute Command Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineExecuteCommandResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/execute' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/machines/{id}/execute");
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
              url := "http://localhost/api/v1/machines/{id}/execute"
              method := "put"
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
    - title: Uploads a file to a virtual machine
      description: This endpoint executes a command on a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/upload
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
        - name: executeRequest
          required: false
          type: body
          value_type: object
          description: Machine Upload file Command Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineUploadResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/upload' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/machines/{id}/upload");
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
              url := "http://localhost/api/v1/machines/{id}/upload"
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
    - title: Renames a virtual machine
      description: This endpoint Renames a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/rename
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
        - name: renameRequest
          required: false
          type: body
          value_type: object
          description: Machine Rename Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ParallelsVM
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/rename' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/machines/{id}/rename");
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
              url := "http://localhost/api/v1/machines/{id}/rename"
              method := "put"
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
    - title: Registers a virtual machine
      description: This endpoint registers a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/register
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
        - name: registerRequest
          required: false
          type: body
          value_type: object
          description: Machine Register Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ParallelsVM
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/register' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/machines/register");
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
              url := "http://localhost/api/v1/machines/register"
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
    - title: Unregister a virtual machine
      description: This endpoint unregister a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/unregister
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
        - name: unregisterRequest
          required: false
          type: body
          value_type: object
          description: Machine Unregister Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ApiCommonResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/unregister' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/machines/{id}/unregister");
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
              url := "http://localhost/api/v1/machines/{id}/unregister"
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
    - title: Creates a virtual machine
      description: This endpoint creates a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines
      method: post
      parameters:
        - name: createRequest
          required: false
          type: body
          value_type: object
          description: New Machine Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CreateVirtualMachineResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/machines");
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
              url := "http://localhost/api/v1/machines"
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
    - title: Creates a virtual machine asynchronously
      description: This endpoint creates a virtual machine in the background and returns a Job ID to track progress
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/async
      method: post
      parameters:
        - name: createRequest
          required: false
          type: body
          value_type: object
          description: New Machine Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "202"
          code_description: Accepted
          title: models.JobResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/async' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/machines/async");
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
              url := "http://localhost/api/v1/machines/async"
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
    - title: Creates a snapshot for a virtual machine
      description: This endpoint creates a snapshot for a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/snapshots
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
        - name: createRequest
          required: false
          type: body
          value_type: object
          description: Create Snapshot Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "202"
          code_description: Accepted
          title: models.ApiCommonResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/snapshots' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/machines/{id}/snapshots");
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
              url := "http://localhost/api/v1/machines/{id}/snapshots"
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
    - title: Deletes a snapshot of a virtual machine
      description: This endpoint deletes a snapshot of a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/snapshots/{snapshot_id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
        - name: snapshot_id
          required: true
          type: path
          value_type: string
          description: Snapshot ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/snapshots/{snapshot_id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/machines/{id}/snapshots/{snapshot_id}");
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
              url := "http://localhost/api/v1/machines/{id}/snapshots/{snapshot_id}"
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
    - title: Deletes all snapshots of a virtual machine
      description: This endpoint deletes all snapshots of a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/snapshots
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/snapshots' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/machines/{id}/snapshots");
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
              url := "http://localhost/api/v1/machines/{id}/snapshots"
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
    - title: Lists snapshots of a virtual machine
      description: This endpoint lists snapshots of a virtual machine
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/snapshots
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "202"
          code_description: Accepted
          title: models.ApiCommonResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/snapshots' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/machines/{id}/snapshots");
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
              url := "http://localhost/api/v1/machines/{id}/snapshots"
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
    - title: Reverts a virtual machine to a snapshot
      description: This endpoint reverts a virtual machine to a snapshot
      requires_authorization: true
      category: Machines
      category_path: machines
      path: /v1/machines/{id}/snapshots/{snapshot_id}/revert
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Machine ID
        - name: snapshot_id
          required: true
          type: path
          value_type: string
          description: Snapshot ID
        - name: revertRequest
          required: false
          type: body
          value_type: object
          description: Revert Snapshot Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "202"
          code_description: Accepted
          title: models.ApiCommonResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/machines/{id}/snapshots/{snapshot_id}/revert' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/machines/{id}/snapshots/{snapshot_id}/revert");
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
              url := "http://localhost/api/v1/machines/{id}/snapshots/{snapshot_id}/revert"
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
    - title: Gets all hosts from the orchestrator
      description: This endpoint returns all hosts from the orchestrator
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.OrchestratorHostResponse'
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/hosts");
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
              url := "http://localhost/api/v1/orchestrator/hosts"
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
    - title: Gets a host from the orchestrator
      description: This endpoint returns a host from the orchestrator
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.OrchestratorHostResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/hosts/{id}");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}"
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
    - title: Gets a host hardware info from the orchestrator
      description: This endpoint returns a host hardware info from the orchestrator
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/hardware
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.SystemUsageResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/hardware' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/hosts/{id}/hardware");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/hardware"
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
    - title: Register a Host in the orchestrator
      description: This endpoint register a host in the orchestrator
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts
      method: post
      parameters:
        - name: hostRequest
          required: false
          type: body
          value_type: object
          description: Host Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.OrchestratorHostResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts");
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
              url := "http://localhost/api/v1/orchestrator/hosts"
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
    - title: Unregister a host from the orchestrator
      description: This endpoint unregister a host from the orchestrator
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/orchestrator/hosts/{id}");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}"
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
    - title: Enable a host in the orchestrator
      description: This endpoint will enable an existing host in the orchestrator
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/enable
      method: put
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.OrchestratorHostResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/enable' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/enable");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/enable"
              method := "put"
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
    - title: Disable a host in the orchestrator
      description: This endpoint will disable an existing host in the orchestrator
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/disable
      method: put
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.OrchestratorHostResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/disable' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/disable");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/disable"
              method := "put"
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
    - title: Update a Host in the orchestrator
      description: This endpoint updates a host in the orchestrator
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts
      method: put
      parameters:
        - name: hostRequest
          required: false
          type: body
          value_type: object
          description: Host Update Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.OrchestratorHostResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts");
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
              url := "http://localhost/api/v1/orchestrator/hosts"
              method := "put"
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
    - title: Get orchestrator resource overview
      description: This endpoint returns orchestrator resource overview
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/overview/resources
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.HostResourceOverviewResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/overview/resources' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/overview/resources");
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
              url := "http://localhost/api/v1/orchestrator/overview/resources"
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
    - title: Get orchestrator host resources
      description: This endpoint returns orchestrator host resources
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/overview/{id}/resources
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.HostResourceOverviewResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/overview/{id}/resources' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/overview/{id}/resources");
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
              url := "http://localhost/api/v1/orchestrator/overview/{id}/resources"
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
    - title: Get orchestrator Virtual Machines
      description: This endpoint returns orchestrator Virtual Machines
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.ParallelsVM'
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/machines");
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
              url := "http://localhost/api/v1/orchestrator/machines"
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
    - title: Get orchestrator Virtual Machine
      description: This endpoint returns orchestrator Virtual Machine by its ID
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ParallelsVM
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/machines/{id}");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}"
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
    - title: Deletes orchestrator virtual machine
      description: This endpoint deletes orchestrator virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
        - name: force
          required: false
          type: query
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/orchestrator/machines/{id}");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}"
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
    - title: Get orchestrator virtual machine status
      description: This endpoint returns orchestrator virtual machine status
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/status
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ParallelsVM
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/status' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/machines/{id}/status");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/status"
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
    - title: Renames orchestrator virtual machine
      description: This endpoint renames orchestrator virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/rename
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ParallelsVM
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/rename' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{id}/rename");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/rename"
              method := "put"
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
    - title: Configures orchestrator virtual machine
      description: This endpoint configures orchestrator virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/set
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineConfigResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/set' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{id}/set");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/set"
              method := "put"
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
    - title: Starts orchestrator virtual machine
      description: This endpoint starts orchestrator virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/start
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/start' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{id}/start");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/start"
              method := "put"
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
    - title: Stops orchestrator virtual machine
      description: This endpoint sops orchestrator virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/stop
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
        - name: force
          required: false
          type: query
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/stop' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{id}/stop");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/stop"
              method := "put"
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
    - title: Restarts orchestrator virtual machine
      description: This endpoint restarts orchestrator virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/restart
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/restart' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{id}/restart");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/restart"
              method := "put"
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
    - title: Suspends orchestrator virtual machine
      description: This endpoint suspends orchestrator virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/suspend
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/suspend' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{id}/suspend");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/suspend"
              method := "put"
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
    - title: Resumes orchestrator virtual machine
      description: This endpoint resumes orchestrator virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/resume
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/resume' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{id}/resume");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/resume"
              method := "put"
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
    - title: Resets orchestrator virtual machine
      description: This endpoint resets orchestrator virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/reset
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/reset' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{id}/reset");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/reset"
              method := "put"
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
    - title: Pauses orchestrator virtual machine
      description: This endpoint pauses orchestrator virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/pause
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/pause' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{id}/pause");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/pause"
              method := "put"
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
    - title: Clones orchestrator virtual machine
      description: This endpoint clones orchestrator virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/clone
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
        - name: configRequest
          required: false
          type: body
          value_type: object
          description: Machine Clone Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineCloneCommandResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/clone' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{id}/clone");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/clone"
              method := "put"
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
    - title: Executes a command in a orchestrator virtual machine
      description: This endpoint executes a command in a orchestrator virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/execute
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineConfigResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/execute' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{id}/execute");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/execute"
              method := "put"
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
    - title: Get orchestrator host virtual machines
      description: This endpoint returns orchestrator host virtual machines
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.ParallelsVM'
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/hosts/{id}/machines");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines"
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
    - title: Get orchestrator host virtual machine
      description: This endpoint returns orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ParallelsVM
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}"
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
    - title: Deletes orchestrator host virtual machine
      description: This endpoint deletes orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
        - name: force
          required: false
          type: query
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}"
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
    - title: Get orchestrator host virtual machine status
      description: This endpoint returns orchestrator host virtual machine status
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/status
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ParallelsVM
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/status' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/status");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/status"
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
    - title: Renames orchestrator host virtual machine
      description: This endpoint renames orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/rename
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ParallelsVM
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/rename' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/rename");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/rename"
              method := "put"
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
    - title: Configures orchestrator host virtual machine
      description: This endpoint configures orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/set
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineConfigResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/set' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/set");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/set"
              method := "put"
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
    - title: Starts orchestrator host virtual machine
      description: This endpoint starts orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/start
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/start' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/start");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/start"
              method := "put"
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
    - title: Stops orchestrator host virtual machine
      description: This endpoint stops orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/stop
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
        - name: force
          required: false
          type: query
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/stop' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/stop");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/stop"
              method := "put"
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
    - title: Restarts orchestrator host virtual machine
      description: This endpoint restarts orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/restart
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/restart' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/restart");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/restart"
              method := "put"
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
    - title: Suspends orchestrator host virtual machine
      description: This endpoint suspends orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/suspend
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/suspend' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/suspend");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/suspend"
              method := "put"
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
    - title: Resumes orchestrator host virtual machine
      description: This endpoint resumes orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/resume
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/resume' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/resume");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/resume"
              method := "put"
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
    - title: Resets orchestrator host virtual machine
      description: This endpoint resets orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/reset
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/reset' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/reset");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/reset"
              method := "put"
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
    - title: Pauses orchestrator host virtual machine
      description: This endpoint pauses orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/pause
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineOperationResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/pause' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/pause");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/pause"
              method := "put"
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
    - title: Clones orchestrator host virtual machine
      description: This endpoint clones orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/clone
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
        - name: configRequest
          required: false
          type: body
          value_type: object
          description: Machine Clone Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineCloneCommandResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/clone' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/clone");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/clone"
              method := "put"
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
    - title: Executes a command in a orchestrator host virtual machine
      description: This endpoint executes a command in a orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/execute
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.VirtualMachineConfigResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/execute' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/execute");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/execute"
              method := "put"
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
    - title: Lists snapshots of orchestrator host virtual machine
      description: This endpoint lists snapshots of orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ListVMSnapshotResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots"
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
    - title: Creates a snapshot for orchestrator host virtual machine
      description: This endpoint creates a snapshot for orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
        - name: createRequest
          required: false
          type: body
          value_type: object
          description: Create Snapshot Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "202"
          code_description: Accepted
          title: models.CreateVMSnapshotResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots"
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
    - title: Deletes all snapshots of orchestrator host virtual machine
      description: This endpoint deletes all snapshots of orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots"
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
    - title: Deletes a snapshot of orchestrator host virtual machine
      description: This endpoint deletes a snapshot of orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
        - name: snapshot_id
          required: true
          type: path
          value_type: string
          description: Snapshot ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}"
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
    - title: Reverts orchestrator host virtual machine to a snapshot
      description: This endpoint reverts orchestrator host virtual machine to a snapshot
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}/revert
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
        - name: snapshot_id
          required: true
          type: path
          value_type: string
          description: Snapshot ID
        - name: revertRequest
          required: false
          type: body
          value_type: object
          description: Revert Snapshot Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "202"
          code_description: Accepted
          title: models.ApiCommonResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}/revert' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}/revert");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/snapshots/{snapshot_id}/revert"
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
    - title: Lists snapshots of an orchestrator virtual machine
      description: This endpoint lists snapshots of an orchestrator virtual machine (host resolved automatically)
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/snapshots
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ListVMSnapshotResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/snapshots' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/machines/{id}/snapshots");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/snapshots"
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
    - title: Creates a snapshot for an orchestrator virtual machine
      description: This endpoint creates a snapshot for an orchestrator virtual machine (host resolved automatically)
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/snapshots
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
        - name: createRequest
          required: false
          type: body
          value_type: object
          description: Create Snapshot Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "202"
          code_description: Accepted
          title: models.CreateVMSnapshotResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/snapshots' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/machines/{id}/snapshots");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/snapshots"
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
    - title: Deletes all snapshots of an orchestrator virtual machine
      description: This endpoint deletes all snapshots of an orchestrator virtual machine (host resolved automatically)
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/snapshots
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/snapshots' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/orchestrator/machines/{id}/snapshots");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/snapshots"
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
    - title: Deletes a snapshot of an orchestrator virtual machine
      description: This endpoint deletes a snapshot of an orchestrator virtual machine (host resolved automatically)
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/snapshots/{snapshot_id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
        - name: snapshot_id
          required: true
          type: path
          value_type: string
          description: Snapshot ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/snapshots/{snapshot_id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/orchestrator/machines/{id}/snapshots/{snapshot_id}");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/snapshots/{snapshot_id}"
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
    - title: Reverts an orchestrator virtual machine to a snapshot
      description: This endpoint reverts an orchestrator virtual machine to a snapshot (host resolved automatically)
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{id}/snapshots/{snapshot_id}/revert
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
        - name: snapshot_id
          required: true
          type: path
          value_type: string
          description: Snapshot ID
        - name: revertRequest
          required: false
          type: body
          value_type: object
          description: Revert Snapshot Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "202"
          code_description: Accepted
          title: models.ApiCommonResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{id}/snapshots/{snapshot_id}/revert' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/machines/{id}/snapshots/{snapshot_id}/revert");
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
              url := "http://localhost/api/v1/orchestrator/machines/{id}/snapshots/{snapshot_id}/revert"
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
    - title: Register a virtual machine in a orchestrator host
      description: This endpoint registers a virtual machine in a orchestrator host
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/register
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: request
          required: false
          type: body
          value_type: object
          description: Register Virtual Machine Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ParallelsVM
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/register' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/register");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/register"
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
    - title: Unregister a virtual machine in a orchestrator host
      description: This endpoint unregister a virtual machine in a orchestrator host
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/{vmId}/unregister
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: vmId
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
        - name: request
          required: false
          type: body
          value_type: object
          description: Register Virtual Machine Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ParallelsVM
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/unregister' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/unregister");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/unregister"
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
    - title: Creates a orchestrator host virtual machine
      description: This endpoint creates a orchestrator host virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: request
          required: false
          type: body
          value_type: object
          description: Create Virtual Machine Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CreateVirtualMachineResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/machines");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines"
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
    - title: Creates a virtual machine in one of the hosts for the orchestrator
      description: This endpoint creates a virtual machine in one of the hosts for the orchestrator
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines
      method: post
      parameters:
        - name: request
          required: false
          type: body
          value_type: object
          description: Create Virtual Machine Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.CreateVirtualMachineResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/machines");
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
              url := "http://localhost/api/v1/orchestrator/machines"
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
    - title: Gets orchestrator host reverse proxy configuration
      description: This endpoint returns orchestrator host reverse proxy configuration
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/reverse-proxy
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ReverseProxy
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy"
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
    - title: Gets orchestrator host reverse proxy hosts
      description: This endpoint returns orchestrator host reverse proxy hosts
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.ReverseProxyHost'
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts"
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
    - title: Gets orchestrator host reverse proxy hosts
      description: This endpoint returns orchestrator host reverse proxy hosts
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}"
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
    - title: Creates a orchestrator host reverse proxy host
      description: This endpoint creates a orchestrator host reverse proxy host
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts
      method: post
      parameters:
        - name: request
          required: false
          type: body
          value_type: object
          description: Create Host Reverse Proxy Host Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts"
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
    - title: Updates an orchestrator host reverse proxy host
      description: This endpoint updates an orchestrator host reverse proxy host
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}
      method: put
      parameters:
        - name: request
          required: false
          type: body
          value_type: object
          description: Update Host Reverse Proxy Host Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}"
              method := "put"
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
    - title: Deletes an orchestrator host reverse proxy host
      description: This endpoint deletes an orchestrator host reverse proxy host
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: reverse_proxy_host_id
          required: true
          type: path
          value_type: string
          description: Reverse Proxy Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}"
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
    - title: Upserts an orchestrator host reverse proxy host http route
      description: This endpoint upserts an orchestrator host reverse proxy host http route
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes
      method: post
      parameters:
        - name: request
          required: false
          type: body
          value_type: object
          description: Upsert Host Reverse Proxy Host Http Routes Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes"
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
    - title: Deletes an orchestrator host reverse proxy host http route
      description: This endpoint deletes an orchestrator host reverse proxy host http route
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes/{route_id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: reverse_proxy_host_id
          required: true
          type: path
          value_type: string
          description: Reverse Proxy Host ID
        - name: route_id
          required: true
          type: path
          value_type: string
          description: Route ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes/{route_id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes/{route_id}");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes/{route_id}"
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
    - title: Update an orchestrator host reverse proxy host tcp route
      description: This endpoint updates an orchestrator host reverse proxy host tcp route
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/tcp_route
      method: post
      parameters:
        - name: request
          required: false
          type: body
          value_type: object
          description: Update Host Reverse Proxy Host tcp Routes Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/tcp_route' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/tcp_route");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/tcp_route"
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
    - title: Restarts orchestrator host reverse proxy
      description: This endpoint restarts orchestrator host reverse proxy
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/reverse-proxy/restart
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/restart' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/restart");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/restart"
              method := "put"
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
    - title: Enables orchestrator host reverse proxy
      description: This endpoint enables orchestrator host reverse proxy
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/reverse-proxy/enable
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/enable' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/enable");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/enable"
              method := "put"
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
    - title: Disables orchestrator host reverse proxy
      description: This endpoint disables orchestrator host reverse proxy
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/reverse-proxy/disable
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/disable' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/disable");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/disable"
              method := "put"
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
    - title: Gets orchestrator host catalog cache
      description: This endpoint returns orchestrator host catalog cache
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/catalog/cache
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/catalog/cache' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/hosts/{id}/catalog/cache");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/catalog/cache"
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
    - title: Deletes an orchestrator host cache items
      description: This endpoint deletes an orchestrator host cache items
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/catalog/cache
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/catalog/cache' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/orchestrator/hosts/{id}/catalog/cache");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/catalog/cache"
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
    - title: Deletes an orchestrator host cache item and all its children
      description: This endpoint deletes an orchestrator host cache item and all its children
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: catalog_id
          required: true
          type: path
          value_type: string
          description: Catalog ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}"
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
    - title: Deletes an orchestrator host cache item and all its children
      description: This endpoint deletes an orchestrator host cache item and all its children
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}/{catalog_version}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: catalog_id
          required: true
          type: path
          value_type: string
          description: Catalog ID
        - name: catalog_version
          required: true
          type: path
          value_type: string
          description: Catalog Version
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}/{catalog_version}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}/{catalog_version}");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/catalog/cache/{catalog_id}/{catalog_version}"
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
    - title: Gets the orchestrator host system logs from the disk
      description: This endpoint returns the orchestrator host system logs from the disk
      requires_authorization: true
      category: Config
      category_path: config
      path: /v1/orchestrator/hosts/{id}/logs
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/logs' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/hosts/{id}/logs");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/logs"
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
    - title: Streams the system logs via WebSocket
      description: This endpoint streams the system logs in real-time via WebSocket
      requires_authorization: true
      category: Config
      category_path: config
      path: /logs/stream
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/logs/stream' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/logs/stream");
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
              url := "http://localhost/api/logs/stream"
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
    - title: Create an enrollment token
      description: Generates a short-lived, single-use token that allows a freshly installed agent to register itself with the orchestrator without requiring a permanent credential.
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/enrollment-token
      method: post
      parameters:
        - name: request
          required: false
          type: body
          value_type: object
          description: Enrollment token request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "201"
          code_description: Created
          title: models.CreateEnrollmentTokenResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/enrollment-token' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/enrollment-token");
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
              url := "http://localhost/api/v1/orchestrator/enrollment-token"
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
    - title: Validate an enrollment token
      description: Public endpoint that checks whether an enrollment token is valid, unused, and not expired. Used by agents before starting the registration flow.
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/enrollment-token/{token}/validate
      method: get
      parameters:
        - name: token
          required: true
          type: path
          value_type: string
          description: Enrollment token value
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ValidateEnrollmentTokenResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/enrollment-token/{token}/validate' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/enrollment-token/{token}/validate");
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
              url := "http://localhost/api/v1/orchestrator/enrollment-token/{token}/validate"
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
    - title: Deploy and register an agent via SSH (synchronous)
      description: SSHes into a remote host, installs the devops agent, and registers it with this orchestrator. Blocks until the operation completes.
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/deploy
      method: post
      parameters:
        - name: request
          required: false
          type: body
          value_type: object
          description: Deploy request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "201"
          code_description: Created
          title: models.DeployOrchestratorHostResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/deploy' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/deploy");
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
              url := "http://localhost/api/v1/orchestrator/hosts/deploy"
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
    - title: Deploy and register an agent via SSH (asynchronous)
      description: SSHes into a remote host, installs the devops agent, and registers it with this orchestrator. Returns a job ID immediately; poll /jobs/{id} for status.
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/deploy/async
      method: post
      parameters:
        - name: request
          required: false
          type: body
          value_type: object
          description: Deploy request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "202"
          code_description: Accepted
          title: models.JobResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/deploy/async' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/deploy/async");
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
              url := "http://localhost/api/v1/orchestrator/hosts/deploy/async"
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
    - title: Creates a virtual machine in one of the orchestrator hosts asynchronously
      description: This endpoint creates a virtual machine in one of the orchestrator hosts in the background and returns a Job ID to track progress
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/async
      method: post
      parameters:
        - name: request
          required: false
          type: body
          value_type: object
          description: Create Virtual Machine Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "202"
          code_description: Accepted
          title: models.JobResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/async' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/machines/async");
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
              url := "http://localhost/api/v1/orchestrator/machines/async"
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
    - title: Creates a virtual machine in a specific orchestrator host asynchronously
      description: This endpoint creates a virtual machine in a specific orchestrator host in the background and returns a Job ID to track progress
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts/{id}/machines/async
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Host ID
        - name: request
          required: false
          type: body
          value_type: object
          description: Create Virtual Machine Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "202"
          code_description: Accepted
          title: models.JobResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/async' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/async");
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/async"
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
    - title: Gets all the packer templates
      description: This endpoint returns all the packer templates. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
      requires_authorization: true
      category: Packer Templates
      category_path: packer_templates
      path: /v1/templates/packer
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.PackerTemplateResponse'
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
        - code_block: "curl --location 'http://localhost/api/v1/templates/packer' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/templates/packer");
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
              url := "http://localhost/api/v1/templates/packer"
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
    - title: Gets a packer template
      description: This endpoint returns a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
      requires_authorization: true
      category: Packer Templates
      category_path: packer_templates
      path: /v1/templates/packer/{id}
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Packer Template ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.PackerTemplateResponse
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
        - code_block: "curl --location 'http://localhost/api/v1/templates/packer/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/templates/packer/{id}");
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
              url := "http://localhost/api/v1/templates/packer/{id}"
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
    - title: Creates a packer template
      description: This endpoint creates a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
      requires_authorization: true
      category: Packer Templates
      category_path: packer_templates
      path: '/v1/templates/packer '
      method: post
      parameters:
        - name: createPackerTemplateRequest
          required: false
          type: body
          value_type: object
          description: Create Packer Template Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.PackerTemplateResponse
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
        - code_block: "curl --location 'http://localhost/api/v1/templates/packer ' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/templates/packer ");
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
              url := "http://localhost/api/v1/templates/packer "
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
    - title: Updates a packer template
      description: This endpoint updates a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
      requires_authorization: true
      category: Packer Templates
      category_path: packer_templates
      path: '/v1/templates/packer/{id} '
      method: PUT
      parameters:
        - name: createPackerTemplateRequest
          required: false
          type: body
          value_type: object
          description: Update Packer Template Request
          body: '{ object }'
        - name: id
          required: true
          type: path
          value_type: string
          description: Packer Template ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.PackerTemplateResponse
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
        - code_block: "curl --location 'http://localhost/api/v1/templates/packer/{id} ' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/templates/packer/{id} ");
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
              url := "http://localhost/api/v1/templates/packer/{id} "
              method := "PUT"
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
    - title: Deletes a packer template
      description: This endpoint deletes a packer template. **DEPRECATED:** This endpoint will be deprecated in the future, please upgrade your calls to use the catalog service, see https://parallels.github.io/prl-devops-service/docs/devops/catalog/overview/
      requires_authorization: true
      category: Packer Templates
      category_path: packer_templates
      path: '/v1/templates/packer/{id} '
      method: DELETE
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Packer Template ID
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
        - code_block: "curl --location 'http://localhost/api/v1/templates/packer/{id} ' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/templates/packer/{id} ");
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
              url := "http://localhost/api/v1/templates/packer/{id} "
              method := "DELETE"
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
    - title: Gets reverse proxy configuration
      description: This endpoint returns the reverse proxy configuration
      requires_authorization: true
      category: ReverseProxy
      category_path: reverseproxy
      path: /v1/reverse-proxy
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ReverseProxy
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/reverse-proxy' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/reverse-proxy");
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
              url := "http://localhost/api/v1/reverse-proxy"
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
    - title: Gets all the reverse proxy hosts
      description: This endpoint returns all the reverse proxy hosts
      requires_authorization: true
      category: ReverseProxy
      category_path: reverseproxy
      path: /v1/reverse-proxy/hosts
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.ReverseProxyHost'
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/reverse-proxy/hosts' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/reverse-proxy/hosts");
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
              url := "http://localhost/api/v1/reverse-proxy/hosts"
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
    - title: Gets all the reverse proxy host
      description: This endpoint returns a reverse proxy host
      requires_authorization: true
      category: ReverseProxy
      category_path: reverseproxy
      path: '/v1/reverse-proxy/hosts/{id} '
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Reverse Proxy Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/reverse-proxy/hosts/{id} ' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/reverse-proxy/hosts/{id} ");
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
              url := "http://localhost/api/v1/reverse-proxy/hosts/{id} "
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
    - title: Creates a reverse proxy host
      description: This endpoint creates a reverse proxy host
      requires_authorization: true
      category: ReverseProxy
      category_path: reverseproxy
      path: /v1/reverse-proxy/hosts
      method: post
      parameters:
        - name: reverse_proxy_create_request
          required: false
          type: body
          value_type: object
          description: Reverse Host Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/reverse-proxy/hosts' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/reverse-proxy/hosts");
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
              url := "http://localhost/api/v1/reverse-proxy/hosts"
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
    - title: Updates a reverse proxy host
      description: This endpoint creates a reverse proxy host
      requires_authorization: true
      category: ReverseProxy
      category_path: reverseproxy
      path: /v1/reverse-proxy/hosts/{id}
      method: put
      parameters:
        - name: reverse_proxy_update_request
          required: false
          type: body
          value_type: object
          description: Reverse Host Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/reverse-proxy/hosts/{id}' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/reverse-proxy/hosts/{id}");
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
              url := "http://localhost/api/v1/reverse-proxy/hosts/{id}"
              method := "put"
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
    - title: Delete a a reverse proxy host
      description: This endpoint Deletes a reverse proxy host
      requires_authorization: true
      category: ReverseProxy
      category_path: reverseproxy
      path: /v1/reverse-proxy/hosts/{id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Reverse Proxy Host ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/reverse-proxy/hosts/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/reverse-proxy/hosts/{id}");
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
              url := "http://localhost/api/v1/reverse-proxy/hosts/{id}"
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
    - title: Creates or updates a reverse proxy host HTTP route
      description: This endpoint creates or updates a reverse proxy host HTTP route
      requires_authorization: true
      category: ReverseProxy
      category_path: reverseproxy
      path: /v1/reverse-proxy/hosts/{id}/http_route
      method: post
      parameters:
        - name: reverse_proxy_http_route_request
          required: false
          type: body
          value_type: object
          description: Reverse Host Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/reverse-proxy/hosts/{id}/http_route' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/reverse-proxy/hosts/{id}/http_route");
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
              url := "http://localhost/api/v1/reverse-proxy/hosts/{id}/http_route"
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
    - title: Delete a a reverse proxy host HTTP route
      description: This endpoint Deletes a reverse proxy host HTTP route
      requires_authorization: true
      category: ReverseProxy
      category_path: reverseproxy
      path: /v1/reverse-proxy/hosts/{id}/http_routes/{http_route_id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Reverse Proxy Host ID
        - name: http_route_id
          required: true
          type: path
          value_type: string
          description: Reverse Proxy Host HTTP Route ID
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/reverse-proxy/hosts/{id}/http_routes/{http_route_id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/reverse-proxy/hosts/{id}/http_routes/{http_route_id}");
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
              url := "http://localhost/api/v1/reverse-proxy/hosts/{id}/http_routes/{http_route_id}"
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
    - title: Updates the order of a reverse proxy host HTTP route
      description: This endpoint reorders HTTP routes for a reverse proxy host
      requires_authorization: true
      category: ReverseProxy
      category_path: reverseproxy
      path: /v1/reverse-proxy/hosts/{id}/http_routes/order
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Reverse Proxy Host ID
        - name: reverse_proxy_http_route_reorder_request
          required: false
          type: body
          value_type: object
          description: Reverse Proxy Host HTTP Route Reorder Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/reverse-proxy/hosts/{id}/http_routes/order' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/reverse-proxy/hosts/{id}/http_routes/order");
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
              url := "http://localhost/api/v1/reverse-proxy/hosts/{id}/http_routes/order"
              method := "put"
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
    - title: Updates a reverse proxy host TCP route
      description: This endpoint updates a reverse proxy host TCP route
      requires_authorization: true
      category: ReverseProxy
      category_path: reverseproxy
      path: /v1/reverse-proxy/hosts/{id}/http_routes
      method: post
      parameters:
        - name: reverse_proxy_tcp_route_request
          required: false
          type: body
          value_type: object
          description: Reverse Host Request
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/reverse-proxy/hosts/{id}/http_routes' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/reverse-proxy/hosts/{id}/http_routes");
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
              url := "http://localhost/api/v1/reverse-proxy/hosts/{id}/http_routes"
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
    - title: Restarts the reverse proxy
      description: This endpoint will restart the reverse proxy
      requires_authorization: true
      category: ReverseProxy
      category_path: reverseproxy
      path: /v1/reverse-proxy/restart
      method: put
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/reverse-proxy/restart' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/reverse-proxy/restart");
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
              url := "http://localhost/api/v1/reverse-proxy/restart"
              method := "put"
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
    - title: Enable the reverse proxy
      description: This endpoint will enable the reverse proxy
      requires_authorization: true
      category: ReverseProxy
      category_path: reverseproxy
      path: /v1/reverse-proxy/enable
      method: put
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/reverse-proxy/enable' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/reverse-proxy/enable");
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
              url := "http://localhost/api/v1/reverse-proxy/enable"
              method := "put"
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
    - title: Disable the reverse proxy
      description: This endpoint will disable the reverse proxy
      requires_authorization: true
      category: ReverseProxy
      category_path: reverseproxy
      path: /v1/reverse-proxy/disable
      method: put
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/reverse-proxy/disable' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/reverse-proxy/disable");
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
              url := "http://localhost/api/v1/reverse-proxy/disable"
              method := "put"
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
    - title: Execute SSH Command
      description: Executes a command on a remote host via SSH
      requires_authorization: true
      category: SSH
      category_path: ssh
      path: /v1/ssh/execute
      method: post
      parameters:
        - name: sshRequest
          required: false
          type: body
          value_type: object
          description: Body
          body: |-
            {
              "command": "string",
              "enable_insecure_key": "bool",
              "host": "string",
              "key": "string",
              "password": "string",
              "port": "int",
              "username": "string"
            }
      response_blocks:
        - code_block: |-
            {
              "error": "string",
              "output": "string"
            }
          code: "200"
          code_description: OK
          title: SshExecutionResponse
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
        - code_block: '{ object }'
          code: "500"
          code_description: Internal Server Error
          title: models.ApiErrorDiagnosticsResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/ssh/execute' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"command\": \"string\",\n  \"enable_insecure_key\": \"bool\",\n  \"host\": \"string\",\n  \"key\": \"string\",\n  \"password\": \"string\",\n  \"port\": \"int\",\n  \"username\": \"string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/ssh/execute");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "command": "string",
              "enable_insecure_key": "bool",
              "host": "string",
              "key": "string",
              "password": "string",
              "port": "int",
              "username": "string"
            }");
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
              url := "http://localhost/api/v1/ssh/execute"
              method := "post"
              payload := strings.NewReader(`{
              "command": "string",
              "enable_insecure_key": "bool",
              "host": "string",
              "key": "string",
              "password": "string",
              "port": "int",
              "username": "string"
            }`)
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
    - title: Gets all user configs
      description: This endpoint returns all configuration entries for the authenticated user
      requires_authorization: true
      category: User Configs
      category_path: user_configs
      path: /v1/user/configs
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.UserConfigResponse'
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.ApiErrorDiagnosticsResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/user/configs' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/user/configs");
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
              url := "http://localhost/api/v1/user/configs"
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
    - title: Gets a user config by id or slug
      description: This endpoint returns a single configuration entry for the authenticated user
      requires_authorization: true
      category: User Configs
      category_path: user_configs
      path: /v1/user/configs/{id}
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Config ID or Slug
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.UserConfigResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "404"
          code_description: Not Found
          title: models.ApiErrorDiagnosticsResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/user/configs/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/user/configs/{id}");
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
              url := "http://localhost/api/v1/user/configs/{id}"
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
    - title: Creates a user config
      description: This endpoint creates a configuration entry for the authenticated user
      requires_authorization: true
      category: User Configs
      category_path: user_configs
      path: /v1/user/configs
      method: post
      parameters:
        - name: userConfig
          required: false
          type: body
          value_type: object
          description: Body
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "201"
          code_description: Created
          title: models.UserConfigResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.ApiErrorDiagnosticsResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/user/configs' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/user/configs");
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
              url := "http://localhost/api/v1/user/configs"
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
    - title: Updates a user config
      description: This endpoint updates a configuration entry for the authenticated user
      requires_authorization: true
      category: User Configs
      category_path: user_configs
      path: /v1/user/configs/{id}
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Config ID or Slug
        - name: userConfig
          required: false
          type: body
          value_type: object
          description: Body
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.UserConfigResponse
          language: json
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "404"
          code_description: Not Found
          title: models.ApiErrorDiagnosticsResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/user/configs/{id}' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/user/configs/{id}");
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
              url := "http://localhost/api/v1/user/configs/{id}"
              method := "put"
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
    - title: Deletes a user config
      description: This endpoint deletes a configuration entry for the authenticated user
      requires_authorization: true
      category: User Configs
      category_path: user_configs
      path: /v1/user/configs/{id}
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Config ID or Slug
      response_blocks:
        - code_block: '{ object }'
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.ApiErrorDiagnosticsResponse
          language: json
        - code_block: '{ object }'
          code: "404"
          code_description: Not Found
          title: models.ApiErrorDiagnosticsResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/user/configs/{id}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/user/configs/{id}");
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
              url := "http://localhost/api/v1/user/configs/{id}"
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
    - title: Gets all the users
      description: This endpoint returns all the users
      requires_authorization: true
      category: Users
      category_path: users
      path: '/v1/auth/users '
      method: get
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: '[]models.ApiUser'
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/users ' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/auth/users ");
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
              url := "http://localhost/api/v1/auth/users "
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
    - title: Gets a user
      description: This endpoint returns a user
      requires_authorization: true
      category: Users
      category_path: users
      path: '/v1/auth/users/{id} '
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: User ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
          title: models.ApiUser
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/users/{id} ' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/auth/users/{id} ");
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
              url := "http://localhost/api/v1/auth/users/{id} "
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
    - title: Creates a user
      description: This endpoint creates a user
      requires_authorization: true
      category: Users
      category_path: users
      path: '/v1/auth/users '
      method: post
      parameters:
        - name: body
          required: false
          type: body
          value_type: object
          description: User
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "201"
          code_description: Created
          title: models.ApiUser
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/users ' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/auth/users ");
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
              url := "http://localhost/api/v1/auth/users "
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
    - title: Deletes a user
      description: This endpoint deletes a user
      requires_authorization: true
      category: Users
      category_path: users
      path: '/v1/auth/users/{id} '
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: User ID
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/users/{id} ' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/auth/users/{id} ");
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
              url := "http://localhost/api/v1/auth/users/{id} "
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
    - title: Update a user
      description: This endpoint updates a user
      requires_authorization: true
      category: Users
      category_path: users
      path: '/v1/auth/users/{id} '
      method: put
      parameters:
        - name: body
          required: false
          type: body
          value_type: object
          description: User
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "202"
          code_description: Accepted
          title: models.ApiUser
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/users/{id} ' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/auth/users/{id} ");
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
              url := "http://localhost/api/v1/auth/users/{id} "
              method := "put"
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
    - title: Gets all the roles for a user
      description: This endpoint returns all the roles for a user
      requires_authorization: true
      category: Users
      category_path: users
      path: '/v1/auth/users/{id}/roles '
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: User ID
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/users/{id}/roles ' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/auth/users/{id}/roles ");
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
              url := "http://localhost/api/v1/auth/users/{id}/roles "
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
    - title: Adds a role to a user
      description: This endpoint adds a role to a user
      requires_authorization: true
      category: Users
      category_path: users
      path: '/v1/auth/users/{id}/roles '
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: User ID
        - name: body
          required: false
          type: body
          value_type: object
          description: Role Name
          body: '{ object }'
      response_blocks:
        - code_block: '{ object }'
          code: "201"
          code_description: Created
          title: models.RoleRequest
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/users/{id}/roles ' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/auth/users/{id}/roles ");
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
              url := "http://localhost/api/v1/auth/users/{id}/roles "
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
    - title: Removes a role from a user
      description: This endpoint removes a role from a user
      requires_authorization: true
      category: Users
      category_path: users
      path: '/v1/auth/users/{id}/roles/{role_id} '
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: User ID
        - name: role_id
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/users/{id}/roles/{role_id} ' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/auth/users/{id}/roles/{role_id} ");
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
              url := "http://localhost/api/v1/auth/users/{id}/roles/{role_id} "
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
    - title: Gets all the claims for a user
      description: This endpoint returns all the claims for a user
      requires_authorization: true
      category: Users
      category_path: users
      path: '/v1/auth/users/{id}/claims '
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: User ID
      response_blocks:
        - code_block: '{ object }'
          code: "200"
          code_description: OK
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/users/{id}/claims ' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/auth/users/{id}/claims ");
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
              url := "http://localhost/api/v1/auth/users/{id}/claims "
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
    - title: Adds a claim to a user
      description: This endpoint adds a claim to a user
      requires_authorization: true
      category: Users
      category_path: users
      path: '/v1/auth/users/{id}/claims '
      method: post
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: User ID
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
          title: models.ClaimRequest
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/users/{id}/claims ' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{ object }'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/auth/users/{id}/claims ");
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
              url := "http://localhost/api/v1/auth/users/{id}/claims "
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
    - title: Removes a claim from a user
      description: This endpoint removes a claim from a user
      requires_authorization: true
      category: Users
      category_path: users
      path: '/v1/auth/users/{id}/claims/{claim_id} '
      method: delete
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: User ID
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
        - code_block: "curl --location 'http://localhost/api/v1/auth/users/{id}/claims/{claim_id} ' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/auth/users/{id}/claims/{claim_id} ");
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
              url := "http://localhost/api/v1/auth/users/{id}/claims/{claim_id} "
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
    - title: Gets the API Health Probe
      description: This endpoint returns the API Health Probe
      requires_authorization: true
      category: Config
      category_path: config
      path: /health/probe
      method: get
      response_blocks:
        - code_block: |-
            {
              "additionalProp1": "string",
              "additionalProp2": "string",
              "additionalProp3": "string"
            }
          code: "200"
          code_description: OK
          title: map[string]string
          language: json
        - code_block: '{ object }'
          code: "402"
          code_description: Payment Required
          title: models.ApiErrorResponse
          language: json
        - code_block: '{ object }'
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/health/probe' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/health/probe");
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
              url := "http://localhost/api/health/probe"
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

---
# API Documentation

This document describes the REST API for the service.


