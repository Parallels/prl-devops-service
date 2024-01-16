# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
