---
layout: api
title: Orchestrator
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
    - title: Gets all hosts from the orchestrator
      description: This endpoint returns all hosts from the orchestrator
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/hosts
      method: get
      response_blocks:
        - code_block: |-
            [
              {
                "architecture": "string",
                "cpu_model": "string",
                "description": "string",
                "detailed_resources": "*HostResources",
                "devops_version": "string",
                "enabled": "bool",
                "external_ip_address": "string",
                "host": "string",
                "id": "string",
                "is_reverse_proxy_enabled": "bool",
                "os_name": "string",
                "os_version": "string",
                "parallels_desktop_licensed": "bool",
                "parallels_desktop_version": "string",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "resources": {
                  "disk_size": "float64",
                  "free_disk_size": "float64",
                  "logical_cpu_count": "int64",
                  "memory_size": "float64",
                  "physical_cpu_count": "int64",
                  "total_apple_vms": "int64"
                },
                "reverse_proxy": "*HostReverseProxy",
                "reverse_proxy_hosts": "[]*ReverseProxyHost",
                "state": "string",
                "tags": "[]string"
              }
            ]
          code: "200"
          code_description: OK
          title: '[]models.OrchestratorHostResponse'
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
        - code_block: |-
            [
              {
                "architecture": "string",
                "cpu_model": "string",
                "description": "string",
                "detailed_resources": "*HostResources",
                "devops_version": "string",
                "enabled": "bool",
                "external_ip_address": "string",
                "host": "string",
                "id": "string",
                "is_reverse_proxy_enabled": "bool",
                "os_name": "string",
                "os_version": "string",
                "parallels_desktop_licensed": "bool",
                "parallels_desktop_version": "string",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "resources": {
                  "disk_size": "float64",
                  "free_disk_size": "float64",
                  "logical_cpu_count": "int64",
                  "memory_size": "float64",
                  "physical_cpu_count": "int64",
                  "total_apple_vms": "int64"
                },
                "reverse_proxy": "*HostReverseProxy",
                "reverse_proxy_hosts": "[]*ReverseProxyHost",
                "state": "string",
                "tags": "[]string"
              }
            ]
          code: "200"
          code_description: OK
          title: models.OrchestratorHostResponse
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
        - code_block: |-
            {
              "cpu_brand": "string",
              "cpu_type": "string",
              "devops_version": "string",
              "external_ip_address": "string",
              "is_log_streaming_enabled": "bool",
              "is_reverse_proxy_enabled": "bool",
              "os_name": "string",
              "os_version": "string",
              "parallels_desktop_licensed": "bool",
              "parallels_desktop_version": "string",
              "reverse_proxy": "*SystemReverseProxy",
              "system_reserved": "*SystemUsageItem",
              "total": "*SystemUsageItem",
              "total_available": "*SystemUsageItem",
              "total_in_use": "*SystemUsageItem",
              "total_reserved": "*SystemUsageItem"
            }
          code: "200"
          code_description: OK
          title: models.SystemUsageResponse
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
          body: |-
            {
              "authentication": "*OrchestratorAuthentication",
              "description": "string",
              "host": "string",
              "port": "string",
              "prefix": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "schema": "string",
              "tags": "[]string"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "architecture": "string",
                "cpu_model": "string",
                "description": "string",
                "detailed_resources": "*HostResources",
                "devops_version": "string",
                "enabled": "bool",
                "external_ip_address": "string",
                "host": "string",
                "id": "string",
                "is_reverse_proxy_enabled": "bool",
                "os_name": "string",
                "os_version": "string",
                "parallels_desktop_licensed": "bool",
                "parallels_desktop_version": "string",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "resources": {
                  "disk_size": "float64",
                  "free_disk_size": "float64",
                  "logical_cpu_count": "int64",
                  "memory_size": "float64",
                  "physical_cpu_count": "int64",
                  "total_apple_vms": "int64"
                },
                "reverse_proxy": "*HostReverseProxy",
                "reverse_proxy_hosts": "[]*ReverseProxyHost",
                "state": "string",
                "tags": "[]string"
              }
            ]
          code: "200"
          code_description: OK
          title: models.OrchestratorHostResponse
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"authentication\": \"*OrchestratorAuthentication\",\n  \"description\": \"string\",\n  \"host\": \"string\",\n  \"port\": \"string\",\n  \"prefix\": \"string\",\n  \"required_claims\": \"[]string\",\n  \"required_roles\": \"[]string\",\n  \"schema\": \"string\",\n  \"tags\": \"[]string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "authentication": "*OrchestratorAuthentication",
              "description": "string",
              "host": "string",
              "port": "string",
              "prefix": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "schema": "string",
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
              url := "http://localhost/api/v1/orchestrator/hosts"
              method := "post"
              payload := strings.NewReader(`{
              "authentication": "*OrchestratorAuthentication",
              "description": "string",
              "host": "string",
              "port": "string",
              "prefix": "string",
              "required_claims": "[]string",
              "required_roles": "[]string",
              "schema": "string",
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
        - code_block: |-
            [
              {
                "architecture": "string",
                "cpu_model": "string",
                "description": "string",
                "detailed_resources": "*HostResources",
                "devops_version": "string",
                "enabled": "bool",
                "external_ip_address": "string",
                "host": "string",
                "id": "string",
                "is_reverse_proxy_enabled": "bool",
                "os_name": "string",
                "os_version": "string",
                "parallels_desktop_licensed": "bool",
                "parallels_desktop_version": "string",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "resources": {
                  "disk_size": "float64",
                  "free_disk_size": "float64",
                  "logical_cpu_count": "int64",
                  "memory_size": "float64",
                  "physical_cpu_count": "int64",
                  "total_apple_vms": "int64"
                },
                "reverse_proxy": "*HostReverseProxy",
                "reverse_proxy_hosts": "[]*ReverseProxyHost",
                "state": "string",
                "tags": "[]string"
              }
            ]
          code: "200"
          code_description: OK
          title: models.OrchestratorHostResponse
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
        - code_block: |-
            [
              {
                "architecture": "string",
                "cpu_model": "string",
                "description": "string",
                "detailed_resources": "*HostResources",
                "devops_version": "string",
                "enabled": "bool",
                "external_ip_address": "string",
                "host": "string",
                "id": "string",
                "is_reverse_proxy_enabled": "bool",
                "os_name": "string",
                "os_version": "string",
                "parallels_desktop_licensed": "bool",
                "parallels_desktop_version": "string",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "resources": {
                  "disk_size": "float64",
                  "free_disk_size": "float64",
                  "logical_cpu_count": "int64",
                  "memory_size": "float64",
                  "physical_cpu_count": "int64",
                  "total_apple_vms": "int64"
                },
                "reverse_proxy": "*HostReverseProxy",
                "reverse_proxy_hosts": "[]*ReverseProxyHost",
                "state": "string",
                "tags": "[]string"
              }
            ]
          code: "200"
          code_description: OK
          title: models.OrchestratorHostResponse
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
          body: |-
            {
              "authentication": "*OrchestratorAuthentication",
              "description": "string",
              "host": "string",
              "port": "string",
              "prefix": "string",
              "schema": "string"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "architecture": "string",
                "cpu_model": "string",
                "description": "string",
                "detailed_resources": "*HostResources",
                "devops_version": "string",
                "enabled": "bool",
                "external_ip_address": "string",
                "host": "string",
                "id": "string",
                "is_reverse_proxy_enabled": "bool",
                "os_name": "string",
                "os_version": "string",
                "parallels_desktop_licensed": "bool",
                "parallels_desktop_version": "string",
                "required_claims": "[]string",
                "required_roles": "[]string",
                "resources": {
                  "disk_size": "float64",
                  "free_disk_size": "float64",
                  "logical_cpu_count": "int64",
                  "memory_size": "float64",
                  "physical_cpu_count": "int64",
                  "total_apple_vms": "int64"
                },
                "reverse_proxy": "*HostReverseProxy",
                "reverse_proxy_hosts": "[]*ReverseProxyHost",
                "state": "string",
                "tags": "[]string"
              }
            ]
          code: "200"
          code_description: OK
          title: models.OrchestratorHostResponse
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"authentication\": \"*OrchestratorAuthentication\",\n  \"description\": \"string\",\n  \"host\": \"string\",\n  \"port\": \"string\",\n  \"prefix\": \"string\",\n  \"schema\": \"string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "authentication": "*OrchestratorAuthentication",
              "description": "string",
              "host": "string",
              "port": "string",
              "prefix": "string",
              "schema": "string"
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
              url := "http://localhost/api/v1/orchestrator/hosts"
              method := "put"
              payload := strings.NewReader(`{
              "authentication": "*OrchestratorAuthentication",
              "description": "string",
              "host": "string",
              "port": "string",
              "prefix": "string",
              "schema": "string"
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
    - title: Get orchestrator resource overview
      description: This endpoint returns orchestrator resource overview
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/overview/resources
      method: get
      response_blocks:
        - code_block: |-
            {
              "cpu_brand": "string",
              "cpu_type": "string",
              "system_reserved": "HostResourceItem",
              "total": "HostResourceItem",
              "total_available": "HostResourceItem",
              "total_in_use": "HostResourceItem",
              "total_reserved": "HostResourceItem"
            }
          code: "200"
          code_description: OK
          title: models.HostResourceOverviewResponse
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
        - code_block: |-
            {
              "cpu_brand": "string",
              "cpu_type": "string",
              "system_reserved": "HostResourceItem",
              "total": "HostResourceItem",
              "total_available": "HostResourceItem",
              "total_in_use": "HostResourceItem",
              "total_reserved": "HostResourceItem"
            }
          code: "200"
          code_description: OK
          title: models.HostResourceOverviewResponse
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
        - code_block: |-
            [
              {
                "Advanced": {
                  "Public SSH keys synchronization": "string",
                  "Rosetta Linux": "string",
                  "Share host location": "string",
                  "Show developer tools": "string",
                  "Swipe from edges": "string",
                  "VM hostname synchronization": "string"
                },
                "Allow select boot device": "string",
                "BIOS type": "string",
                "Boot order": "string",
                "Coherence": {
                  "Auto-switch to full screen": "string",
                  "Disable aero": "string",
                  "Hide minimized windows": "string",
                  "Show Windows systray in Mac menu": "string"
                },
                "Description": "string",
                "EFI Secure boot": "string",
                "Expiration": {
                  "enabled": "bool"
                },
                "External boot device": "string",
                "Fullscreen": {
                  "Activate spaces on click": "string",
                  "Gamma control": "string",
                  "Optimize for games": "string",
                  "Scale view mode": "string",
                  "Use all displays": "string"
                },
                "Guest Shared Folders": {
                  "Automount": "string",
                  "enabled": "bool"
                },
                "GuestTools": {
                  "state": "string",
                  "version": "string"
                },
                "Hardware": {
                  "cdrom0": {
                    "enabled": "bool",
                    "image": "string",
                    "port": "string",
                    "state": "string"
                  },
                  "cpu": {
                    "VT-x": "bool",
                    "accl": "string",
                    "auto": "string",
                    "cpus": "int64",
                    "hotplug": "bool",
                    "mode": "string",
                    "type": "string"
                  },
                  "hdd0": {
                    "enabled": "bool",
                    "image": "string",
                    "online-compact": "string",
                    "port": "string",
                    "size": "string",
                    "type": "string"
                  },
                  "memory": {
                    "auto": "string",
                    "hotplug": "bool",
                    "size": "string"
                  },
                  "memory_quota": {
                    "auto": "string"
                  },
                  "net0": {
                    "card": "string",
                    "enabled": "bool",
                    "mac": "string",
                    "type": "string"
                  },
                  "sound0": {
                    "enabled": "bool",
                    "mixer": "string",
                    "output": "string"
                  },
                  "usb": {
                    "enabled": "bool"
                  },
                  "video": {
                    "3d-acceleration": "string",
                    "adapter-type": "string",
                    "automatic-video-memory": "string",
                    "high-resolution": "string",
                    "high-resolution-in-guest": "string",
                    "native-scaling-in-guest": "string",
                    "size": "string",
                    "vertical-sync": "string"
                  }
                },
                "Home": "string",
                "Home path": "string",
                "Host Shared Folders": "map[string]unknown",
                "Host defined sharing": "string",
                "ID": "string",
                "Miscellaneous Sharing": {
                  "Shared clipboard": "string",
                  "Shared cloud": "string"
                },
                "Modality": {
                  "Capture mouse clicks": "string",
                  "Opacity (percentage)": "int64",
                  "Show on all spaces ": "string",
                  "Stay on top": "string"
                },
                "Mouse and Keyboard": {
                  "Keyboard optimization mode": "string",
                  "Smart mouse optimized for games": "string",
                  "Smooth scrolling": "string",
                  "Sticky mouse": "string"
                },
                "Name": "string",
                "Network": {
                  "Conditioned": "string",
                  "Inbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "Outbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "ipAddresses": [
                    {
                      "ip": "string",
                      "type": "string"
                    }
                  ]
                },
                "OS": "string",
                "Optimization": {
                  "Adaptive hypervisor": "string",
                  "Auto compress virtual disks": "string",
                  "Disabled Windows logo": "string",
                  "Faster virtual machine": "string",
                  "Hypervisor type": "string",
                  "Longer battery life": "string",
                  "Nested virtualization": "string",
                  "PMU virtualization": "string",
                  "Resource quota": "string",
                  "Show battery status": "string"
                },
                "Print Management": {
                  "Show host printer UI": "string",
                  "Synchronize default printer": "string",
                  "Synchronize with host printers": "string"
                },
                "Restore Image": "string",
                "SMBIOS settings": {
                  "BIOS Version": "string",
                  "Board Manufacturer": "string",
                  "System serial number": "string"
                },
                "Security": {
                  "Archived": "string",
                  "Configuration is locked": "string",
                  "Custom password protection": "string",
                  "Encrypted": "string",
                  "Packed": "string",
                  "Protected": "string",
                  "TPM enabled": "string",
                  "TPM type": "string"
                },
                "Shared Applications": {
                  "Bounce dock icon when app flashes": "string",
                  "Guest-to-host apps sharing": "string",
                  "Host-to-guest apps sharing": "string",
                  "Show guest apps folder in Dock": "string",
                  "Show guest notifications": "string",
                  "enabled": "bool"
                },
                "Shared Profile": {
                  "enabled": "bool"
                },
                "Smart Guard": {
                  "enabled": "bool"
                },
                "SmartMount": {
                  "CD/DVD drives": "string",
                  "Network shares": "string",
                  "Removable drives": "string",
                  "enabled": "bool"
                },
                "Startup and Shutdown": {
                  "Autostart": "string",
                  "Autostart delay": "int64",
                  "Autostop": "string",
                  "On shutdown": "string",
                  "On window close": "string",
                  "Pause idle": "string",
                  "Startup view": "string",
                  "Undo disks": "string"
                },
                "State": "string",
                "Template": "string",
                "Time Synchronization": {
                  "Interval (in seconds)": "int64",
                  "Smart mode": "string",
                  "Timezone synchronization disabled": "string",
                  "enabled": "bool"
                },
                "Travel mode": {
                  "Enter condition": "string",
                  "Enter threshold": "int64",
                  "Quit condition": "string"
                },
                "Type": "string",
                "USB and Bluetooth": {
                  "Automatic sharing bluetooth": "string",
                  "Automatic sharing cameras": "string",
                  "Automatic sharing gamepads": "string",
                  "Automatic sharing smart cards": "string",
                  "Support USB 3.0": "string"
                },
                "Uptime": "string",
                "host": "string",
                "host_external_ip_address": "string",
                "host_id": "string",
                "host_state": "string",
                "host_url": "string",
                "internal_ip_address": "string",
                "user": "string"
              }
            ]
          code: "200"
          code_description: OK
          title: '[]models.ParallelsVM'
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
        - code_block: |-
            [
              {
                "Advanced": {
                  "Public SSH keys synchronization": "string",
                  "Rosetta Linux": "string",
                  "Share host location": "string",
                  "Show developer tools": "string",
                  "Swipe from edges": "string",
                  "VM hostname synchronization": "string"
                },
                "Allow select boot device": "string",
                "BIOS type": "string",
                "Boot order": "string",
                "Coherence": {
                  "Auto-switch to full screen": "string",
                  "Disable aero": "string",
                  "Hide minimized windows": "string",
                  "Show Windows systray in Mac menu": "string"
                },
                "Description": "string",
                "EFI Secure boot": "string",
                "Expiration": {
                  "enabled": "bool"
                },
                "External boot device": "string",
                "Fullscreen": {
                  "Activate spaces on click": "string",
                  "Gamma control": "string",
                  "Optimize for games": "string",
                  "Scale view mode": "string",
                  "Use all displays": "string"
                },
                "Guest Shared Folders": {
                  "Automount": "string",
                  "enabled": "bool"
                },
                "GuestTools": {
                  "state": "string",
                  "version": "string"
                },
                "Hardware": {
                  "cdrom0": {
                    "enabled": "bool",
                    "image": "string",
                    "port": "string",
                    "state": "string"
                  },
                  "cpu": {
                    "VT-x": "bool",
                    "accl": "string",
                    "auto": "string",
                    "cpus": "int64",
                    "hotplug": "bool",
                    "mode": "string",
                    "type": "string"
                  },
                  "hdd0": {
                    "enabled": "bool",
                    "image": "string",
                    "online-compact": "string",
                    "port": "string",
                    "size": "string",
                    "type": "string"
                  },
                  "memory": {
                    "auto": "string",
                    "hotplug": "bool",
                    "size": "string"
                  },
                  "memory_quota": {
                    "auto": "string"
                  },
                  "net0": {
                    "card": "string",
                    "enabled": "bool",
                    "mac": "string",
                    "type": "string"
                  },
                  "sound0": {
                    "enabled": "bool",
                    "mixer": "string",
                    "output": "string"
                  },
                  "usb": {
                    "enabled": "bool"
                  },
                  "video": {
                    "3d-acceleration": "string",
                    "adapter-type": "string",
                    "automatic-video-memory": "string",
                    "high-resolution": "string",
                    "high-resolution-in-guest": "string",
                    "native-scaling-in-guest": "string",
                    "size": "string",
                    "vertical-sync": "string"
                  }
                },
                "Home": "string",
                "Home path": "string",
                "Host Shared Folders": "map[string]unknown",
                "Host defined sharing": "string",
                "ID": "string",
                "Miscellaneous Sharing": {
                  "Shared clipboard": "string",
                  "Shared cloud": "string"
                },
                "Modality": {
                  "Capture mouse clicks": "string",
                  "Opacity (percentage)": "int64",
                  "Show on all spaces ": "string",
                  "Stay on top": "string"
                },
                "Mouse and Keyboard": {
                  "Keyboard optimization mode": "string",
                  "Smart mouse optimized for games": "string",
                  "Smooth scrolling": "string",
                  "Sticky mouse": "string"
                },
                "Name": "string",
                "Network": {
                  "Conditioned": "string",
                  "Inbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "Outbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "ipAddresses": [
                    {
                      "ip": "string",
                      "type": "string"
                    }
                  ]
                },
                "OS": "string",
                "Optimization": {
                  "Adaptive hypervisor": "string",
                  "Auto compress virtual disks": "string",
                  "Disabled Windows logo": "string",
                  "Faster virtual machine": "string",
                  "Hypervisor type": "string",
                  "Longer battery life": "string",
                  "Nested virtualization": "string",
                  "PMU virtualization": "string",
                  "Resource quota": "string",
                  "Show battery status": "string"
                },
                "Print Management": {
                  "Show host printer UI": "string",
                  "Synchronize default printer": "string",
                  "Synchronize with host printers": "string"
                },
                "Restore Image": "string",
                "SMBIOS settings": {
                  "BIOS Version": "string",
                  "Board Manufacturer": "string",
                  "System serial number": "string"
                },
                "Security": {
                  "Archived": "string",
                  "Configuration is locked": "string",
                  "Custom password protection": "string",
                  "Encrypted": "string",
                  "Packed": "string",
                  "Protected": "string",
                  "TPM enabled": "string",
                  "TPM type": "string"
                },
                "Shared Applications": {
                  "Bounce dock icon when app flashes": "string",
                  "Guest-to-host apps sharing": "string",
                  "Host-to-guest apps sharing": "string",
                  "Show guest apps folder in Dock": "string",
                  "Show guest notifications": "string",
                  "enabled": "bool"
                },
                "Shared Profile": {
                  "enabled": "bool"
                },
                "Smart Guard": {
                  "enabled": "bool"
                },
                "SmartMount": {
                  "CD/DVD drives": "string",
                  "Network shares": "string",
                  "Removable drives": "string",
                  "enabled": "bool"
                },
                "Startup and Shutdown": {
                  "Autostart": "string",
                  "Autostart delay": "int64",
                  "Autostop": "string",
                  "On shutdown": "string",
                  "On window close": "string",
                  "Pause idle": "string",
                  "Startup view": "string",
                  "Undo disks": "string"
                },
                "State": "string",
                "Template": "string",
                "Time Synchronization": {
                  "Interval (in seconds)": "int64",
                  "Smart mode": "string",
                  "Timezone synchronization disabled": "string",
                  "enabled": "bool"
                },
                "Travel mode": {
                  "Enter condition": "string",
                  "Enter threshold": "int64",
                  "Quit condition": "string"
                },
                "Type": "string",
                "USB and Bluetooth": {
                  "Automatic sharing bluetooth": "string",
                  "Automatic sharing cameras": "string",
                  "Automatic sharing gamepads": "string",
                  "Automatic sharing smart cards": "string",
                  "Support USB 3.0": "string"
                },
                "Uptime": "string",
                "host": "string",
                "host_external_ip_address": "string",
                "host_id": "string",
                "host_state": "string",
                "host_url": "string",
                "internal_ip_address": "string",
                "user": "string"
              }
            ]
          code: "200"
          code_description: OK
          title: models.ParallelsVM
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
      path: /v1/orchestrator/machines/{vmId}/status
      method: get
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: |-
            [
              {
                "Advanced": {
                  "Public SSH keys synchronization": "string",
                  "Rosetta Linux": "string",
                  "Share host location": "string",
                  "Show developer tools": "string",
                  "Swipe from edges": "string",
                  "VM hostname synchronization": "string"
                },
                "Allow select boot device": "string",
                "BIOS type": "string",
                "Boot order": "string",
                "Coherence": {
                  "Auto-switch to full screen": "string",
                  "Disable aero": "string",
                  "Hide minimized windows": "string",
                  "Show Windows systray in Mac menu": "string"
                },
                "Description": "string",
                "EFI Secure boot": "string",
                "Expiration": {
                  "enabled": "bool"
                },
                "External boot device": "string",
                "Fullscreen": {
                  "Activate spaces on click": "string",
                  "Gamma control": "string",
                  "Optimize for games": "string",
                  "Scale view mode": "string",
                  "Use all displays": "string"
                },
                "Guest Shared Folders": {
                  "Automount": "string",
                  "enabled": "bool"
                },
                "GuestTools": {
                  "state": "string",
                  "version": "string"
                },
                "Hardware": {
                  "cdrom0": {
                    "enabled": "bool",
                    "image": "string",
                    "port": "string",
                    "state": "string"
                  },
                  "cpu": {
                    "VT-x": "bool",
                    "accl": "string",
                    "auto": "string",
                    "cpus": "int64",
                    "hotplug": "bool",
                    "mode": "string",
                    "type": "string"
                  },
                  "hdd0": {
                    "enabled": "bool",
                    "image": "string",
                    "online-compact": "string",
                    "port": "string",
                    "size": "string",
                    "type": "string"
                  },
                  "memory": {
                    "auto": "string",
                    "hotplug": "bool",
                    "size": "string"
                  },
                  "memory_quota": {
                    "auto": "string"
                  },
                  "net0": {
                    "card": "string",
                    "enabled": "bool",
                    "mac": "string",
                    "type": "string"
                  },
                  "sound0": {
                    "enabled": "bool",
                    "mixer": "string",
                    "output": "string"
                  },
                  "usb": {
                    "enabled": "bool"
                  },
                  "video": {
                    "3d-acceleration": "string",
                    "adapter-type": "string",
                    "automatic-video-memory": "string",
                    "high-resolution": "string",
                    "high-resolution-in-guest": "string",
                    "native-scaling-in-guest": "string",
                    "size": "string",
                    "vertical-sync": "string"
                  }
                },
                "Home": "string",
                "Home path": "string",
                "Host Shared Folders": "map[string]unknown",
                "Host defined sharing": "string",
                "ID": "string",
                "Miscellaneous Sharing": {
                  "Shared clipboard": "string",
                  "Shared cloud": "string"
                },
                "Modality": {
                  "Capture mouse clicks": "string",
                  "Opacity (percentage)": "int64",
                  "Show on all spaces ": "string",
                  "Stay on top": "string"
                },
                "Mouse and Keyboard": {
                  "Keyboard optimization mode": "string",
                  "Smart mouse optimized for games": "string",
                  "Smooth scrolling": "string",
                  "Sticky mouse": "string"
                },
                "Name": "string",
                "Network": {
                  "Conditioned": "string",
                  "Inbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "Outbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "ipAddresses": [
                    {
                      "ip": "string",
                      "type": "string"
                    }
                  ]
                },
                "OS": "string",
                "Optimization": {
                  "Adaptive hypervisor": "string",
                  "Auto compress virtual disks": "string",
                  "Disabled Windows logo": "string",
                  "Faster virtual machine": "string",
                  "Hypervisor type": "string",
                  "Longer battery life": "string",
                  "Nested virtualization": "string",
                  "PMU virtualization": "string",
                  "Resource quota": "string",
                  "Show battery status": "string"
                },
                "Print Management": {
                  "Show host printer UI": "string",
                  "Synchronize default printer": "string",
                  "Synchronize with host printers": "string"
                },
                "Restore Image": "string",
                "SMBIOS settings": {
                  "BIOS Version": "string",
                  "Board Manufacturer": "string",
                  "System serial number": "string"
                },
                "Security": {
                  "Archived": "string",
                  "Configuration is locked": "string",
                  "Custom password protection": "string",
                  "Encrypted": "string",
                  "Packed": "string",
                  "Protected": "string",
                  "TPM enabled": "string",
                  "TPM type": "string"
                },
                "Shared Applications": {
                  "Bounce dock icon when app flashes": "string",
                  "Guest-to-host apps sharing": "string",
                  "Host-to-guest apps sharing": "string",
                  "Show guest apps folder in Dock": "string",
                  "Show guest notifications": "string",
                  "enabled": "bool"
                },
                "Shared Profile": {
                  "enabled": "bool"
                },
                "Smart Guard": {
                  "enabled": "bool"
                },
                "SmartMount": {
                  "CD/DVD drives": "string",
                  "Network shares": "string",
                  "Removable drives": "string",
                  "enabled": "bool"
                },
                "Startup and Shutdown": {
                  "Autostart": "string",
                  "Autostart delay": "int64",
                  "Autostop": "string",
                  "On shutdown": "string",
                  "On window close": "string",
                  "Pause idle": "string",
                  "Startup view": "string",
                  "Undo disks": "string"
                },
                "State": "string",
                "Template": "string",
                "Time Synchronization": {
                  "Interval (in seconds)": "int64",
                  "Smart mode": "string",
                  "Timezone synchronization disabled": "string",
                  "enabled": "bool"
                },
                "Travel mode": {
                  "Enter condition": "string",
                  "Enter threshold": "int64",
                  "Quit condition": "string"
                },
                "Type": "string",
                "USB and Bluetooth": {
                  "Automatic sharing bluetooth": "string",
                  "Automatic sharing cameras": "string",
                  "Automatic sharing gamepads": "string",
                  "Automatic sharing smart cards": "string",
                  "Support USB 3.0": "string"
                },
                "Uptime": "string",
                "host": "string",
                "host_external_ip_address": "string",
                "host_id": "string",
                "host_state": "string",
                "host_url": "string",
                "internal_ip_address": "string",
                "user": "string"
              }
            ]
          code: "200"
          code_description: OK
          title: models.ParallelsVM
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{vmId}/status' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Get, "http://localhost/api/v1/orchestrator/machines/{vmId}/status");
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
              url := "http://localhost/api/v1/orchestrator/machines/{vmId}/status"
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
        - code_block: |-
            [
              {
                "Advanced": {
                  "Public SSH keys synchronization": "string",
                  "Rosetta Linux": "string",
                  "Share host location": "string",
                  "Show developer tools": "string",
                  "Swipe from edges": "string",
                  "VM hostname synchronization": "string"
                },
                "Allow select boot device": "string",
                "BIOS type": "string",
                "Boot order": "string",
                "Coherence": {
                  "Auto-switch to full screen": "string",
                  "Disable aero": "string",
                  "Hide minimized windows": "string",
                  "Show Windows systray in Mac menu": "string"
                },
                "Description": "string",
                "EFI Secure boot": "string",
                "Expiration": {
                  "enabled": "bool"
                },
                "External boot device": "string",
                "Fullscreen": {
                  "Activate spaces on click": "string",
                  "Gamma control": "string",
                  "Optimize for games": "string",
                  "Scale view mode": "string",
                  "Use all displays": "string"
                },
                "Guest Shared Folders": {
                  "Automount": "string",
                  "enabled": "bool"
                },
                "GuestTools": {
                  "state": "string",
                  "version": "string"
                },
                "Hardware": {
                  "cdrom0": {
                    "enabled": "bool",
                    "image": "string",
                    "port": "string",
                    "state": "string"
                  },
                  "cpu": {
                    "VT-x": "bool",
                    "accl": "string",
                    "auto": "string",
                    "cpus": "int64",
                    "hotplug": "bool",
                    "mode": "string",
                    "type": "string"
                  },
                  "hdd0": {
                    "enabled": "bool",
                    "image": "string",
                    "online-compact": "string",
                    "port": "string",
                    "size": "string",
                    "type": "string"
                  },
                  "memory": {
                    "auto": "string",
                    "hotplug": "bool",
                    "size": "string"
                  },
                  "memory_quota": {
                    "auto": "string"
                  },
                  "net0": {
                    "card": "string",
                    "enabled": "bool",
                    "mac": "string",
                    "type": "string"
                  },
                  "sound0": {
                    "enabled": "bool",
                    "mixer": "string",
                    "output": "string"
                  },
                  "usb": {
                    "enabled": "bool"
                  },
                  "video": {
                    "3d-acceleration": "string",
                    "adapter-type": "string",
                    "automatic-video-memory": "string",
                    "high-resolution": "string",
                    "high-resolution-in-guest": "string",
                    "native-scaling-in-guest": "string",
                    "size": "string",
                    "vertical-sync": "string"
                  }
                },
                "Home": "string",
                "Home path": "string",
                "Host Shared Folders": "map[string]unknown",
                "Host defined sharing": "string",
                "ID": "string",
                "Miscellaneous Sharing": {
                  "Shared clipboard": "string",
                  "Shared cloud": "string"
                },
                "Modality": {
                  "Capture mouse clicks": "string",
                  "Opacity (percentage)": "int64",
                  "Show on all spaces ": "string",
                  "Stay on top": "string"
                },
                "Mouse and Keyboard": {
                  "Keyboard optimization mode": "string",
                  "Smart mouse optimized for games": "string",
                  "Smooth scrolling": "string",
                  "Sticky mouse": "string"
                },
                "Name": "string",
                "Network": {
                  "Conditioned": "string",
                  "Inbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "Outbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "ipAddresses": [
                    {
                      "ip": "string",
                      "type": "string"
                    }
                  ]
                },
                "OS": "string",
                "Optimization": {
                  "Adaptive hypervisor": "string",
                  "Auto compress virtual disks": "string",
                  "Disabled Windows logo": "string",
                  "Faster virtual machine": "string",
                  "Hypervisor type": "string",
                  "Longer battery life": "string",
                  "Nested virtualization": "string",
                  "PMU virtualization": "string",
                  "Resource quota": "string",
                  "Show battery status": "string"
                },
                "Print Management": {
                  "Show host printer UI": "string",
                  "Synchronize default printer": "string",
                  "Synchronize with host printers": "string"
                },
                "Restore Image": "string",
                "SMBIOS settings": {
                  "BIOS Version": "string",
                  "Board Manufacturer": "string",
                  "System serial number": "string"
                },
                "Security": {
                  "Archived": "string",
                  "Configuration is locked": "string",
                  "Custom password protection": "string",
                  "Encrypted": "string",
                  "Packed": "string",
                  "Protected": "string",
                  "TPM enabled": "string",
                  "TPM type": "string"
                },
                "Shared Applications": {
                  "Bounce dock icon when app flashes": "string",
                  "Guest-to-host apps sharing": "string",
                  "Host-to-guest apps sharing": "string",
                  "Show guest apps folder in Dock": "string",
                  "Show guest notifications": "string",
                  "enabled": "bool"
                },
                "Shared Profile": {
                  "enabled": "bool"
                },
                "Smart Guard": {
                  "enabled": "bool"
                },
                "SmartMount": {
                  "CD/DVD drives": "string",
                  "Network shares": "string",
                  "Removable drives": "string",
                  "enabled": "bool"
                },
                "Startup and Shutdown": {
                  "Autostart": "string",
                  "Autostart delay": "int64",
                  "Autostop": "string",
                  "On shutdown": "string",
                  "On window close": "string",
                  "Pause idle": "string",
                  "Startup view": "string",
                  "Undo disks": "string"
                },
                "State": "string",
                "Template": "string",
                "Time Synchronization": {
                  "Interval (in seconds)": "int64",
                  "Smart mode": "string",
                  "Timezone synchronization disabled": "string",
                  "enabled": "bool"
                },
                "Travel mode": {
                  "Enter condition": "string",
                  "Enter threshold": "int64",
                  "Quit condition": "string"
                },
                "Type": "string",
                "USB and Bluetooth": {
                  "Automatic sharing bluetooth": "string",
                  "Automatic sharing cameras": "string",
                  "Automatic sharing gamepads": "string",
                  "Automatic sharing smart cards": "string",
                  "Support USB 3.0": "string"
                },
                "Uptime": "string",
                "host": "string",
                "host_external_ip_address": "string",
                "host_id": "string",
                "host_state": "string",
                "host_url": "string",
                "internal_ip_address": "string",
                "user": "string"
              }
            ]
          code: "200"
          code_description: OK
          title: models.ParallelsVM
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
      path: /v1/orchestrator/machines/{vmId}/set
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: |-
            {
              "operations": [
                {
                  "error": "string",
                  "group": "string",
                  "operation": "string",
                  "status": "string"
                }
              ]
            }
          code: "200"
          code_description: OK
          title: models.VirtualMachineConfigResponse
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{vmId}/set' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{vmId}/set");
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
              url := "http://localhost/api/v1/orchestrator/machines/{vmId}/set"
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
      path: /v1/orchestrator/machines/{vmId}/start
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: |-
            {
              "operations": [
                {
                  "error": "string",
                  "group": "string",
                  "operation": "string",
                  "status": "string"
                }
              ]
            }
          code: "200"
          code_description: OK
          title: models.VirtualMachineConfigResponse
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{vmId}/start' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{vmId}/start");
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
              url := "http://localhost/api/v1/orchestrator/machines/{vmId}/start"
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
      path: /v1/orchestrator/machines/{vmId}/stop
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: |-
            {
              "operations": [
                {
                  "error": "string",
                  "group": "string",
                  "operation": "string",
                  "status": "string"
                }
              ]
            }
          code: "200"
          code_description: OK
          title: models.VirtualMachineConfigResponse
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{vmId}/stop' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{vmId}/stop");
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
              url := "http://localhost/api/v1/orchestrator/machines/{vmId}/stop"
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
    - title: Executes a command in a orchestrator virtual machine
      description: This endpoint executes a command in a orchestrator virtual machine
      requires_authorization: true
      category: Orchestrator
      category_path: orchestrator
      path: /v1/orchestrator/machines/{vmId}/execute
      method: put
      parameters:
        - name: id
          required: true
          type: path
          value_type: string
          description: Virtual Machine ID
      response_blocks:
        - code_block: |-
            {
              "operations": [
                {
                  "error": "string",
                  "group": "string",
                  "operation": "string",
                  "status": "string"
                }
              ]
            }
          code: "200"
          code_description: OK
          title: models.VirtualMachineConfigResponse
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines/{vmId}/execute' \n--header 'Authorization ••••••'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/machines/{vmId}/execute");
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
              url := "http://localhost/api/v1/orchestrator/machines/{vmId}/execute"
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
        - code_block: |-
            [
              {
                "Advanced": {
                  "Public SSH keys synchronization": "string",
                  "Rosetta Linux": "string",
                  "Share host location": "string",
                  "Show developer tools": "string",
                  "Swipe from edges": "string",
                  "VM hostname synchronization": "string"
                },
                "Allow select boot device": "string",
                "BIOS type": "string",
                "Boot order": "string",
                "Coherence": {
                  "Auto-switch to full screen": "string",
                  "Disable aero": "string",
                  "Hide minimized windows": "string",
                  "Show Windows systray in Mac menu": "string"
                },
                "Description": "string",
                "EFI Secure boot": "string",
                "Expiration": {
                  "enabled": "bool"
                },
                "External boot device": "string",
                "Fullscreen": {
                  "Activate spaces on click": "string",
                  "Gamma control": "string",
                  "Optimize for games": "string",
                  "Scale view mode": "string",
                  "Use all displays": "string"
                },
                "Guest Shared Folders": {
                  "Automount": "string",
                  "enabled": "bool"
                },
                "GuestTools": {
                  "state": "string",
                  "version": "string"
                },
                "Hardware": {
                  "cdrom0": {
                    "enabled": "bool",
                    "image": "string",
                    "port": "string",
                    "state": "string"
                  },
                  "cpu": {
                    "VT-x": "bool",
                    "accl": "string",
                    "auto": "string",
                    "cpus": "int64",
                    "hotplug": "bool",
                    "mode": "string",
                    "type": "string"
                  },
                  "hdd0": {
                    "enabled": "bool",
                    "image": "string",
                    "online-compact": "string",
                    "port": "string",
                    "size": "string",
                    "type": "string"
                  },
                  "memory": {
                    "auto": "string",
                    "hotplug": "bool",
                    "size": "string"
                  },
                  "memory_quota": {
                    "auto": "string"
                  },
                  "net0": {
                    "card": "string",
                    "enabled": "bool",
                    "mac": "string",
                    "type": "string"
                  },
                  "sound0": {
                    "enabled": "bool",
                    "mixer": "string",
                    "output": "string"
                  },
                  "usb": {
                    "enabled": "bool"
                  },
                  "video": {
                    "3d-acceleration": "string",
                    "adapter-type": "string",
                    "automatic-video-memory": "string",
                    "high-resolution": "string",
                    "high-resolution-in-guest": "string",
                    "native-scaling-in-guest": "string",
                    "size": "string",
                    "vertical-sync": "string"
                  }
                },
                "Home": "string",
                "Home path": "string",
                "Host Shared Folders": "map[string]unknown",
                "Host defined sharing": "string",
                "ID": "string",
                "Miscellaneous Sharing": {
                  "Shared clipboard": "string",
                  "Shared cloud": "string"
                },
                "Modality": {
                  "Capture mouse clicks": "string",
                  "Opacity (percentage)": "int64",
                  "Show on all spaces ": "string",
                  "Stay on top": "string"
                },
                "Mouse and Keyboard": {
                  "Keyboard optimization mode": "string",
                  "Smart mouse optimized for games": "string",
                  "Smooth scrolling": "string",
                  "Sticky mouse": "string"
                },
                "Name": "string",
                "Network": {
                  "Conditioned": "string",
                  "Inbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "Outbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "ipAddresses": [
                    {
                      "ip": "string",
                      "type": "string"
                    }
                  ]
                },
                "OS": "string",
                "Optimization": {
                  "Adaptive hypervisor": "string",
                  "Auto compress virtual disks": "string",
                  "Disabled Windows logo": "string",
                  "Faster virtual machine": "string",
                  "Hypervisor type": "string",
                  "Longer battery life": "string",
                  "Nested virtualization": "string",
                  "PMU virtualization": "string",
                  "Resource quota": "string",
                  "Show battery status": "string"
                },
                "Print Management": {
                  "Show host printer UI": "string",
                  "Synchronize default printer": "string",
                  "Synchronize with host printers": "string"
                },
                "Restore Image": "string",
                "SMBIOS settings": {
                  "BIOS Version": "string",
                  "Board Manufacturer": "string",
                  "System serial number": "string"
                },
                "Security": {
                  "Archived": "string",
                  "Configuration is locked": "string",
                  "Custom password protection": "string",
                  "Encrypted": "string",
                  "Packed": "string",
                  "Protected": "string",
                  "TPM enabled": "string",
                  "TPM type": "string"
                },
                "Shared Applications": {
                  "Bounce dock icon when app flashes": "string",
                  "Guest-to-host apps sharing": "string",
                  "Host-to-guest apps sharing": "string",
                  "Show guest apps folder in Dock": "string",
                  "Show guest notifications": "string",
                  "enabled": "bool"
                },
                "Shared Profile": {
                  "enabled": "bool"
                },
                "Smart Guard": {
                  "enabled": "bool"
                },
                "SmartMount": {
                  "CD/DVD drives": "string",
                  "Network shares": "string",
                  "Removable drives": "string",
                  "enabled": "bool"
                },
                "Startup and Shutdown": {
                  "Autostart": "string",
                  "Autostart delay": "int64",
                  "Autostop": "string",
                  "On shutdown": "string",
                  "On window close": "string",
                  "Pause idle": "string",
                  "Startup view": "string",
                  "Undo disks": "string"
                },
                "State": "string",
                "Template": "string",
                "Time Synchronization": {
                  "Interval (in seconds)": "int64",
                  "Smart mode": "string",
                  "Timezone synchronization disabled": "string",
                  "enabled": "bool"
                },
                "Travel mode": {
                  "Enter condition": "string",
                  "Enter threshold": "int64",
                  "Quit condition": "string"
                },
                "Type": "string",
                "USB and Bluetooth": {
                  "Automatic sharing bluetooth": "string",
                  "Automatic sharing cameras": "string",
                  "Automatic sharing gamepads": "string",
                  "Automatic sharing smart cards": "string",
                  "Support USB 3.0": "string"
                },
                "Uptime": "string",
                "host": "string",
                "host_external_ip_address": "string",
                "host_id": "string",
                "host_state": "string",
                "host_url": "string",
                "internal_ip_address": "string",
                "user": "string"
              }
            ]
          code: "200"
          code_description: OK
          title: '[]models.ParallelsVM'
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
        - code_block: |-
            [
              {
                "Advanced": {
                  "Public SSH keys synchronization": "string",
                  "Rosetta Linux": "string",
                  "Share host location": "string",
                  "Show developer tools": "string",
                  "Swipe from edges": "string",
                  "VM hostname synchronization": "string"
                },
                "Allow select boot device": "string",
                "BIOS type": "string",
                "Boot order": "string",
                "Coherence": {
                  "Auto-switch to full screen": "string",
                  "Disable aero": "string",
                  "Hide minimized windows": "string",
                  "Show Windows systray in Mac menu": "string"
                },
                "Description": "string",
                "EFI Secure boot": "string",
                "Expiration": {
                  "enabled": "bool"
                },
                "External boot device": "string",
                "Fullscreen": {
                  "Activate spaces on click": "string",
                  "Gamma control": "string",
                  "Optimize for games": "string",
                  "Scale view mode": "string",
                  "Use all displays": "string"
                },
                "Guest Shared Folders": {
                  "Automount": "string",
                  "enabled": "bool"
                },
                "GuestTools": {
                  "state": "string",
                  "version": "string"
                },
                "Hardware": {
                  "cdrom0": {
                    "enabled": "bool",
                    "image": "string",
                    "port": "string",
                    "state": "string"
                  },
                  "cpu": {
                    "VT-x": "bool",
                    "accl": "string",
                    "auto": "string",
                    "cpus": "int64",
                    "hotplug": "bool",
                    "mode": "string",
                    "type": "string"
                  },
                  "hdd0": {
                    "enabled": "bool",
                    "image": "string",
                    "online-compact": "string",
                    "port": "string",
                    "size": "string",
                    "type": "string"
                  },
                  "memory": {
                    "auto": "string",
                    "hotplug": "bool",
                    "size": "string"
                  },
                  "memory_quota": {
                    "auto": "string"
                  },
                  "net0": {
                    "card": "string",
                    "enabled": "bool",
                    "mac": "string",
                    "type": "string"
                  },
                  "sound0": {
                    "enabled": "bool",
                    "mixer": "string",
                    "output": "string"
                  },
                  "usb": {
                    "enabled": "bool"
                  },
                  "video": {
                    "3d-acceleration": "string",
                    "adapter-type": "string",
                    "automatic-video-memory": "string",
                    "high-resolution": "string",
                    "high-resolution-in-guest": "string",
                    "native-scaling-in-guest": "string",
                    "size": "string",
                    "vertical-sync": "string"
                  }
                },
                "Home": "string",
                "Home path": "string",
                "Host Shared Folders": "map[string]unknown",
                "Host defined sharing": "string",
                "ID": "string",
                "Miscellaneous Sharing": {
                  "Shared clipboard": "string",
                  "Shared cloud": "string"
                },
                "Modality": {
                  "Capture mouse clicks": "string",
                  "Opacity (percentage)": "int64",
                  "Show on all spaces ": "string",
                  "Stay on top": "string"
                },
                "Mouse and Keyboard": {
                  "Keyboard optimization mode": "string",
                  "Smart mouse optimized for games": "string",
                  "Smooth scrolling": "string",
                  "Sticky mouse": "string"
                },
                "Name": "string",
                "Network": {
                  "Conditioned": "string",
                  "Inbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "Outbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "ipAddresses": [
                    {
                      "ip": "string",
                      "type": "string"
                    }
                  ]
                },
                "OS": "string",
                "Optimization": {
                  "Adaptive hypervisor": "string",
                  "Auto compress virtual disks": "string",
                  "Disabled Windows logo": "string",
                  "Faster virtual machine": "string",
                  "Hypervisor type": "string",
                  "Longer battery life": "string",
                  "Nested virtualization": "string",
                  "PMU virtualization": "string",
                  "Resource quota": "string",
                  "Show battery status": "string"
                },
                "Print Management": {
                  "Show host printer UI": "string",
                  "Synchronize default printer": "string",
                  "Synchronize with host printers": "string"
                },
                "Restore Image": "string",
                "SMBIOS settings": {
                  "BIOS Version": "string",
                  "Board Manufacturer": "string",
                  "System serial number": "string"
                },
                "Security": {
                  "Archived": "string",
                  "Configuration is locked": "string",
                  "Custom password protection": "string",
                  "Encrypted": "string",
                  "Packed": "string",
                  "Protected": "string",
                  "TPM enabled": "string",
                  "TPM type": "string"
                },
                "Shared Applications": {
                  "Bounce dock icon when app flashes": "string",
                  "Guest-to-host apps sharing": "string",
                  "Host-to-guest apps sharing": "string",
                  "Show guest apps folder in Dock": "string",
                  "Show guest notifications": "string",
                  "enabled": "bool"
                },
                "Shared Profile": {
                  "enabled": "bool"
                },
                "Smart Guard": {
                  "enabled": "bool"
                },
                "SmartMount": {
                  "CD/DVD drives": "string",
                  "Network shares": "string",
                  "Removable drives": "string",
                  "enabled": "bool"
                },
                "Startup and Shutdown": {
                  "Autostart": "string",
                  "Autostart delay": "int64",
                  "Autostop": "string",
                  "On shutdown": "string",
                  "On window close": "string",
                  "Pause idle": "string",
                  "Startup view": "string",
                  "Undo disks": "string"
                },
                "State": "string",
                "Template": "string",
                "Time Synchronization": {
                  "Interval (in seconds)": "int64",
                  "Smart mode": "string",
                  "Timezone synchronization disabled": "string",
                  "enabled": "bool"
                },
                "Travel mode": {
                  "Enter condition": "string",
                  "Enter threshold": "int64",
                  "Quit condition": "string"
                },
                "Type": "string",
                "USB and Bluetooth": {
                  "Automatic sharing bluetooth": "string",
                  "Automatic sharing cameras": "string",
                  "Automatic sharing gamepads": "string",
                  "Automatic sharing smart cards": "string",
                  "Support USB 3.0": "string"
                },
                "Uptime": "string",
                "host": "string",
                "host_external_ip_address": "string",
                "host_id": "string",
                "host_state": "string",
                "host_url": "string",
                "internal_ip_address": "string",
                "user": "string"
              }
            ]
          code: "200"
          code_description: OK
          title: models.ParallelsVM
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
        - code_block: |-
            [
              {
                "Advanced": {
                  "Public SSH keys synchronization": "string",
                  "Rosetta Linux": "string",
                  "Share host location": "string",
                  "Show developer tools": "string",
                  "Swipe from edges": "string",
                  "VM hostname synchronization": "string"
                },
                "Allow select boot device": "string",
                "BIOS type": "string",
                "Boot order": "string",
                "Coherence": {
                  "Auto-switch to full screen": "string",
                  "Disable aero": "string",
                  "Hide minimized windows": "string",
                  "Show Windows systray in Mac menu": "string"
                },
                "Description": "string",
                "EFI Secure boot": "string",
                "Expiration": {
                  "enabled": "bool"
                },
                "External boot device": "string",
                "Fullscreen": {
                  "Activate spaces on click": "string",
                  "Gamma control": "string",
                  "Optimize for games": "string",
                  "Scale view mode": "string",
                  "Use all displays": "string"
                },
                "Guest Shared Folders": {
                  "Automount": "string",
                  "enabled": "bool"
                },
                "GuestTools": {
                  "state": "string",
                  "version": "string"
                },
                "Hardware": {
                  "cdrom0": {
                    "enabled": "bool",
                    "image": "string",
                    "port": "string",
                    "state": "string"
                  },
                  "cpu": {
                    "VT-x": "bool",
                    "accl": "string",
                    "auto": "string",
                    "cpus": "int64",
                    "hotplug": "bool",
                    "mode": "string",
                    "type": "string"
                  },
                  "hdd0": {
                    "enabled": "bool",
                    "image": "string",
                    "online-compact": "string",
                    "port": "string",
                    "size": "string",
                    "type": "string"
                  },
                  "memory": {
                    "auto": "string",
                    "hotplug": "bool",
                    "size": "string"
                  },
                  "memory_quota": {
                    "auto": "string"
                  },
                  "net0": {
                    "card": "string",
                    "enabled": "bool",
                    "mac": "string",
                    "type": "string"
                  },
                  "sound0": {
                    "enabled": "bool",
                    "mixer": "string",
                    "output": "string"
                  },
                  "usb": {
                    "enabled": "bool"
                  },
                  "video": {
                    "3d-acceleration": "string",
                    "adapter-type": "string",
                    "automatic-video-memory": "string",
                    "high-resolution": "string",
                    "high-resolution-in-guest": "string",
                    "native-scaling-in-guest": "string",
                    "size": "string",
                    "vertical-sync": "string"
                  }
                },
                "Home": "string",
                "Home path": "string",
                "Host Shared Folders": "map[string]unknown",
                "Host defined sharing": "string",
                "ID": "string",
                "Miscellaneous Sharing": {
                  "Shared clipboard": "string",
                  "Shared cloud": "string"
                },
                "Modality": {
                  "Capture mouse clicks": "string",
                  "Opacity (percentage)": "int64",
                  "Show on all spaces ": "string",
                  "Stay on top": "string"
                },
                "Mouse and Keyboard": {
                  "Keyboard optimization mode": "string",
                  "Smart mouse optimized for games": "string",
                  "Smooth scrolling": "string",
                  "Sticky mouse": "string"
                },
                "Name": "string",
                "Network": {
                  "Conditioned": "string",
                  "Inbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "Outbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "ipAddresses": [
                    {
                      "ip": "string",
                      "type": "string"
                    }
                  ]
                },
                "OS": "string",
                "Optimization": {
                  "Adaptive hypervisor": "string",
                  "Auto compress virtual disks": "string",
                  "Disabled Windows logo": "string",
                  "Faster virtual machine": "string",
                  "Hypervisor type": "string",
                  "Longer battery life": "string",
                  "Nested virtualization": "string",
                  "PMU virtualization": "string",
                  "Resource quota": "string",
                  "Show battery status": "string"
                },
                "Print Management": {
                  "Show host printer UI": "string",
                  "Synchronize default printer": "string",
                  "Synchronize with host printers": "string"
                },
                "Restore Image": "string",
                "SMBIOS settings": {
                  "BIOS Version": "string",
                  "Board Manufacturer": "string",
                  "System serial number": "string"
                },
                "Security": {
                  "Archived": "string",
                  "Configuration is locked": "string",
                  "Custom password protection": "string",
                  "Encrypted": "string",
                  "Packed": "string",
                  "Protected": "string",
                  "TPM enabled": "string",
                  "TPM type": "string"
                },
                "Shared Applications": {
                  "Bounce dock icon when app flashes": "string",
                  "Guest-to-host apps sharing": "string",
                  "Host-to-guest apps sharing": "string",
                  "Show guest apps folder in Dock": "string",
                  "Show guest notifications": "string",
                  "enabled": "bool"
                },
                "Shared Profile": {
                  "enabled": "bool"
                },
                "Smart Guard": {
                  "enabled": "bool"
                },
                "SmartMount": {
                  "CD/DVD drives": "string",
                  "Network shares": "string",
                  "Removable drives": "string",
                  "enabled": "bool"
                },
                "Startup and Shutdown": {
                  "Autostart": "string",
                  "Autostart delay": "int64",
                  "Autostop": "string",
                  "On shutdown": "string",
                  "On window close": "string",
                  "Pause idle": "string",
                  "Startup view": "string",
                  "Undo disks": "string"
                },
                "State": "string",
                "Template": "string",
                "Time Synchronization": {
                  "Interval (in seconds)": "int64",
                  "Smart mode": "string",
                  "Timezone synchronization disabled": "string",
                  "enabled": "bool"
                },
                "Travel mode": {
                  "Enter condition": "string",
                  "Enter threshold": "int64",
                  "Quit condition": "string"
                },
                "Type": "string",
                "USB and Bluetooth": {
                  "Automatic sharing bluetooth": "string",
                  "Automatic sharing cameras": "string",
                  "Automatic sharing gamepads": "string",
                  "Automatic sharing smart cards": "string",
                  "Support USB 3.0": "string"
                },
                "Uptime": "string",
                "host": "string",
                "host_external_ip_address": "string",
                "host_id": "string",
                "host_state": "string",
                "host_url": "string",
                "internal_ip_address": "string",
                "user": "string"
              }
            ]
          code: "200"
          code_description: OK
          title: models.ParallelsVM
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
        - code_block: |-
            [
              {
                "Advanced": {
                  "Public SSH keys synchronization": "string",
                  "Rosetta Linux": "string",
                  "Share host location": "string",
                  "Show developer tools": "string",
                  "Swipe from edges": "string",
                  "VM hostname synchronization": "string"
                },
                "Allow select boot device": "string",
                "BIOS type": "string",
                "Boot order": "string",
                "Coherence": {
                  "Auto-switch to full screen": "string",
                  "Disable aero": "string",
                  "Hide minimized windows": "string",
                  "Show Windows systray in Mac menu": "string"
                },
                "Description": "string",
                "EFI Secure boot": "string",
                "Expiration": {
                  "enabled": "bool"
                },
                "External boot device": "string",
                "Fullscreen": {
                  "Activate spaces on click": "string",
                  "Gamma control": "string",
                  "Optimize for games": "string",
                  "Scale view mode": "string",
                  "Use all displays": "string"
                },
                "Guest Shared Folders": {
                  "Automount": "string",
                  "enabled": "bool"
                },
                "GuestTools": {
                  "state": "string",
                  "version": "string"
                },
                "Hardware": {
                  "cdrom0": {
                    "enabled": "bool",
                    "image": "string",
                    "port": "string",
                    "state": "string"
                  },
                  "cpu": {
                    "VT-x": "bool",
                    "accl": "string",
                    "auto": "string",
                    "cpus": "int64",
                    "hotplug": "bool",
                    "mode": "string",
                    "type": "string"
                  },
                  "hdd0": {
                    "enabled": "bool",
                    "image": "string",
                    "online-compact": "string",
                    "port": "string",
                    "size": "string",
                    "type": "string"
                  },
                  "memory": {
                    "auto": "string",
                    "hotplug": "bool",
                    "size": "string"
                  },
                  "memory_quota": {
                    "auto": "string"
                  },
                  "net0": {
                    "card": "string",
                    "enabled": "bool",
                    "mac": "string",
                    "type": "string"
                  },
                  "sound0": {
                    "enabled": "bool",
                    "mixer": "string",
                    "output": "string"
                  },
                  "usb": {
                    "enabled": "bool"
                  },
                  "video": {
                    "3d-acceleration": "string",
                    "adapter-type": "string",
                    "automatic-video-memory": "string",
                    "high-resolution": "string",
                    "high-resolution-in-guest": "string",
                    "native-scaling-in-guest": "string",
                    "size": "string",
                    "vertical-sync": "string"
                  }
                },
                "Home": "string",
                "Home path": "string",
                "Host Shared Folders": "map[string]unknown",
                "Host defined sharing": "string",
                "ID": "string",
                "Miscellaneous Sharing": {
                  "Shared clipboard": "string",
                  "Shared cloud": "string"
                },
                "Modality": {
                  "Capture mouse clicks": "string",
                  "Opacity (percentage)": "int64",
                  "Show on all spaces ": "string",
                  "Stay on top": "string"
                },
                "Mouse and Keyboard": {
                  "Keyboard optimization mode": "string",
                  "Smart mouse optimized for games": "string",
                  "Smooth scrolling": "string",
                  "Sticky mouse": "string"
                },
                "Name": "string",
                "Network": {
                  "Conditioned": "string",
                  "Inbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "Outbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "ipAddresses": [
                    {
                      "ip": "string",
                      "type": "string"
                    }
                  ]
                },
                "OS": "string",
                "Optimization": {
                  "Adaptive hypervisor": "string",
                  "Auto compress virtual disks": "string",
                  "Disabled Windows logo": "string",
                  "Faster virtual machine": "string",
                  "Hypervisor type": "string",
                  "Longer battery life": "string",
                  "Nested virtualization": "string",
                  "PMU virtualization": "string",
                  "Resource quota": "string",
                  "Show battery status": "string"
                },
                "Print Management": {
                  "Show host printer UI": "string",
                  "Synchronize default printer": "string",
                  "Synchronize with host printers": "string"
                },
                "Restore Image": "string",
                "SMBIOS settings": {
                  "BIOS Version": "string",
                  "Board Manufacturer": "string",
                  "System serial number": "string"
                },
                "Security": {
                  "Archived": "string",
                  "Configuration is locked": "string",
                  "Custom password protection": "string",
                  "Encrypted": "string",
                  "Packed": "string",
                  "Protected": "string",
                  "TPM enabled": "string",
                  "TPM type": "string"
                },
                "Shared Applications": {
                  "Bounce dock icon when app flashes": "string",
                  "Guest-to-host apps sharing": "string",
                  "Host-to-guest apps sharing": "string",
                  "Show guest apps folder in Dock": "string",
                  "Show guest notifications": "string",
                  "enabled": "bool"
                },
                "Shared Profile": {
                  "enabled": "bool"
                },
                "Smart Guard": {
                  "enabled": "bool"
                },
                "SmartMount": {
                  "CD/DVD drives": "string",
                  "Network shares": "string",
                  "Removable drives": "string",
                  "enabled": "bool"
                },
                "Startup and Shutdown": {
                  "Autostart": "string",
                  "Autostart delay": "int64",
                  "Autostop": "string",
                  "On shutdown": "string",
                  "On window close": "string",
                  "Pause idle": "string",
                  "Startup view": "string",
                  "Undo disks": "string"
                },
                "State": "string",
                "Template": "string",
                "Time Synchronization": {
                  "Interval (in seconds)": "int64",
                  "Smart mode": "string",
                  "Timezone synchronization disabled": "string",
                  "enabled": "bool"
                },
                "Travel mode": {
                  "Enter condition": "string",
                  "Enter threshold": "int64",
                  "Quit condition": "string"
                },
                "Type": "string",
                "USB and Bluetooth": {
                  "Automatic sharing bluetooth": "string",
                  "Automatic sharing cameras": "string",
                  "Automatic sharing gamepads": "string",
                  "Automatic sharing smart cards": "string",
                  "Support USB 3.0": "string"
                },
                "Uptime": "string",
                "host": "string",
                "host_external_ip_address": "string",
                "host_id": "string",
                "host_state": "string",
                "host_url": "string",
                "internal_ip_address": "string",
                "user": "string"
              }
            ]
          code: "200"
          code_description: OK
          title: models.ParallelsVM
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
        - code_block: |-
            {
              "operations": [
                {
                  "error": "string",
                  "group": "string",
                  "operation": "string",
                  "status": "string"
                }
              ]
            }
          code: "200"
          code_description: OK
          title: models.VirtualMachineConfigResponse
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
        - code_block: |-
            {
              "operations": [
                {
                  "error": "string",
                  "group": "string",
                  "operation": "string",
                  "status": "string"
                }
              ]
            }
          code: "200"
          code_description: OK
          title: models.VirtualMachineConfigResponse
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
    - title: Starts orchestrator host virtual machine
      description: This endpoint starts orchestrator host virtual machine
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
      response_blocks:
        - code_block: |-
            {
              "operations": [
                {
                  "error": "string",
                  "group": "string",
                  "operation": "string",
                  "status": "string"
                }
              ]
            }
          code: "200"
          code_description: OK
          title: models.VirtualMachineConfigResponse
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
        - code_block: |-
            {
              "operations": [
                {
                  "error": "string",
                  "group": "string",
                  "operation": "string",
                  "status": "string"
                }
              ]
            }
          code: "200"
          code_description: OK
          title: models.VirtualMachineConfigResponse
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
          body: |-
            {
              "delay_applying_restrictions": "bool",
              "force": "bool",
              "machine_name": "string",
              "owner": "string",
              "path": "string",
              "regenerate_source_uuid": "bool",
              "uuid": "string"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "Advanced": {
                  "Public SSH keys synchronization": "string",
                  "Rosetta Linux": "string",
                  "Share host location": "string",
                  "Show developer tools": "string",
                  "Swipe from edges": "string",
                  "VM hostname synchronization": "string"
                },
                "Allow select boot device": "string",
                "BIOS type": "string",
                "Boot order": "string",
                "Coherence": {
                  "Auto-switch to full screen": "string",
                  "Disable aero": "string",
                  "Hide minimized windows": "string",
                  "Show Windows systray in Mac menu": "string"
                },
                "Description": "string",
                "EFI Secure boot": "string",
                "Expiration": {
                  "enabled": "bool"
                },
                "External boot device": "string",
                "Fullscreen": {
                  "Activate spaces on click": "string",
                  "Gamma control": "string",
                  "Optimize for games": "string",
                  "Scale view mode": "string",
                  "Use all displays": "string"
                },
                "Guest Shared Folders": {
                  "Automount": "string",
                  "enabled": "bool"
                },
                "GuestTools": {
                  "state": "string",
                  "version": "string"
                },
                "Hardware": {
                  "cdrom0": {
                    "enabled": "bool",
                    "image": "string",
                    "port": "string",
                    "state": "string"
                  },
                  "cpu": {
                    "VT-x": "bool",
                    "accl": "string",
                    "auto": "string",
                    "cpus": "int64",
                    "hotplug": "bool",
                    "mode": "string",
                    "type": "string"
                  },
                  "hdd0": {
                    "enabled": "bool",
                    "image": "string",
                    "online-compact": "string",
                    "port": "string",
                    "size": "string",
                    "type": "string"
                  },
                  "memory": {
                    "auto": "string",
                    "hotplug": "bool",
                    "size": "string"
                  },
                  "memory_quota": {
                    "auto": "string"
                  },
                  "net0": {
                    "card": "string",
                    "enabled": "bool",
                    "mac": "string",
                    "type": "string"
                  },
                  "sound0": {
                    "enabled": "bool",
                    "mixer": "string",
                    "output": "string"
                  },
                  "usb": {
                    "enabled": "bool"
                  },
                  "video": {
                    "3d-acceleration": "string",
                    "adapter-type": "string",
                    "automatic-video-memory": "string",
                    "high-resolution": "string",
                    "high-resolution-in-guest": "string",
                    "native-scaling-in-guest": "string",
                    "size": "string",
                    "vertical-sync": "string"
                  }
                },
                "Home": "string",
                "Home path": "string",
                "Host Shared Folders": "map[string]unknown",
                "Host defined sharing": "string",
                "ID": "string",
                "Miscellaneous Sharing": {
                  "Shared clipboard": "string",
                  "Shared cloud": "string"
                },
                "Modality": {
                  "Capture mouse clicks": "string",
                  "Opacity (percentage)": "int64",
                  "Show on all spaces ": "string",
                  "Stay on top": "string"
                },
                "Mouse and Keyboard": {
                  "Keyboard optimization mode": "string",
                  "Smart mouse optimized for games": "string",
                  "Smooth scrolling": "string",
                  "Sticky mouse": "string"
                },
                "Name": "string",
                "Network": {
                  "Conditioned": "string",
                  "Inbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "Outbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "ipAddresses": [
                    {
                      "ip": "string",
                      "type": "string"
                    }
                  ]
                },
                "OS": "string",
                "Optimization": {
                  "Adaptive hypervisor": "string",
                  "Auto compress virtual disks": "string",
                  "Disabled Windows logo": "string",
                  "Faster virtual machine": "string",
                  "Hypervisor type": "string",
                  "Longer battery life": "string",
                  "Nested virtualization": "string",
                  "PMU virtualization": "string",
                  "Resource quota": "string",
                  "Show battery status": "string"
                },
                "Print Management": {
                  "Show host printer UI": "string",
                  "Synchronize default printer": "string",
                  "Synchronize with host printers": "string"
                },
                "Restore Image": "string",
                "SMBIOS settings": {
                  "BIOS Version": "string",
                  "Board Manufacturer": "string",
                  "System serial number": "string"
                },
                "Security": {
                  "Archived": "string",
                  "Configuration is locked": "string",
                  "Custom password protection": "string",
                  "Encrypted": "string",
                  "Packed": "string",
                  "Protected": "string",
                  "TPM enabled": "string",
                  "TPM type": "string"
                },
                "Shared Applications": {
                  "Bounce dock icon when app flashes": "string",
                  "Guest-to-host apps sharing": "string",
                  "Host-to-guest apps sharing": "string",
                  "Show guest apps folder in Dock": "string",
                  "Show guest notifications": "string",
                  "enabled": "bool"
                },
                "Shared Profile": {
                  "enabled": "bool"
                },
                "Smart Guard": {
                  "enabled": "bool"
                },
                "SmartMount": {
                  "CD/DVD drives": "string",
                  "Network shares": "string",
                  "Removable drives": "string",
                  "enabled": "bool"
                },
                "Startup and Shutdown": {
                  "Autostart": "string",
                  "Autostart delay": "int64",
                  "Autostop": "string",
                  "On shutdown": "string",
                  "On window close": "string",
                  "Pause idle": "string",
                  "Startup view": "string",
                  "Undo disks": "string"
                },
                "State": "string",
                "Template": "string",
                "Time Synchronization": {
                  "Interval (in seconds)": "int64",
                  "Smart mode": "string",
                  "Timezone synchronization disabled": "string",
                  "enabled": "bool"
                },
                "Travel mode": {
                  "Enter condition": "string",
                  "Enter threshold": "int64",
                  "Quit condition": "string"
                },
                "Type": "string",
                "USB and Bluetooth": {
                  "Automatic sharing bluetooth": "string",
                  "Automatic sharing cameras": "string",
                  "Automatic sharing gamepads": "string",
                  "Automatic sharing smart cards": "string",
                  "Support USB 3.0": "string"
                },
                "Uptime": "string",
                "host": "string",
                "host_external_ip_address": "string",
                "host_id": "string",
                "host_state": "string",
                "host_url": "string",
                "internal_ip_address": "string",
                "user": "string"
              }
            ]
          code: "200"
          code_description: OK
          title: models.ParallelsVM
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/register' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"delay_applying_restrictions\": \"bool\",\n  \"force\": \"bool\",\n  \"machine_name\": \"string\",\n  \"owner\": \"string\",\n  \"path\": \"string\",\n  \"regenerate_source_uuid\": \"bool\",\n  \"uuid\": \"string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/register");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "delay_applying_restrictions": "bool",
              "force": "bool",
              "machine_name": "string",
              "owner": "string",
              "path": "string",
              "regenerate_source_uuid": "bool",
              "uuid": "string"
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/register"
              method := "post"
              payload := strings.NewReader(`{
              "delay_applying_restrictions": "bool",
              "force": "bool",
              "machine_name": "string",
              "owner": "string",
              "path": "string",
              "regenerate_source_uuid": "bool",
              "uuid": "string"
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
          body: |-
            {
              "clean_source_uuid": "bool",
              "id": "string",
              "owner": "string"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "Advanced": {
                  "Public SSH keys synchronization": "string",
                  "Rosetta Linux": "string",
                  "Share host location": "string",
                  "Show developer tools": "string",
                  "Swipe from edges": "string",
                  "VM hostname synchronization": "string"
                },
                "Allow select boot device": "string",
                "BIOS type": "string",
                "Boot order": "string",
                "Coherence": {
                  "Auto-switch to full screen": "string",
                  "Disable aero": "string",
                  "Hide minimized windows": "string",
                  "Show Windows systray in Mac menu": "string"
                },
                "Description": "string",
                "EFI Secure boot": "string",
                "Expiration": {
                  "enabled": "bool"
                },
                "External boot device": "string",
                "Fullscreen": {
                  "Activate spaces on click": "string",
                  "Gamma control": "string",
                  "Optimize for games": "string",
                  "Scale view mode": "string",
                  "Use all displays": "string"
                },
                "Guest Shared Folders": {
                  "Automount": "string",
                  "enabled": "bool"
                },
                "GuestTools": {
                  "state": "string",
                  "version": "string"
                },
                "Hardware": {
                  "cdrom0": {
                    "enabled": "bool",
                    "image": "string",
                    "port": "string",
                    "state": "string"
                  },
                  "cpu": {
                    "VT-x": "bool",
                    "accl": "string",
                    "auto": "string",
                    "cpus": "int64",
                    "hotplug": "bool",
                    "mode": "string",
                    "type": "string"
                  },
                  "hdd0": {
                    "enabled": "bool",
                    "image": "string",
                    "online-compact": "string",
                    "port": "string",
                    "size": "string",
                    "type": "string"
                  },
                  "memory": {
                    "auto": "string",
                    "hotplug": "bool",
                    "size": "string"
                  },
                  "memory_quota": {
                    "auto": "string"
                  },
                  "net0": {
                    "card": "string",
                    "enabled": "bool",
                    "mac": "string",
                    "type": "string"
                  },
                  "sound0": {
                    "enabled": "bool",
                    "mixer": "string",
                    "output": "string"
                  },
                  "usb": {
                    "enabled": "bool"
                  },
                  "video": {
                    "3d-acceleration": "string",
                    "adapter-type": "string",
                    "automatic-video-memory": "string",
                    "high-resolution": "string",
                    "high-resolution-in-guest": "string",
                    "native-scaling-in-guest": "string",
                    "size": "string",
                    "vertical-sync": "string"
                  }
                },
                "Home": "string",
                "Home path": "string",
                "Host Shared Folders": "map[string]unknown",
                "Host defined sharing": "string",
                "ID": "string",
                "Miscellaneous Sharing": {
                  "Shared clipboard": "string",
                  "Shared cloud": "string"
                },
                "Modality": {
                  "Capture mouse clicks": "string",
                  "Opacity (percentage)": "int64",
                  "Show on all spaces ": "string",
                  "Stay on top": "string"
                },
                "Mouse and Keyboard": {
                  "Keyboard optimization mode": "string",
                  "Smart mouse optimized for games": "string",
                  "Smooth scrolling": "string",
                  "Sticky mouse": "string"
                },
                "Name": "string",
                "Network": {
                  "Conditioned": "string",
                  "Inbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "Outbound": {
                    "Bandwidth": "string",
                    "Delay": "string",
                    "Packet Loss": "string"
                  },
                  "ipAddresses": [
                    {
                      "ip": "string",
                      "type": "string"
                    }
                  ]
                },
                "OS": "string",
                "Optimization": {
                  "Adaptive hypervisor": "string",
                  "Auto compress virtual disks": "string",
                  "Disabled Windows logo": "string",
                  "Faster virtual machine": "string",
                  "Hypervisor type": "string",
                  "Longer battery life": "string",
                  "Nested virtualization": "string",
                  "PMU virtualization": "string",
                  "Resource quota": "string",
                  "Show battery status": "string"
                },
                "Print Management": {
                  "Show host printer UI": "string",
                  "Synchronize default printer": "string",
                  "Synchronize with host printers": "string"
                },
                "Restore Image": "string",
                "SMBIOS settings": {
                  "BIOS Version": "string",
                  "Board Manufacturer": "string",
                  "System serial number": "string"
                },
                "Security": {
                  "Archived": "string",
                  "Configuration is locked": "string",
                  "Custom password protection": "string",
                  "Encrypted": "string",
                  "Packed": "string",
                  "Protected": "string",
                  "TPM enabled": "string",
                  "TPM type": "string"
                },
                "Shared Applications": {
                  "Bounce dock icon when app flashes": "string",
                  "Guest-to-host apps sharing": "string",
                  "Host-to-guest apps sharing": "string",
                  "Show guest apps folder in Dock": "string",
                  "Show guest notifications": "string",
                  "enabled": "bool"
                },
                "Shared Profile": {
                  "enabled": "bool"
                },
                "Smart Guard": {
                  "enabled": "bool"
                },
                "SmartMount": {
                  "CD/DVD drives": "string",
                  "Network shares": "string",
                  "Removable drives": "string",
                  "enabled": "bool"
                },
                "Startup and Shutdown": {
                  "Autostart": "string",
                  "Autostart delay": "int64",
                  "Autostop": "string",
                  "On shutdown": "string",
                  "On window close": "string",
                  "Pause idle": "string",
                  "Startup view": "string",
                  "Undo disks": "string"
                },
                "State": "string",
                "Template": "string",
                "Time Synchronization": {
                  "Interval (in seconds)": "int64",
                  "Smart mode": "string",
                  "Timezone synchronization disabled": "string",
                  "enabled": "bool"
                },
                "Travel mode": {
                  "Enter condition": "string",
                  "Enter threshold": "int64",
                  "Quit condition": "string"
                },
                "Type": "string",
                "USB and Bluetooth": {
                  "Automatic sharing bluetooth": "string",
                  "Automatic sharing cameras": "string",
                  "Automatic sharing gamepads": "string",
                  "Automatic sharing smart cards": "string",
                  "Support USB 3.0": "string"
                },
                "Uptime": "string",
                "host": "string",
                "host_external_ip_address": "string",
                "host_id": "string",
                "host_state": "string",
                "host_url": "string",
                "internal_ip_address": "string",
                "user": "string"
              }
            ]
          code: "200"
          code_description: OK
          title: models.ParallelsVM
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/unregister' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"clean_source_uuid\": \"bool\",\n  \"id\": \"string\",\n  \"owner\": \"string\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/unregister");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "clean_source_uuid": "bool",
              "id": "string",
              "owner": "string"
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines/{vmId}/unregister"
              method := "post"
              payload := strings.NewReader(`{
              "clean_source_uuid": "bool",
              "id": "string",
              "owner": "string"
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
          body: |-
            {
              "architecture": "string",
              "catalog_manifest": "*CreateCatalogVirtualMachineRequest",
              "name": "string",
              "owner": "string",
              "packer_template": "*CreatePackerVirtualMachineRequest",
              "start_on_create": "bool",
              "vagrant_box": "*CreateVagrantMachineRequest"
            }
      response_blocks:
        - code_block: |-
            {
              "current_state": "string",
              "host": "string",
              "id": "string",
              "name": "string",
              "owner": "string"
            }
          code: "200"
          code_description: OK
          title: models.CreateVirtualMachineResponse
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/machines' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"architecture\": \"string\",\n  \"catalog_manifest\": \"*CreateCatalogVirtualMachineRequest\",\n  \"name\": \"string\",\n  \"owner\": \"string\",\n  \"packer_template\": \"*CreatePackerVirtualMachineRequest\",\n  \"start_on_create\": \"bool\",\n  \"vagrant_box\": \"*CreateVagrantMachineRequest\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/machines");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "architecture": "string",
              "catalog_manifest": "*CreateCatalogVirtualMachineRequest",
              "name": "string",
              "owner": "string",
              "packer_template": "*CreatePackerVirtualMachineRequest",
              "start_on_create": "bool",
              "vagrant_box": "*CreateVagrantMachineRequest"
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/machines"
              method := "post"
              payload := strings.NewReader(`{
              "architecture": "string",
              "catalog_manifest": "*CreateCatalogVirtualMachineRequest",
              "name": "string",
              "owner": "string",
              "packer_template": "*CreatePackerVirtualMachineRequest",
              "start_on_create": "bool",
              "vagrant_box": "*CreateVagrantMachineRequest"
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
          body: |-
            {
              "architecture": "string",
              "catalog_manifest": "*CreateCatalogVirtualMachineRequest",
              "name": "string",
              "owner": "string",
              "packer_template": "*CreatePackerVirtualMachineRequest",
              "start_on_create": "bool",
              "vagrant_box": "*CreateVagrantMachineRequest"
            }
      response_blocks:
        - code_block: |-
            {
              "current_state": "string",
              "host": "string",
              "id": "string",
              "name": "string",
              "owner": "string"
            }
          code: "200"
          code_description: OK
          title: models.CreateVirtualMachineResponse
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/machines' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"architecture\": \"string\",\n  \"catalog_manifest\": \"*CreateCatalogVirtualMachineRequest\",\n  \"name\": \"string\",\n  \"owner\": \"string\",\n  \"packer_template\": \"*CreatePackerVirtualMachineRequest\",\n  \"start_on_create\": \"bool\",\n  \"vagrant_box\": \"*CreateVagrantMachineRequest\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/machines");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "architecture": "string",
              "catalog_manifest": "*CreateCatalogVirtualMachineRequest",
              "name": "string",
              "owner": "string",
              "packer_template": "*CreatePackerVirtualMachineRequest",
              "start_on_create": "bool",
              "vagrant_box": "*CreateVagrantMachineRequest"
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
              url := "http://localhost/api/v1/orchestrator/machines"
              method := "post"
              payload := strings.NewReader(`{
              "architecture": "string",
              "catalog_manifest": "*CreateCatalogVirtualMachineRequest",
              "name": "string",
              "owner": "string",
              "packer_template": "*CreatePackerVirtualMachineRequest",
              "start_on_create": "bool",
              "vagrant_box": "*CreateVagrantMachineRequest"
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
        - code_block: |-
            {
              "enabled": "bool",
              "host": "string",
              "port": "string"
            }
          code: "200"
          code_description: OK
          title: models.ReverseProxy
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
        - code_block: |-
            [
              {
                "cors": "*ReverseProxyHostCors",
                "host": "string",
                "http_routes": "[]*ReverseProxyHostHttpRoute",
                "id": "string",
                "port": "string",
                "tcp_route": "*ReverseProxyHostTcpRoute",
                "tls": "*ReverseProxyHostTls"
              }
            ]
          code: "200"
          code_description: OK
          title: '[]models.ReverseProxyHost'
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
        - code_block: |-
            [
              {
                "cors": "*ReverseProxyHostCors",
                "host": "string",
                "http_routes": "[]*ReverseProxyHostHttpRoute",
                "id": "string",
                "port": "string",
                "tcp_route": "*ReverseProxyHostTcpRoute",
                "tls": "*ReverseProxyHostTls"
              }
            ]
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
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
          body: |-
            {
              "cors": "*ReverseProxyHostCors",
              "host": "string",
              "http_routes": "[]*ReverseProxyHostHttpRoute",
              "port": "string",
              "tcp_route": "*ReverseProxyHostTcpRoute",
              "tls": "*ReverseProxyHostTls"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "cors": "*ReverseProxyHostCors",
                "host": "string",
                "http_routes": "[]*ReverseProxyHostHttpRoute",
                "id": "string",
                "port": "string",
                "tcp_route": "*ReverseProxyHostTcpRoute",
                "tls": "*ReverseProxyHostTls"
              }
            ]
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"cors\": \"*ReverseProxyHostCors\",\n  \"host\": \"string\",\n  \"http_routes\": \"[]*ReverseProxyHostHttpRoute\",\n  \"port\": \"string\",\n  \"tcp_route\": \"*ReverseProxyHostTcpRoute\",\n  \"tls\": \"*ReverseProxyHostTls\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "cors": "*ReverseProxyHostCors",
              "host": "string",
              "http_routes": "[]*ReverseProxyHostHttpRoute",
              "port": "string",
              "tcp_route": "*ReverseProxyHostTcpRoute",
              "tls": "*ReverseProxyHostTls"
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts"
              method := "post"
              payload := strings.NewReader(`{
              "cors": "*ReverseProxyHostCors",
              "host": "string",
              "http_routes": "[]*ReverseProxyHostHttpRoute",
              "port": "string",
              "tcp_route": "*ReverseProxyHostTcpRoute",
              "tls": "*ReverseProxyHostTls"
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
          body: |-
            {
              "cors": "*ReverseProxyHostCors",
              "host": "string",
              "port": "string",
              "tls": "*ReverseProxyHostTls"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "cors": "*ReverseProxyHostCors",
                "host": "string",
                "http_routes": "[]*ReverseProxyHostHttpRoute",
                "id": "string",
                "port": "string",
                "tcp_route": "*ReverseProxyHostTcpRoute",
                "tls": "*ReverseProxyHostTls"
              }
            ]
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"cors\": \"*ReverseProxyHostCors\",\n  \"host\": \"string\",\n  \"port\": \"string\",\n  \"tls\": \"*ReverseProxyHostTls\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Put, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "cors": "*ReverseProxyHostCors",
              "host": "string",
              "port": "string",
              "tls": "*ReverseProxyHostTls"
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}"
              method := "put"
              payload := strings.NewReader(`{
              "cors": "*ReverseProxyHostCors",
              "host": "string",
              "port": "string",
              "tls": "*ReverseProxyHostTls"
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
          body: |-
            {
              "cors": "*ReverseProxyHostCors",
              "host": "string",
              "port": "string",
              "tls": "*ReverseProxyHostTls"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "cors": "*ReverseProxyHostCors",
                "host": "string",
                "http_routes": "[]*ReverseProxyHostHttpRoute",
                "id": "string",
                "port": "string",
                "tcp_route": "*ReverseProxyHostTcpRoute",
                "tls": "*ReverseProxyHostTls"
              }
            ]
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"cors\": \"*ReverseProxyHostCors\",\n  \"host\": \"string\",\n  \"port\": \"string\",\n  \"tls\": \"*ReverseProxyHostTls\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "cors": "*ReverseProxyHostCors",
              "host": "string",
              "port": "string",
              "tls": "*ReverseProxyHostTls"
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/http_routes"
              method := "post"
              payload := strings.NewReader(`{
              "cors": "*ReverseProxyHostCors",
              "host": "string",
              "port": "string",
              "tls": "*ReverseProxyHostTls"
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
          body: |-
            {
              "cors": "*ReverseProxyHostCors",
              "host": "string",
              "port": "string",
              "tls": "*ReverseProxyHostTls"
            }
      response_blocks:
        - code_block: |-
            [
              {
                "cors": "*ReverseProxyHostCors",
                "host": "string",
                "http_routes": "[]*ReverseProxyHostHttpRoute",
                "id": "string",
                "port": "string",
                "tcp_route": "*ReverseProxyHostTcpRoute",
                "tls": "*ReverseProxyHostTls"
              }
            ]
          code: "200"
          code_description: OK
          title: models.ReverseProxyHost
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
        - code_block: "curl --location 'http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/tcp_route' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"cors\": \"*ReverseProxyHostCors\",\n  \"host\": \"string\",\n  \"port\": \"string\",\n  \"tls\": \"*ReverseProxyHostTls\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/tcp_route");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "cors": "*ReverseProxyHostCors",
              "host": "string",
              "port": "string",
              "tls": "*ReverseProxyHostTls"
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
              url := "http://localhost/api/v1/orchestrator/hosts/{id}/reverse-proxy/hosts/{reverse_proxy_host_id}/tcp_route"
              method := "post"
              payload := strings.NewReader(`{
              "cors": "*ReverseProxyHostCors",
              "host": "string",
              "port": "string",
              "tls": "*ReverseProxyHostTls"
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

---
# Orchestrator endpoints 

 This document contains the endpoints for the Orchestrator category.


