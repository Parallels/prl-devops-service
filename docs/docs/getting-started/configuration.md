---
layout: page
title: Getting Started
subtitle: Configuration
menubar: docs_menu
show_sidebar: false
toc: true
---

The DevOps Service can be configured in various ways, such as using `environment variables`, `command line flags`, or a `configuration file`. The order used to determine the configuration is as follows:

1. Environment Variables
2. Command Line Flags
3. Configuration File

## Configuration File

Using the configuration file is the most flexible way to configure the DevOps Service. It allows you to configure the service in a single file and share it with your team or other instances. Additionally, it allows for a shortened command line when starting the service.

You can create a configuration file either in the current binary folder or anywhere in your system. If you have created it in the current binary folder, the service will look for it and use it as the configuration file. Otherwise, you can specify the path to the configuration file using the `--config` flag.

If you have a configuration file in the current binary folder but pass the `--config` flag, the service will use the configuration file specified in the flag, allowing you to effectively have multiple configurations.

Please find below a simple example of a configuration file, which is a `yaml` object:

```yaml
environment:
  api_port: 5570
  log_level: DEBUG
```

The root object of the configuration file is the environment object, which contains all the environment variables that the service will use. The service will look for the environment object and use the values inside it. Below is a list of all the environment variables that the service will use. These will be the same as the command line flags.

### System

| Flag | Description | Default Value |
| ---- | ----------- | ------------- |
| MODE | This can be either `api` or `orchestrator`, and specifies the mode that the service will run in | api |
| ROOT_PASSWORD | The root password that will be used to update the root password of the virtual machine | |
| DATABASE_FOLDER_ENV_VAR | The folder where the database will be stored | /User/Folder/.prl-devops-service |
| CATALOG_CACHE_FOLDER | The folder where the catalog cache will be stored | /User/Folder/.prl-devops-service/catalog |

### Rest API

| Flag | Description | Default Value |
| ---- | ----------- | ------------- |
| API_PORT | The port that the service will listen on | 8080 |
| API_PREFIX | The prefix that will be used for the api endpoints | /api |
| LOG_LEVEL | The log level of the service | info |
| HMAC_SECRET | The secret that will be used to sign the jwt tokens | |
| ENCRYPTION_PRIVATE_KEY | The private key that will be used to encrypt the database at rest. You can generate one with the `gen-rsa` command | |
| TLS_ENABLED | Specifies whether the service should use tls | false |
| TLS_PORT | The port that the service will listen on for tls | 8443 |
| TLS_CERTIFICATE | A base64 encoded certificate string | |
| TLS_PRIVATE_KEY | A base64 encoded private key string | |
| DISABLE_CATALOG_CACHING | Specifies whether the service should disable the catalog caching | false |
| USE_ORCHESTRATOR_RESOURCES | Specifies whether the service is running in orchestrator mode, which allows the service to use the resources of the orchestrator | false |
| ORCHESTRATOR_PULL_FREQUENCY_SECONDS | The frequency in seconds that the orchestrator will sync with the other hosts in seconds | 30 |
| CORS_ALLOWED_HEADERS | The headers that are allowed in the cors policy | "X-Requested-With, authorization, content-type" |
| CORS_ALLOWED_ORIGINS | The origins that are allowed in the cors policy | "*" |
| CORS_ALLOWED_METHODS | The methods that are allowed in the cors policy | "GET, HEAD, POST, PUT, DELETE, OPTIONS" |
| ENABLE_PACKER_PLUGIN | Specifies whether the service should enable the packer plugin | false |
| ENABLE_VAGRANT_PLUGIN | Specifies whether the service should enable the vagrant plugin | false |

### Json Web Tokens

| Flag | Description | Default Value |
| ---- | ----------- | ------------- |
| JWT_SIGN_ALGORITHM | The algorithm that will be used to sign the jwt tokens. This can be either `HS256`, `RS256`, `HS384`, `RS384`, `HS512`, `RS512` | HS256 |
| JWT_PRIVATE_KEY | The private key that will be used to sign the jwt tokens. This is only required if you are using `RS256`, `RS384` or `RS512` | |
| JWT_HMACS_SECRET | The secret that will be used to sign the jwt tokens. This is only required if you are using `HS256`, `HS384` or `HS512`. Defaults to random | 
| JWT_DURATION | The duration that the jwt token will be valid for. You can use the following format, for example, 5 minutes would be `5m` or 1 hour would be `1h` | 15m |

### Password Complexity

| Flag | Description | Default Value |
| ---- | ----------- | ------------- |
| SECURITY_PASSWORD_MIN_PASSWORD_LENGTH | The minimum length that the password should be. The minimum is 8 | 12 |
| SECURITY_PASSWORD_MAX_PASSWORD_LENGTH | The maximum length that the password should be. The maximum is 40 | 40 |
| SECURITY_PASSWORD_REQUIRE_UPPERCASE | Specifies whether the password should require at least one uppercase character | true |
| SECURITY_PASSWORD_REQUIRE_LOWERCASE | Specifies whether the password should require at least one lowercase character | true |
| SECURITY_PASSWORD_REQUIRE_NUMBER | Specifies whether the password should require at least one number | true |
| SECURITY_PASSWORD_REQUIRE_SPECIAL_CHAR | Specifies whether the password should require at least one special character | true |
| SECURITY_PASSWORD_SALT_PASSWORD | Specifies whether the password should be salted | true |

### Brute Force Protection

| Flag | Description | Default Value |
| ---- | ----------- | ------------- |
| BRUTE_FORCE_MAX_LOGIN_ATTEMPTS | The maximum number of login attempts before the account is locked | 5 |
| BRUTE_FORCE_LOCKOUT_DURATION | The duration that the account will be locked for. You can use the following format, for example, 5 minutes would be `5m` or 1 hour would be `1h` | 5s |
| BRUTE_FORCE_INCREMENTAL_WAIT | Specifies whether the wait period should be incremental. If set to false, the wait period will be the same for each failed attempt | true |