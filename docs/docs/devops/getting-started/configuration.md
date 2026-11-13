---
layout: page
title: Getting Started
subtitle: Configuration
menubar: docs_devops_menu
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

| Flag                                            | Description                                                                                                                                   | Default Value                                                                                     |
| ----------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| MODE                                            | This can be either `api` or `orchestrator`, and specifies the mode that the service will run in                                               | api                                                                                               |
| ROOT_PASSWORD                                   | The root password that will be used to update the root password of the virtual machine                                                        |                                                                                                   |
| DATABASE_FOLDER                                 | The folder where the database will be stored                                                                                                  | /User/Folder/.prl-devops-service                                                                  |
| DATABASE_NUMBER_BACKUP_FILES                    | The number of backup files that the database will keep                                                                                        | 10                                                                                                |
| DATABASE_BACKUP_INTERVAL_MINUTES                | The interval in minutes that the database will be backed up in minutes                                                                        | 120 minutes                                                                                       |
| DATABASE_SAVE_INTERVAL_MINUTES                  | The interval in minutes that the database will be saved in minutes                                                                            | 5 minutes                                                                                         |
| CATALOG_CACHE_FOLDER                            | The folder where the catalog cache will be stored                                                                                             | /User/Folder/.prl-devops-service/catalog                                                          |
| CATALOG_COMPRESS_VM                             | Specifies whether the virtual machines in the catalog should be compressed                                                                    | false                                                                                             |
| CATALOG_COMPRESS_VM_RATIO                       | The ratio that will be used to determine whether the virtual machine should be compressed best_speed/balanced/best_compression/no_compression | best_compression                                                                                  |
| CATALOG_ENABLE_PROVIDER_CREDENTIALS_OBFUSCATION | Specifies whether the provider credentials in the catalog should be obfuscated                                                                | true                                                                                              |
| VIRTUAL_MACHINES_FOLDER                         | The folder where the virtual machines will be stored                                                                                          | users/`<username>`/Parallels                                                                      |
| SYSTEM_RESERVED_CPU                             | The number of cpu cores that will be reserved for the system and not used for Orchestrator                                                    | 1                                                                                                 |
| SYSTEM_RESERVED_MEMORY                          | The amount of memory that will be reserved for the system and not used for Orchestrator in Mb's                                               | 2048                                                                                              |
| SYSTEM_RESERVED_DISK                            | The amount of disk space that will be reserved for the system and not used for Orchestrator in Mb's                                           | 20000                                                                                             |
| MAX_VMS_PER_HOST                                | Hard maximum number of virtual machines this host may host, applied even when CPU, memory and disk are still free. `0` means unlimited          | 0                                                                                                 |
| MAX_VMS_PER_HOST_MODE                           | Which virtual machines the limit counts: `any` (every VM registered on the host, stopped ones included) or `running` (VMs occupying a slot)      | any                                                                                               |
| SYSTEM_AUTO_RECOVER_DATABASE                    | Specifies whether the system should auto recover the database if it is corrupted                                                              | true                                                                                              |
| ENABLE_REVERSE_PROXY                            | Specifies whether the reverse proxy should be enabled                                                                                         | false                                                                                             |
| REVERSE_PROXY_PORT                              | The port that the reverse proxy will listen on                                                                                                | 5080                                                                                              |
| REVERSE_PROXY_HOST                              | The host that the reverse proxy will listen on                                                                                                | 0.0.0.0                                                                                           |
| LOG_TO_FILE                                     | Specifies whether the logs should be written to a file                                                                                        | false                                                                                             |
| LOG_FILE_PATH                                   | if LOG_TO_FILE is true, the path to the log file                                                                                              | for root user /var/log/prldevops.log. for normal user users/`<username>`/.prl-devops-service/logs |
| DISABLE_CATALOG_PROVIDER_STREAMING              | Specifies whether the catalog provider streaming should be disabled                                                                           | false                                                                                             |
| ENABLE_CANARY                                   | Specifies whether the canary feature should be enabled                                                                                        | false                                                                                             |
| ENABLE_BETA                                     | Specifies whether the beta feature should be enabled                                                                                          | false                                                                                             |

### Database

The service stores its state either in a local SQLite database file (the
default, zero configuration) or in a PostgreSQL server (enablement in
progress — SQLite is the supported backend today). These settings select
and configure the backend. The legacy JSON data store
(`data.json`) is deprecated; see the _Upgrading to the SQL database_
page for what happens on the first start after an upgrade, and the
project changelog for the planned removal version.

