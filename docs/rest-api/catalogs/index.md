---
layout: api
title: Catalogs
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
        - anchor: /v1/catalog/pull-put
          method: put
          path: /v1/catalog/pull
          description: This endpoint pulls a remote catalog manifest
          title: Pull a remote catalog manifest
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
        - anchor: /v1/catalog/cache-get
          method: get
          path: /v1/catalog/cache
          description: This endpoint returns all the remote catalog cache if any
          title: Gets catalog cache
        - anchor: /v1/catalog/cache-delete
          method: delete
          path: /v1/catalog/cache
          description: This endpoint returns all the remote catalog cache if any
          title: Deletes all catalog cache
        - anchor: /v1/catalog/cache/{catalogId}-delete
          method: delete
          path: /v1/catalog/cache/{catalogId}
          description: This endpoint returns all the remote catalog cache if any and all its versions
          title: Deletes catalog cache item and all its versions
        - anchor: /v1/catalog/cache/{catalogId}/{version}-delete
          method: delete
          path: /v1/catalog/cache/{catalogId}/{version}
          description: This endpoint deletes a version of a cache ite,
          title: Deletes catalog cache version item
    - name: Claims
      path: claims
      endpoints:
        - anchor: /v1/auth/claims-get
          method: get
          path: /v1/auth/claims
          description: This endpoint returns all the claims
          title: Gets all the claims
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
        - anchor: /v1/machines/{id}/start-get
          method: get
          path: /v1/machines/{id}/start
          description: This endpoint starts a virtual machine
          title: Starts a virtual machine
        - anchor: /v1/machines/{id}/stop-get
          method: get
          path: /v1/machines/{id}/stop
          description: This endpoint stops a virtual machine
          title: Stops a virtual machine
        - anchor: /v1/machines/{id}/restart-get
          method: get
          path: /v1/machines/{id}/restart
          description: This endpoint restarts a virtual machine
          title: Restarts a virtual machine
        - anchor: /v1/machines/{id}/suspend-get
          method: get
          path: /v1/machines/{id}/suspend
          description: This endpoint suspends a virtual machine
          title: Suspends a virtual machine
        - anchor: /v1/machines/{id}/resume-get
          method: get
          path: /v1/machines/{id}/resume
          description: This endpoint resumes a virtual machine
          title: Resumes a virtual machine
        - anchor: /v1/machines/{id}/reset-get
          method: get
          path: /v1/machines/{id}/reset
          description: This endpoint reset a virtual machine
          title: Reset a virtual machine
        - anchor: /v1/machines/{id}/pause-get
          method: get
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
        - anchor: /v1/orchestrator/machines/{vmId}/status-get
          method: get
          path: /v1/orchestrator/machines/{vmId}/status
          description: This endpoint returns orchestrator virtual machine status
          title: Get orchestrator virtual machine status
        - anchor: /v1/orchestrator/machines/{id}/rename-put
          method: put
          path: /v1/orchestrator/machines/{id}/rename
          description: This endpoint renames orchestrator virtual machine
          title: Renames orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{vmId}/set-put
          method: put
          path: /v1/orchestrator/machines/{vmId}/set
          description: This endpoint configures orchestrator virtual machine
          title: Configures orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{vmId}/start-put
          method: put
          path: /v1/orchestrator/machines/{vmId}/start
          description: This endpoint starts orchestrator virtual machine
          title: Starts orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{vmId}/stop-put
          method: put
          path: /v1/orchestrator/machines/{vmId}/stop
          description: This endpoint sops orchestrator virtual machine
          title: Stops orchestrator virtual machine
        - anchor: /v1/orchestrator/machines/{vmId}/execute-put
          method: put
          path: /v1/orchestrator/machines/{vmId}/execute
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
          description: This endpoint starts orchestrator host virtual machine
          title: Starts orchestrator host virtual machine
        - anchor: /v1/orchestrator/hosts/{id}/machines/{vmId}/execute-put
          method: put
          path: /v1/orchestrator/hosts/{id}/machines/{vmId}/execute
          description: This endpoint executes a command in a orchestrator host virtual machine
          title: Executes a command in a orchestrator host virtual machine
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
    - name: Packer Templates
      path: packer_templates
      endpoints:
        - anchor: /v1/templates/packer-get
          method: get
          path: /v1/templates/packer
          description: This endpoint returns all the packer templates
          title: Gets all the packer templates
        - anchor: /v1/templates/packer/{id}-get
          method: get
          path: /v1/templates/packer/{id}
          description: This endpoint returns a packer template
          title: Gets a packer template
        - anchor: /v1/templates/packer -post
          method: post
          path: '/v1/templates/packer '
          description: This endpoint creates a packer template
          title: Creates a packer template
        - anchor: /v1/templates/packer/{id} -PUT
          method: PUT
          path: '/v1/templates/packer/{id} '
          description: This endpoint updates a packer template
          title: Updates a packer template
        - anchor: /v1/templates/packer/{id} -DELETE
          method: DELETE
          path: '/v1/templates/packer/{id} '
          description: This endpoint deletes a packer template
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
          description: This endpoint returns a role
          title: Gets a role
        - anchor: /v1/auth/roles/{id} -delete
          method: delete
          path: '/v1/auth/roles/{id} '
          description: This endpoint deletes a role
          title: Delete a role
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
    - title: Gets all the remote catalogs
      description: This endpoint returns all the remote catalogs
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog
      method: get
      response_blocks:
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: '[]map[string][]models.CatalogManifest'
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
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
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: '[]models.CatalogManifest'
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
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
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
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
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
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
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
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
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
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
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
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
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
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
          body: |-
            {
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"connection\": \"string\",\n  \"required_claims\": \"[]string\",\n  \"required_roles\": \"[]string\",\n  \"tags\": \"[]string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Patch, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims"
              method := "patch"
              payload := strings.NewReader(`{
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
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
          body: |-
            {
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"connection\": \"string\",\n  \"required_claims\": \"[]string\",\n  \"required_roles\": \"[]string\",\n  \"tags\": \"[]string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims"
              method := "delete"
              payload := strings.NewReader(`{
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
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
          body: |-
            {
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/roles' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"connection\": \"string\",\n  \"required_claims\": \"[]string\",\n  \"required_roles\": \"[]string\",\n  \"tags\": \"[]string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Patch, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/roles");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/roles"
              method := "patch"
              payload := strings.NewReader(`{
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
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
          body: |-
            {
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/roles' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"connection\": \"string\",\n  \"required_claims\": \"[]string\",\n  \"required_roles\": \"[]string\",\n  \"tags\": \"[]string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/roles");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/roles"
              method := "delete"
              payload := strings.NewReader(`{
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
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
          body: |-
            {
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/tags' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"connection\": \"string\",\n  \"required_claims\": \"[]string\",\n  \"required_roles\": \"[]string\",\n  \"tags\": \"[]string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Patch, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/tags");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/tags"
              method := "patch"
              payload := strings.NewReader(`{
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
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
          body: |-
            {
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/tags' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"connection\": \"string\",\n  \"required_claims\": \"[]string\",\n  \"required_roles\": \"[]string\",\n  \"tags\": \"[]string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/tags");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/tags"
              method := "delete"
              payload := strings.NewReader(`{
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
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
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
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
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
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
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
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
          body: |-
            {
              "architecture": "string",
              "catalog_id": "string",
              "compress_pack": "bool",
              "connection": "string",
              "description": "string",
              "local_path": "string",
              "minimum_requirements": {
                "cpu": "int",
                "disk": "int",
                "memory": "int"
              },
              "pack_size": "int64",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string",
              "uuid": "string",
              "version": "string"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/push' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"architecture\": \"string\",\n  \"catalog_id\": \"string\",\n  \"compress_pack\": \"bool\",\n  \"connection\": \"string\",\n  \"description\": \"string\",\n  \"local_path\": \"string\",\n  \"minimum_requirements\": {\n    \"cpu\": \"int\",\n    \"disk\": \"int\",\n    \"memory\": \"int\"\n  },\n  \"pack_size\": \"int64\",\n  \"required_claims\": \"[]string\",\n  \"required_roles\": \"[]string\",\n  \"tags\": \"[]string\",\n  \"uuid\": \"string\",\n  \"version\": \"string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/catalog/push");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "architecture": "string",
              "catalog_id": "string",
              "compress_pack": "bool",
              "connection": "string",
              "description": "string",
              "local_path": "string",
              "minimum_requirements": {
                "cpu": "int",
                "disk": "int",
                "memory": "int"
              },
              "pack_size": "int64",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string",
              "uuid": "string",
              "version": "string"
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
              url := "http://localhost/api/v1/catalog/push"
              method := "post"
              payload := strings.NewReader(`{
              "architecture": "string",
              "catalog_id": "string",
              "compress_pack": "bool",
              "connection": "string",
              "description": "string",
              "local_path": "string",
              "minimum_requirements": {
                "cpu": "int",
                "disk": "int",
                "memory": "int"
              },
              "pack_size": "int64",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string",
              "uuid": "string",
              "version": "string"
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
          body: |-
            {
              "architecture": "string",
              "catalog_id": "string",
              "client": "string",
              "connection": "string",
              "machine_name": "string",
              "owner": "string",
              "path": "string",
              "provider_metadata": "map[string]string",
              "start_after_pull": "bool",
              "version": "string"
            }
      response_blocks:
        - code_block: |-
            {
              "architecture": "string",
              "catalog_id": "string",
              "id": "string",
              "local_cache_path": "string",
              "local_path": "string",
              "machine_id": "string",
              "machine_name": "string",
              "manifest": "*VirtualMachineCatalogManifest",
              "version": "string"
            }
          code: "200"
          code_description: OK
          title: models.PullCatalogManifestResponse
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/pull' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"architecture\": \"string\",\n  \"catalog_id\": \"string\",\n  \"client\": \"string\",\n  \"connection\": \"string\",\n  \"machine_name\": \"string\",\n  \"owner\": \"string\",\n  \"path\": \"string\",\n  \"provider_metadata\": \"map[string]string\",\n  \"start_after_pull\": \"bool\",\n  \"version\": \"string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/catalog/pull");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "architecture": "string",
              "catalog_id": "string",
              "client": "string",
              "connection": "string",
              "machine_name": "string",
              "owner": "string",
              "path": "string",
              "provider_metadata": "map[string]string",
              "start_after_pull": "bool",
              "version": "string"
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
              url := "http://localhost/api/v1/catalog/pull"
              method := "put"
              payload := strings.NewReader(`{
              "architecture": "string",
              "catalog_id": "string",
              "client": "string",
              "connection": "string",
              "machine_name": "string",
              "owner": "string",
              "path": "string",
              "provider_metadata": "map[string]string",
              "start_after_pull": "bool",
              "version": "string"
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
          body: |-
            {
              "architecture": "string",
              "catalog_id": "string",
              "connection": "string",
              "provider_metadata": "map[string]string",
              "version": "string"
            }
      response_blocks:
        - code_block: |-
            {
              "id": "string",
              "local_path": "string",
              "machine_name": "string",
              "manifest": "*VirtualMachineCatalogManifest"
            }
          code: "200"
          code_description: OK
          title: models.ImportCatalogManifestResponse
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/import' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"architecture\": \"string\",\n  \"catalog_id\": \"string\",\n  \"connection\": \"string\",\n  \"provider_metadata\": \"map[string]string\",\n  \"version\": \"string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/catalog/import");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "architecture": "string",
              "catalog_id": "string",
              "connection": "string",
              "provider_metadata": "map[string]string",
              "version": "string"
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
              url := "http://localhost/api/v1/catalog/import"
              method := "put"
              payload := strings.NewReader(`{
              "architecture": "string",
              "catalog_id": "string",
              "connection": "string",
              "provider_metadata": "map[string]string",
              "version": "string"
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
          body: |-
            {
              "architecture": "string",
              "catalog_id": "string",
              "connection": "string",
              "description": "string",
              "force": "bool",
              "is_compressed": "bool",
              "machine_remote_path": "string",
              "provider_metadata": "map[string]string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "size": "int64",
              "tags": "[]string",
              "type": "string",
              "version": "string"
            }
      response_blocks:
        - code_block: |-
            {
              "id": "string",
              "local_path": "string",
              "machine_name": "string",
              "manifest": "*VirtualMachineCatalogManifest"
            }
          code: "200"
          code_description: OK
          title: models.ImportVmResponse
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/import-vm' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"architecture\": \"string\",\n  \"catalog_id\": \"string\",\n  \"connection\": \"string\",\n  \"description\": \"string\",\n  \"force\": \"bool\",\n  \"is_compressed\": \"bool\",\n  \"machine_remote_path\": \"string\",\n  \"provider_metadata\": \"map[string]string\",\n  \"required_claims\": \"[]string\",\n  \"required_roles\": \"[]string\",\n  \"size\": \"int64\",\n  \"tags\": \"[]string\",\n  \"type\": \"string\",\n  \"version\": \"string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/catalog/import-vm");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "architecture": "string",
              "catalog_id": "string",
              "connection": "string",
              "description": "string",
              "force": "bool",
              "is_compressed": "bool",
              "machine_remote_path": "string",
              "provider_metadata": "map[string]string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "size": "int64",
              "tags": "[]string",
              "type": "string",
              "version": "string"
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
              url := "http://localhost/api/v1/catalog/import-vm"
              method := "put"
              payload := strings.NewReader(`{
              "architecture": "string",
              "catalog_id": "string",
              "connection": "string",
              "description": "string",
              "force": "bool",
              "is_compressed": "bool",
              "machine_remote_path": "string",
              "provider_metadata": "map[string]string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "size": "int64",
              "tags": "[]string",
              "type": "string",
              "version": "string"
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
          body: |-
            {
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: models.CatalogManifest
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"connection\": \"string\",\n  \"required_claims\": \"[]string\",\n  \"required_roles\": \"[]string\",\n  \"tags\": \"[]string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Patch, "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
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
              url := "http://localhost/api/v1/catalog/{catalogId}/{version}/{architecture}/claims"
              method := "patch"
              payload := strings.NewReader(`{
              "connection": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "tags": "[]string"
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
    - title: Gets catalog cache
      description: This endpoint returns all the remote catalog cache if any
      requires_authorization: true
      category: Catalogs
      category_path: catalogs
      path: /v1/catalog/cache
      method: get
      response_blocks:
        - code_block: |-
            [
              {
                "architecture": "string",
                "catalog_id": "string",
                "created_at": "string",
                "description": "string",
                "download_count": "int",
                "id": "string",
                "is_compressed": "bool",
                "last_downloaded_at": "string",
                "last_downloaded_user": "string",
                "metadata_path": "string",
                "minimum_requirements": "*MinimumSpecRequirement",
                "name": "string",
                "pack_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ],
                "pack_path": "string",
                "pack_relative_path": "string",
                "pack_size": "int64",
                "path": "string",
                "provider": "*CatalogManifestProvider",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "revoked": "bool",
                "revoked_at": "string",
                "revoked_by": "string",
                "size": "int64",
                "tags": "[]string",
                "tainted": "bool",
                "tainted_at": "string",
                "tainted_by": "string",
                "type": "string",
                "untainted_by": "string",
                "updated_at": "string",
                "version": "string",
                "virtual_machine_contents": [
                  {
                    "created_at": "string",
                    "deleted_at": "string",
                    "hash": "string",
                    "is_dir": "bool",
                    "name": "string",
                    "path": "string",
                    "size": "int64",
                    "updated_at": "string"
                  }
                ]
              }
            ]
          code: "200"
          code_description: OK
          title: '[]models.CatalogManifest'
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/cache' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/catalog/cache");
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
              url := "http://localhost/api/v1/catalog/cache"
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
      path: /v1/catalog/cache
      method: delete
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
      response_blocks:
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/cache' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/catalog/cache");
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
              url := "http://localhost/api/v1/catalog/cache"
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
      path: /v1/catalog/cache/{catalogId}
      method: delete
      parameters:
        - name: catalogId
          required: true
          type: path
          value_type: string
          description: Catalog ID
      response_blocks:
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/cache/{catalogId}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/catalog/cache/{catalogId}");
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
              url := "http://localhost/api/v1/catalog/cache/{catalogId}"
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
      path: /v1/catalog/cache/{catalogId}/{version}
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
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "error": "string",
                  "path": "string"
                }
              ]
            }
          code: "400"
          code_description: Bad Request
          title: models.ApiErrorResponse
          language: json
        - code_block: |-
            {
              "error": "OAuthErrorType",
              "error_description": "string",
              "error_uri": "string"
            }
          code: "401"
          code_description: Unauthorized
          title: models.OAuthErrorResponse
          language: json
      example_blocks:
        - code_block: "curl --location 'http://localhost/api/v1/catalog/cache/{catalogId}/{version}' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Delete, "http://localhost/api/v1/catalog/cache/{catalogId}/{version}");
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
              url := "http://localhost/api/v1/catalog/cache/{catalogId}/{version}"
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
# Catalogs endpoints 

 This document contains the endpoints for the Catalogs category.


