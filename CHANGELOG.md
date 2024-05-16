# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [unreleased]

### Added

### Fixed

### Changed

### Deprecated

### Removed

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