| Flag            | Description                                                                                                                       | Default Value                     |
| --------------- | --------------------------------------------------------------------------------------------------------------------------------- | --------------------------------- |
| DB_TYPE         | The database backend to use. One of `sqlite` or `postgres`                                                                        | sqlite                            |
| DB_PATH         | Full file path of the SQLite database file. Only used when DB_TYPE is `sqlite`; ignored when DB_TYPE is `postgres`                | `<DATABASE_FOLDER>/prl-devops.db` |
| DB_DSN          | The PostgreSQL connection string (URL or keyword/value form). Required when DB_TYPE is `postgres`. Passwords are redacted in logs |                                   |
| DB_AUTO_MIGRATE | Specifies whether the service creates or updates the database schema automatically on startup                                     | true                              |

These settings can also be set in the configuration file under the
`environment` object, using the same keys in lower case:

```yaml
environment:
  db_type: postgres
  db_dsn: postgres://<db-user>:<db-password>@db.example.com:5432/devops?sslmode=require
  db_auto_migrate: true
```

Environment variables always take precedence over configuration file
values, as described above.

### Rest API

| Flag                                | Description                                                                                                                                      | Default Value                                   |
| ----------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------- |
| API_PORT                            | The port that the service will listen on                                                                                                         | 8080                                            |
| API_PREFIX                          | The prefix that will be used for the api endpoints                                                                                               | /api                                            |
| LOG_LEVEL                           | The log level of the service                                                                                                                     | info                                            |
| GHOST_JOB_TIMEOUT_MINUTES            | Minutes without job updates before a pending or running job is marked as failed. Must be a positive integer; invalid values fall back to 5.        | 5                                               |
| HMAC_SECRET                         | The secret that will be used to sign the jwt tokens                                                                                              |                                                 |
| ENCRYPTION_PRIVATE_KEY              | The private key that will be used to encrypt the database at rest. You can generate one with the `gen-rsa` command                               |                                                 |
| TLS_ENABLED                         | Specifies whether the service should use tls                                                                                                     | false                                           |
| TLS_PORT                            | The port that the service will listen on for tls                                                                                                 | 8443                                            |
| TLS_CERTIFICATE                     | A base64 encoded certificate string                                                                                                              |                                                 |
| TLS_PRIVATE_KEY                     | A base64 encoded private key string                                                                                                              |                                                 |
| TLS_DISABLE_VALIDATION              | Disables server certificate verification for outbound HTTP requests and orchestrator host WebSocket probes/connections.                          | false                                           |
| DISABLE_HTTP_WHEN_TLS               | Specifies whether HTTP should be disabled when TLS is properly configured (HTTPS-only mode). Validates TLS configuration before disabling HTTP   | false                                           |
| ROOT_PASSWORD                       | The root password to run the service with                                                                                                        |                                                 |
| DISABLE_CATALOG_CACHING             | Specifies whether the service should disable the catalog caching                                                                                 | false                                           |
| TOKEN_DURATION_MINUTES              | The duration in minutes that the token will be valid for in minutes                                                                              | 60                                              |
| USE_ORCHESTRATOR_RESOURCES          | Specifies whether the service is running in orchestrator mode, which allows the service to use the resources of the orchestrator                 | false                                           |
| ORCHESTRATOR_PULL_FREQUENCY_SECONDS | The frequency in seconds that the orchestrator will sync with the other hosts in seconds                                                         | 30                                              |
| ENABLE_CORS                         | Specifies whether the service should enable cors policy                                                                                          | false                                           |
| CORS_ALLOWED_HEADERS                | The headers that are allowed in the cors policy                                                                                                  | "X-Requested-With, authorization, content-type" |
| CORS_ALLOWED_ORIGINS                | The origins that are allowed in the cors policy                                                                                                  | "*"                                             |
| CORS_ALLOWED_METHODS                | The methods that are allowed in the cors policy                                                                                                  | "GET, HEAD, POST, PUT, DELETE, OPTIONS"         |
| ENABLE_PACKER_PLUGIN                | Specifies whether the service should enable the packer plugin                                                                                    | false                                           |
| ENABLE_VAGRANT_PLUGIN               | Specifies whether the service should enable the vagrant plugin                                                                                   | false                                           |
| VM_FORCE_CACHE_REFRESH              | Specifies whether the service should force cache refresh for virtual machines                                                                    | true                                            |
| VM_CACHE_REFRESH_INTERVAL_SECONDS   | The interval in seconds that the service will refresh the cache for virtual machines in seconds. Make sure VM_FORCE_CACHE_REFRESH is set to true | 300                                             |

### Json Web Tokens

