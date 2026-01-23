# Changelog

All notable changes to this project will be documented in this file.
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.10.0] - 2026-01-23

- Enhance GetSizeByteFromString to support additional units and improve error handling
- Fixes #343 
- Added the missing property for the path where we want to clone
- Fixed an issue where operations would fail if there was spaces in the vm name
- Resolve VM registration and retrieval issues
- Fix machine name being incorrectly set to empty during registration
- Fix race condition where cached VMs are not updated immediately after registration 
- Added a retry mech to fetch cache vm if it fails 
- switch to manual VM fetching to ensure data consistency after registration
- Work around prlctl limitation where VM info is incomplete when ID is not specified in command arguments
- Fixes # (issue)
- Feature: Added HasWebsocketEvents field to the Orchestrator Host API,providing clients with a reliable source of truth for real-time event capabilities.
- Improvement: Optimized database performance by implementing lightweight status updates for WebSocket events, avoiding full record rewrites.
- Fix: Enhanced connection reliability by tying the "Connected" status strictly to verified pong responses rather than just TCP connection establishment.
- Fix: Resolved delay in reconnection events by triggering an immediate ping upon connection, ensuring instant status updates.
- Improvement: Refactored connection monitoring to automatically detect and handle stale WebSocket connections without redundant network traffic.
- bugfix: Associate user with added VM in processVmAdded function
- docs: add Events Hub user guides and reference documentation
- Introduces a new documentation section for the Events Hub WebSocket service in `docs/docs/devops/events-hub/`.
- Add `overview.md`: Covers WebSocket protocol details, authentication (Bearer/API Key), connection limits, and available channels (`global`, `pdfm`, `orchestrator`, `health`).
- Add "How-to" guides:
- `how-to-monitor-VM-events.md`: Subscribing to VM lifecycle events with JSON payload examples.
- `how-to-monitor-orchestrator-events.md`: Monitoring cluster-level events like host health and VM state changes.
- `how-to-send-heartbeat.md`: Implementing ping/pong logic to keep connections alive.
- `how-to-send-unsubscribe-requests.md` & `how-to-know-my-client-id.md`: Managing subscriptions and client identity.
- Visualizations: Integrate Mermaid.js todocs: add Events Hub user guides and reference documentation
- Closes #332 
- Add `event-load` test command to spawn a dedicated load test server
- Implement mock auth and realistic PDFM VM event broadcasting (100 events/sec)
- Register HealthService to support bidirectional ping/pong health checks
- Add k6 script ([loadtest/hub_load.js](loadtest/hub_load.js) for capacity testing:
- Subscribes to both `pdfm` and `health` events
- Tracks latency, message loss, and throughput
- Verifies bidirectional communication via periodic pings
- Closes #333 
- New Event Type: orchestrator - Subscribe via /v1/ws/subscribe?event_types=orchestrator
- Real-time Host Health Updates: Receive notifications when orchestrator checks host health
- Real-time VM State Updates: Get instant notifications for VM lifecycle events across all managed hosts
- closes #321 
- **Real-time Updates:** Eliminates polling delay for hosts with WebSocket support
- **Reduced Load:** Significantly decreases API calls for event monitoring
- **Backward Compatible:** Gracefully falls back to polling for legacy hosts
- **Scalable:** Efficient connection pooling and resource management
- **Resilient:** Automatic reconnection and error recovery mechanisms
- Part of #321 
- Loading the cfg file properly
- Corrected log messages
- HealthService: Implemented a singleton HealthService to handle health events (e.g., ping -> pong) using a decoupled Broadcaster interface.
- SystemHandler: Implemented a singleton SystemHandler to handle system events (e.g., client-id) using a decoupled Broadcaster interface.
- Optimized Routing: Refactored handleClientMessage to avoid double-marshalling. It now performs partial parsing of the header and passes the raw []byte payload directly to handlers via RouteMessageCmd.
- Type Safety: Enforced usage of constants.EventType throughout the system
- Testing: Added comprehensive unit tests for HealthService, SystemHandler, and Client (Reader/Writer/Handler), significantly improving code coverage.
- closes #319 
- Now all parallels engine events are posted to event's emitter
- User's can now subscribe to PDfM type events and start listening to real time events through a websocket connection.
- Now Parallels service will broadcast  VM_REMOVED,VM_ADDED and VM_STATE_CHANGED to events emitter
- closes #320 
- After the new event emitter we had a racing condition during the catalog pull process where after registering we would check immediately the presence of a VM but this would not be there due to the event not being fired on time
- Added a new function to get VMs that gets it sync
- Updated dependencies on the cache function to this method
- Created WebSocket connection handler with support for multiple event type - subscriptions.
- Introduce a type‑safe EventType to replace string literals and prevent processing errors.
- Added client read/write functions (clientReader/Clietnwriter) for bidirectional WebSocket - communication.
- Implemented connection limiting per IP address with debug/release build variants.
- Added events controller with WebSocket subscription and unsubscription endpoints.
- Added EventEmitter configuration for ping interval and pong timeout settings.
- Integrated EventEmitter service initialization during application startup.
- Achieved 57% overall code coverage for the eventEmitter package.
- related to #319 
- Implemented EventEmitter for managing WebSocket event broadcasting.
- Created a Hub for handling client connections and message broadcasting.
- Added Client struct to represent connected WebSocket clients.
- Introduced methods for sending messages to specific clients, all clients, or by event type.
- Implemented graceful shutdown for the EventEmitter service.
- Added tests for environment isolation, EventEmitter functionality, and Hub operations.
- Included helper functions for setting up test environments and clients.
- Ensured proper cleanup of environment variables and client connections during tests.
- related to #319 
- Implemented a new Event Monitor function in the current DevOps Parallels Desktop service.
- This will read the output of the Event Monitor from prlctl and update the cached list of VMs depending on the state.
- Now the whole `ParallelsService` uses event emitter to get real-time status of Vm's 
- Now in API/Orchestrator will have the real-time VM status
- PARALLELS_DESKTOP_REFRESH_INTERVAL config is deprecated, vm status is always real-time 
- Fixes [315](https://github.com/Parallels/prl-devops-service/issues/315)
- Fixes [317](https://github.com/Parallels/prl-devops-service/issues/317)
- Removed alludo branding from the documentation
- Added the new Clarity cookie consent form to the page
- docs(catalog): add MinIO storage provider details and update push/pull examples with links
- Updated GitHub links and enhanced folder structure descriptions in `source_code` page
- Added missing configs descriptions for pd files
- Updated configuration documentation with new settings for catalog and reverse proxy
-  Added the ability to disable catalog api credentials obfuscation using env variables (backward compatibility with older clients)
- Introduced new endpoints for retrieving system logs and streaming logs via WebSocket.
- Added endpoint to fetch orchestrator host system logs.
- Implemented endpoint for retrieving orchestrator host reverse proxy configuration.
- Updated error response structure across various endpoints to use "error" instead of "message".
- Enhanced API documentation to reflect the new endpoints and changes in response formats.
- Added support for compressing manifest files during the generation process.
- Introduced new fields in the PushCatalogManifestRequest model to handle compression options.
- Updated the manifest generation logic to calculate original and packed sizes, and log compression details.
- Added command processors for handling new PDFile commands: COMPRESS_PACK, FORCE, IS_COMPRESSED, VM_REMOTE_PATH, VM_SIZE, and VM_TYPE.
- Implemented the runImportVM function to handle the import of VMs with the new parameters.
- Enhanced validation logic to ensure proper handling of new command arguments.
- Refactored the logic for extracting API documentation from Go files to improve clarity and maintainability.
- Added module information retrieval to enhance the handling of import paths in the documentation process.
- Update documentation links to reflect new directory structure for security and catalog sections
- the /v1/auth/users API response will include a detailed ApiErrorDiagnosticsResponse object. This diagnostics object provides a full call stack plus categorized errors and warnings to help developers identify the cause and the failing component for any request failures.
- Fixes # (issue)
- #245 

## [0.9.14] - 2025-09-30

- Fix an issue where decompressing could cause a concurrency issue
- Please include a summary of the change and which issue is fixed. Please also include relevant motivation and context. List any dependencies that are required for this change.
- Fixes # (issue)
- Fixed the dead clicks in the documentation
- License update
- Added a new catalog storage provider for the MinIO an example of a connection string would be (provider=minio;endpoint=http://1localhost:9000;bucket=example;access_key=minioadmin;secret_key=something_secret;use_ssl=false;ignore_cert=true)

## [0.9.13] - 2025-04-29

- Promoted some unstable code to stable
- Added a chunk manager for faster downloads
- Added the new prlcopy script if PD 20.2.0 or above, system will use 'prlcopy' binary else, system will try to compress and upload the file using exec command, Fixes #227 
- Added a better way to manage the blocks
- Incremented the bufferSize for decompression to 40mb
- Fixed a documentation issue
- Added new canary release
- Fixed some extra bugs
- Added a new method for the streaming for testing
- Added an unstable stream for was to test speed on aws
- Please include a summary of the change and which issue is fixed. Please also include relevant motivation and context. List any dependencies that are required for this change.
- Fixes # (issue)
- Fixed some bugs with log streaming
- Fixed an issue where we would not be able to  get the version after deploy
- Fixed an issue where we would get the logs wrong in the container
- Added a new function to create log files
- Updated the helm chart with the logs variables
- implemented the new log channel streamer for the WebSocket to avoid opening the file
- Added a clean option to the beta deployment
- Allow the ability to disable the tls certificate if the environment variable is set
- Added the host logs endpoint to the orchestrator
- Added the host stream logs endpoint to the orchestrator
- Added a way to use a debug version when  running debugger
- Fixed the detection for the default log location
- Fixed an issue where we would not be able to detect the file location for the logs and therefor not stream it
- added an error message if file logger is disabled
- enabled file logger by default
- Added an endpoint to return the log file
- Added an endpoint to stream the logs using websockets
- Fixed an issue where if we would set the log to file from the config.yaml it would not read it at startup
- Added a new script to add the correct changelog to the beta versions
- Added an option to keep old beta versions if needed
- Added a default version to the beta pipeline
- Added a caching integrity check for all cache catalog
- Added the new sign method for the pipeline
- Added the ability to sign the files locally
- Fixed an issue where containers images were not providing telemetry
- Updated some of the steps to build 
- Fixed the issue with the windows build
- Added a strategy to test all build targets during PR
- Fixed an issue where the decompressor would failed to deal with symlink that did not exist
- Fixed an issue where the stream process would not catch all of the errors from the decompressor
- Moved unused code to another file until we release the version
- Added an environment variable to set the path for the log file `PRL_DEVOPS_LOG_FILE_PATH`
- Added an environment variable to disable the streaming process for the remote storage provider `DISABLE_CATALOG_PROVIDER_STREAMING`

## [0.9.12] - 2024-11-13

- Fix timeout for the apis and set the default to 5 hours
- Added more information to the orchestrator hosts endpoint
- Added a fix for the push copy file
- Added a fix in the uninstall of the root service to also clear the database file

## [0.9.11] - 2024-11-12

- Improved orchestrator timeouts
- Added system-reserved data to the orchestrator

## [0.9.10] - 2024-11-11

- Fixed an issue with the CPU threshold that was added twice to the orchestrator
- Fixed an issue with the CPU count as it was not taking into consideration we already had reserved cpus
- Added a higher timeout to orchestrator hosts health checks
- added missing timeout in some of the orchestrator api calls
- Adds the support for an orchestrator to query the hardware endpoint of a host returning the raw data to use with terraform
- Fixed a bug on the execute endpoint where it would give an error if the command was present but the script was not
- Fixed an issue where the orchestrator would not take into account the 2 vms limitation when gathering the necessary space
- fixes #229
- Added the new reverse proxy to allow advanced port forwarding to the target vm
- Added the same endpoints to be controlled by the orchestrator
- Added better hardware info where we added more information about the system
- Fixed an issue with the logs being showed where they should not be
- resolves [PBI] Enable reverse proxy for port forwarding #226
- Added the External IP Address to every virtual machine if known
- Added the Internal IP Address to every virtual machine if running
- Changed the docker image from scratch to alpine to allow for cpu and external ip address
- Added the external IP Address to the config hardware endpoint
- Added the OS name to the config hardware endpoint
- Added the OS version if available to the config hardware endpoint
- Added the new hardware fields to the orchestrator hosts endpointFixes # (issue)
- resolves #224

## [0.9.9] - 2024-10-29

- Fixed an issue where in some endpoints, if the body was wrong, we would receive two json body instead of the correct one
- Added environment variables to the execute endpoint
- Fixed some issue with the pipeline
- Bumps the bundler group with 1 update in the /docs directory: [rexml](https://github.com/ruby/rexml).
- Fixed an issue where the name of the catalog would be wrong when importing vms

## [0.9.8] - 2024-10-17

- Fixed an issue where we would not be able to compile for windows
- Fixed an issue where the clone copy would not always be used
- Fixed an issue where the telemetry could generate a nil pointer issue
- Fixed an issue where the pull file would still need the provider even if not needed
- Fixed an issue where the db was being saved more often that needed

## [0.9.7] - 2024-10-16

- Fixed an issue where the copy would fail if the source and destination were in a removable disk
- Fixed an issue where the metadata would contain the provider credentials in the file
- Fixed an issue where if the vm was very big it could timeout during cp process
- Fixed an issue where running prldevops command in the same machine as running the api would corrupt the database
- Added a new endpoint to update a catalog manifest provider connection string
- Added the ability to use environment variables in the pdfile
- Added the ability to import vms from a remote storage that are not in pdpack format
- Added the ability to auto recover the database from a backup if it is corrupted
- Added a controller to delete the catalog cache
- Added a controller to delete the catalog cache for a specific manifest
- Added a controller to get all the existing catalog cache
- Added a retry to the load of the database to account for slow mount devices
- Please include a summary of the change and which issue is fixed. Please also include relevant motivation and context. List any dependencies that are required for this change.
- Fixes # (issue)
- Fixed a missing image in the catalog documentation

## [0.9.6] - 2024-09-19

### Fixed

- Added the ability to disable the tls validation
  
## [0.9.5] - 2024-09-16

### Fixed

- Fixed the service install plist

## [0.9.4] - 2024-09-09

### Fixed

- Fixed an issue where some pushes would not update the manifest in the catalog

## [0.9.3] - 2024-09-09

### Fixed

- Removed an extra debug line that existed in the code
- Added minimum requirements to the pdfile and catalog db
- Fixed an issue with the pull command that would break if you had more than one machine starting with the same name
- Add `envFrom` key to Helm values

## [0.9.2] - 2024-09-06

### Fixed

- Added a progress message to show how much did we download/upload when using the command line

## [0.9.1] - 2024-09-05

### Fixed

- Updated telemetry event names to use same convention

## [0.9.0] - 2024-09-05

### Fixed

- Fixed an issue with the telemetry in pdfile

## [0.8.8] - 2024-08-23

### Fixed

- Updated the helm chart to allow users to add nginx ingress
- Updated the helm chart to allow multiple storage

## [0.8.7] - 2024-08-20

### Added

- Added client to the pull request from the catalog

## [0.8.6] - 2024-07-11

### Fixed

- Issue when trying to update an orchestrator host that is offline

## [0.8.5] - 2024-07-09

### Fixed

- Issue where the hardware info would crash if only one VM was present

## [0.8.4] - 2024-06-19

### Fixed

- Issue with the workflow that would not change the internal version

## [0.8.3] - 2024-06-19

### Fixed

- Issue that caused a crash when the user was querying the hardware endpoint

## [0.8.2] - 2024-06-18

### Fixed

- Issue that prevented the push of a manifest if the metadata already existed in
  the catalog
- Issue where the delete of a catalog would fail if the ID was used
- Issue with documentation links in the index.md for discord

### Changed

- Added version to all telemetry events
- Added more information to the hardware endpoint

## [0.8.1] - 2024-06-17

### Fixed

- Generation of the Helm Charts
- Issue where some missing details in telemetry would crash

### Changed

- Added better readme documentation
- Updated some documentation

## [0.8.0] - 2024-06-03

### Added

- Added caching mechanism for  local vms, this allows a better management of the
  local vms and reduces the amount of queries to the database, it does have the side
  effect of not showing the vms in real time, but it will show them by default
  every 30s you can change this by setting the environment variable `PARALLELS_DESKTOP_REFRESH_INTERVAL`
- Added a new endpoint to update a host in a orchestrator `[PUT] /api/v1/orchestrator/hosts/{id}`
- Added a backup system for the database, this will backup the database every 2h
  and will keep the last 10 backups by default, this can be changed by setting the
  environment variable `DATABASE_NUMBER_BACKUP_FILES` to the number of
  backups you want to keep and `DATABASE_BACKUP_INTERVAL_MINUTES` to the interval
  you want to backup the database
- Moved the memory database to be saved to a file every 5 minutes, this will allow
  for a better management of the memory database and will allow for a better
  recovery in case of a crash, it will also save on exit or on crash, you can change
  the interval by setting the environment variable `DATABASE_SAVE_INTERVAL_MINUTES`
- Added K6 scripts to perform load tests on the API
- Added a favicon to the documentation

### Fixed

- Fixed an issue where we were trying to get the virtual machines for other users
  when not being a super admin
- Fixed some memory leaks in the orchestrator
- Fixed some other small issues
- Fixed some issues that would report a missing vm immediately after
  creating it

### Changed

- Moved database saving process to a 5 minutes setting to avoid overloading the
  database with too many requests
- Changed the way the orchestrator was checking for VMs status changes to avoid
  overloading the database
- Moved all the old commands to the new exec with context to enable timeouts
- Added a 30 seconds timeout when checking the status of the local vms

## [0.7.1] - 2024-05-29

### Added

- Adding telemetry fix and encryption of PII

## [0.7.0] - 2024-05-18

### Fixed

- Fixed an issue where the user was not being set as super admin even if the
  flag was set to true
- Fixed an issue where the user was not being able to be updated
- Fixed an issue with the service plist

## [0.6.9] - 2024-05-17

### Fixed

- Fixed wrong images being displayed in the documentations

## [0.6.8] - 2024-05-16

### Fixed

- Fixed a issue with the copy system where it would not copy the files correctly

## [0.6.7] - 2024-05-16

### Fixed

- Fixed a issue with the orchestrator where it didn't start the auto refresh

## [0.6.6] - 2024-05-16

### Fixed

- Fixed a issue with the orchestrator where it would delete a vm but reported failed
- Fixed an issue with the orchestrator where sometimes it could report back
  a 500 server error
- Fixed an issue with the copy command that would take a long time to copy

## [0.6.5] - 2024-05-15

### Fixed

- Fixed an issue with telemetry where the user_id was not being sent correctly

## [0.6.4] - 2024-05-15

### Added

- Added a new unified install script to be used  in mac/linux
- Added ability to add/remove tags from catalog manifests
- Added ability to add/remove roles from catalog manifests
- Added ability to add/remove claims from catalog manifests

### Fixed

- Fixed an issue were the executable would not read the configuration file
  correctly if it was on path
- Fixed an issue were we could not add two hosts with the same url and different
  ports
- Improved our helm chart

## [0.6.3] - 2024-05-14

### Changed

- Changed the examples documentation to add more details
- Changed the documentation to have vscode on it

### Fixed

- Fixed an issue were the orchestrator did not allow execute commands from container
- Fixed an issue with the Docker file where it was building for the wrong platform
- Fixed some typos in the documentation

## [0.6.2] - 2024-05-13

### Added

- Added a script to install the service in macOS

### Fixed

- Fixed several issues with the orchestrator
- Fixed an issue where the push command would not read the PROVIDER connection
  string correctly
- Fixed an issue with concurrency configuration saving that could lead to old
  results being saved

## [0.6.1] - 2024-05-02

### Added

- Added some extra commands to the command line interface
- Improved documentation on GitHub Actions and Orchestrator use cases
- Added Start/Stop endpoints to the orchestrator
- Added Amplitude Key to the docker images

### Fixed

- Fixed several issues with the orchestrator
- Fixed and issue with the pull catalog where it could hang

### Changed

- Improved caching methodology to reduce waiting time

## [0.6.0] - 2024-03-26

### Added

- Added the machine id to the output of a pull request in the catalog
- Added the ability do do a catalog pull request without the need to specify the
  local machine path, this will be taken from the user configuration in PD
- Added a spinner to the long running commands for pull and push to notify the
  user that the command is still running
- Added a new endpoint to easily clone a virtual machine `/api/v1/machines/{id}/clone`
- Added the ability to **enable** and **disable** a host from the orchestrator
- Added the ability to configure the API **CORS policies** by passing environment variables

### Fixed

- Fixed an typo in the docker-compose file that would not allow the root password
  to be updated
- Fixed an issue in the pull from the catalog where if there was an error the
  system would crash
- Fixed an issue where the provider would not take into account the host with a
  schema present
- Fixed a bug where the system would crash with a waiting group being negative
- Fixed a bug where queries could get stuck while saving to the database
- Fixed an issue where some credentials would be left behind in temporary files
- Further security fixes to the codebase

### Changed

- Packer Templates and Vagrant box endpoints are now disabled by default due to security
  concerns on remote execution of code, you can enable them by setting the environment
  variable `ENABLE_PACKER_PLUGIN` and `ENABLE_VAGRANT_PLUGIN` to `true`

## [0.5.6] - 2024-03-14

### Added

- Improved logging
- Addressed the gosec security findings
- Simple Reverse proxy
- Adding templates for the issues and bugs
- Add new license agreement to readme
- project rename for prldevops
- Implement amplitude telemetry
- fixing installation filename
- Improved user_id to fix issue with sending telemetry
- Improving the orchestrator host management
- Adding the ability to execute and import in pdfile
- initial github pages revamp
- Fix pull cmd by
- further fixes to the documentations

### Changed

- improved user_id to fix issue with sending telemetry

## [0.4.6] - 2024-01-20

### Fixed

- Fixed an issue where the api base context was not setting the log correctly
  resulting in missing log lines

## [0.4.5] - 2024-01-16

### Added

- added the ability to have a config file for the apiclient, this will help
  users to configure the api with more ease and also will allow them to share
  that same configuration. It will either look for a config file in the current
  directory with the following rules, `config.json`, `config.yaml`, `config.yml`
  you can also add the `.local` before the extension to make it a local config
  you can also specify the path to the config file using the `--config` flag

### Fixed

- fixed a bug where if the JWT token was invalid or empty the client would reset
  the connection without a proper error handling
- fixed a bug if the user would setup the current instance to also be part of the
  orchestrator and the API key would change, then the orchestrator would not be
  able to authenticate

### Changed

- The system will now use the config class to read all of the configuration, this
  will allow us to have a more consistent way of reading the configuration and where
  to search for those values, this allows for example for a parameter to be set in
  either a environment variable, a config file or a command line flag, the order
  of precedence is the following, command line flag, environment variable, config
  file
- updated documentation to reflect the changes in the configuration

## [0.4.4] - 2024-01-12

### Added

- brute force attack protection, this will lock accounts after x attempts, by
  default 5 attempts and will use by default incremental wait periods for each
  failed attempts, all of these parameters can be changed
- added the ability to sign a token with different algorithms, by default it will
  use HS256, but you can change it to RS256, HS384, RS384, HS512, RS512, this will
  cater for the request we had for asymmetric keys
- added a random secret generator for the default HS256 is none is provided, this
  is a change from previous versions where we used the machine id as the secret
  this will increment the security of the default installation
- added a password complexity pipeline for checking if the users passwords adhere
  to the complexity requirements, this can be disabled if required, by default the
  password complexity is enabled and the complexity is set to 12 characters, at least
  one uppercase, one lowercase, one number and one special character
- added a diagnostics class to better cater for errors and exceptions, this will
  allow us to better handle errors and exceptions and return a more meaningful
  error message to the user a the moment is not used in all of the code, but we
  will be adding it to all of the code in the future

### Changed

- added back the ability to hash passwords using the SHA256 algorithm, this was
  removed in a previous version, but we have added it back as some users already
  had passwords hashed using this algorithm and this was breaking them. the default
  installation will use the bcrypt algorithm

### Fixed

- fixed an issue where the token validation endpoint was not working and only accepted
  GET requests, it now accepts only POST requests as expected and documented

## [0.4.3] - 2024-01-09

### Added

- added parallels calls when checking the host's health
- added the ability for the apiclient to have a timeout

### Fixed

- fixed a bug where a host would not show it status correctly

## [Unreleased]

### Added

- Initial project setup
