---
layout: api
title: Config
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
    - title: Gets Parallels Desktop active license
      description: This endpoint returns Parallels Desktop active license
      requires_authorization: true
      category: Config
      category_path: config
      path: /v1/parallels_desktop/key
      method: get
      response_blocks:
        - code_block: |-
            {
              "advanced_restrictions": "string",
              "cpu_total": "int64",
              "deferred_activation": "string",
              "edition": "string",
              "expiration": "string",
              "grace_period_ends_at": "string",
              "is_volume": "string",
              "main_period_ends_at": "string",
              "max_memory": "int64",
              "serial": "string",
              "status": "string",
              "uuid": "string"
            }
          code: "200"
          code_description: OK
          title: models.ParallelsDesktopLicense
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "message": "string",
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
          body: |-
            {
              "all": "bool",
              "run_as": "string",
              "tools": "map[string]InstallToolsRequestItem"
            }
      response_blocks:
        - code_block: |-
            {
              "installed_tools": "map[string]InstallToolsResponseItem",
              "success": "bool"
            }
          code: "200"
          code_description: OK
          title: models.InstallToolsResponse
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "message": "string",
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
        - code_block: "curl --location 'http://localhost/api/v1/config/tools/install' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"all\": \"bool\",\n  \"run_as\": \"string\",\n  \"tools\": \"map[string]InstallToolsRequestItem\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/config/tools/install");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "all": "bool",
              "run_as": "string",
              "tools": "map[string]InstallToolsRequestItem"
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
              url := "http://localhost/api/v1/config/tools/install"
              method := "post"
              payload := strings.NewReader(`{
              "all": "bool",
              "run_as": "string",
              "tools": "map[string]InstallToolsRequestItem"
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
          body: |-
            {
              "all": "bool",
              "run_as": "string",
              "tools": "map[string]UninstallToolsRequestItem",
              "uninstall_dependencies": "bool"
            }
      response_blocks:
        - code_block: |-
            {
              "installed_tools": "map[string]InstallToolsResponseItem",
              "success": "bool"
            }
          code: "200"
          code_description: OK
          title: models.InstallToolsResponse
          language: json
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "message": "string",
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
        - code_block: "curl --location 'http://localhost/api/v1/config/tools/uninstall' \n--header 'Authorization ••••••'\n--header 'Content-Type: application/json' \n--data '{\n  \"all\": \"bool\",\n  \"run_as\": \"string\",\n  \"tools\": \"map[string]UninstallToolsRequestItem\",\n  \"uninstall_dependencies\": \"bool\"\n}'\n"
          title: cURL
          language: powershell
        - code_block: |
            var client = new HttpClient();

            var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost/api/v1/config/tools/uninstall");
            request.Headers.Add("Authorization", "••••••");
            request.Headers.Add("Content-Type", "application/json");
            request.Content = new StringContent("{
              "all": "bool",
              "run_as": "string",
              "tools": "map[string]UninstallToolsRequestItem",
              "uninstall_dependencies": "bool"
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
              url := "http://localhost/api/v1/config/tools/uninstall"
              method := "post"
              payload := strings.NewReader(`{
              "all": "bool",
              "run_as": "string",
              "tools": "map[string]UninstallToolsRequestItem",
              "uninstall_dependencies": "bool"
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
    - title: Restarts the API Service
      description: This endpoint restarts the API Service
      requires_authorization: true
      category: Config
      category_path: config
      path: /v1/config/tools/restart
      method: post
      response_blocks:
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "message": "string",
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
        - code_block: |-
            {
              "cpu_brand": "string",
              "cpu_type": "string",
              "devops_version": "string",
              "external_ip_address": "string",
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
                  "message": "string",
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
        - code_block: |-
            {
              "error_message": "string",
              "healthy": "bool",
              "message": "string",
              "name": "string"
            }
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
        - code_block: |-
            {
              "code": "int",
              "message": "string",
              "stack": [
                {
                  "code": "int",
                  "description": "string",
                  "message": "string",
                  "path": "string"
                }
              ]
            }
          code: "402"
          code_description: Payment Required
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
# Config endpoints 

 This document contains the endpoints for the Config category.