| Flag               | Description                                                                                                                                       | Default Value |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------- | ------------- |
| JWT_SIGN_ALGORITHM | The algorithm that will be used to sign the jwt tokens. This can be either `HS256`, `RS256`, `HS384`, `RS384`, `HS512`, `RS512`                   | HS256         |
| JWT_PRIVATE_KEY    | The private key that will be used to sign the jwt tokens. This is only required if you are using `RS256`, `RS384` or `RS512`                      |               |
| JWT_HMACS_SECRET   | The secret that will be used to sign the jwt tokens. This is only required if you are using `HS256`, `HS384` or `HS512`. Defaults to random       |
| JWT_DURATION       | The duration that the jwt token will be valid for. You can use the following format, for example, 5 minutes would be `5m` or 1 hour would be `1h` | 15m           |

### Password Complexity

| Flag                                   | Description                                                                    | Default Value |
| -------------------------------------- | ------------------------------------------------------------------------------ | ------------- |
| SECURITY_PASSWORD_MIN_PASSWORD_LENGTH  | The minimum length that the password should be. The minimum is 8               | 12            |
| SECURITY_PASSWORD_MAX_PASSWORD_LENGTH  | The maximum length that the password should be. The maximum is 40              | 40            |
| SECURITY_PASSWORD_REQUIRE_UPPERCASE    | Specifies whether the password should require at least one uppercase character | true          |
| SECURITY_PASSWORD_REQUIRE_LOWERCASE    | Specifies whether the password should require at least one lowercase character | true          |
| SECURITY_PASSWORD_REQUIRE_NUMBER       | Specifies whether the password should require at least one number              | true          |
| SECURITY_PASSWORD_REQUIRE_SPECIAL_CHAR | Specifies whether the password should require at least one special character   | true          |
| SECURITY_PASSWORD_SALT_PASSWORD        | Specifies whether the password should be salted                                | true          |

### Job inactivity timeout

Set `GHOST_JOB_TIMEOUT_MINUTES=30` in the service environment to allow 30 minutes
without job updates before marking a pending or running job as failed. Configure
this on both the orchestrator and the hosts, then restart the services. The default
is 5 minutes; missing, invalid, nonpositive, or excessively large values use the
default.

This is an inactivity threshold, not a limit on total VM creation time. The checker
runs every 5 seconds. A timeout publishes a failed job update with the threshold
and last update timestamp in both `error` and `message`. It does not stop the host
operation, so check the VM status before retrying. If a host completes successfully
but the orchestrator times out, check the host's WebSocket event connection.

### Maximum number of VMs per host

Set `MAX_VMS_PER_HOST=10` in the environment of a host to cap the number of virtual
machines that host may run, regardless of how much CPU, memory and disk are still free.
The default `0` (also used for invalid or negative values) means unlimited. Configure it
on the host, not on the orchestrator, and restart the host service.

`MAX_VMS_PER_HOST_MODE` selects which virtual machines the cap counts: `any` (the
default) counts every VM registered on the host, stopped ones included, while `running`
counts only the VMs occupying a hypervisor slot, that is running, paused or mid
transition. Unknown values fall back to `any`.

The limit is enforced in two places. The host service admits each create, clone and
start, and rejects an operation that would exceed the cap with an HTTP 409 naming the
current, pending and maximum numbers; concurrent operations cannot slip past the cap.
The orchestrator additionally refuses to place a new virtual machine on a host that is
already at its limit, and reports the shortage in the `POST /v1/machines/check`
pre-flight response as the `vm_count` check.

Each operation is charged once, to the resource it actually consumes: in `any` mode a
create or a clone takes a slot and a start is free, in `running` mode a start takes a
slot and a create is free. Lowering the value does not stop or migrate VMs that already
exist, it only blocks new ones. The limit covers operations submitted to the service:
virtual machines created directly with `prlctl` or another tool outside the service are
counted once they are registered on the host, but they are not prevented by it. Hosts
without a virtualisation inventory, such as Linux and Windows hosts, do not report a
limit, so the orchestrator keeps placing VMs on them by CPU, memory and disk only.

### Brute Force Protection

| Flag                           | Description                                                                                                                                      | Default Value |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------- |
| BRUTE_FORCE_MAX_LOGIN_ATTEMPTS | The maximum number of login attempts before the account is locked                                                                                | 5             |
| BRUTE_FORCE_LOCKOUT_DURATION   | The duration that the account will be locked for. You can use the following format, for example, 5 minutes would be `5m` or 1 hour would be `1h` | 5s            |
| BRUTE_FORCE_INCREMENTAL_WAIT   | Specifies whether the wait period should be incremental. If set to false, the wait period will be the same for each failed attempt               | true          |
