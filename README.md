# Parallels Desktop DevOps Service

[![License: Fair Source](https://img.shields.io/badge/license-fair-source.svg)](https://fair.io/)
[![Build](https://github.com/Parallels/pd-api-service/actions/workflows/pr.yml/badge.svg)](https://github.com/Parallels/pd-api-service/actions/workflows/pr.yml)
[![Publish](https://github.com/Parallels/pd-api-service/actions/workflows/publish.yml/badge.svg)](https://github.com/Parallels/pd-api-service/actions/workflows/publish.yml)
[![discord](https://dcbadge.vercel.app/api/server/pEwZ254C3d?style=flat&theme=default)](https://discord.gg/pEwZ254C3d)

## Description

This service is a wrapper for the Parallels Desktop DevOps Service. It allows you
to start, stop, pause, resume, and reset a virtual machine. It also allows you to
get the status of a virtual machine.

## Licensing

You can use the service for free with up to 10 users, without needing any
Parallels Desktop Business license. However, if you want to continue using the
service beyond 10 users, you will need to purchase a Parallels Desktop Business
license. The good news is that no extra license is required for this service
once you have purchased the Parallels Desktop Business license.

## Architecture

The Parallels Desktop DevOps is a service that is written in Go and it is a very
light height designed to provide some of the missing remote management
tools for virtual machines running remotely in Parallels Desktop. It uses rest
API to execute the necessary steps. It also has RBAC (Role Based Access Control)
to allow for a secure way of managing virtual machines. You can manage most of
the operations for a Virtual Machine Lifecycle.

## Installation

If you want to run the service locally, you can either build the source code by
cloning this repository or you can use the already provided binaries in the
GitHub releases.

### Build from source

To build the source code, you need to have Go installed on your machine. You can
download it from [here](https://golang.org/dl/). Once you have Go installed, you
can clone this repository and run the following command:

```bash
go build ./src
```

This will create a binary called `pd-api-service` in the root directory of the
repository. You can then run this binary to start the service.

### Download the binary

You can download the binary from the
[releases](https://github.com/Parallels/pd-api-service/releases) page. Once you
have downloaded the binary, you can run it to start the service.

## Usage

Once you start the API it will by default be listening on
[localhost:8080](http://localhost:8080), you can change the port by passing the
`--port=<port_number>` flag.

Once the service is running, you can use the inbuilt swagger documentation to see
the available endpoints and their usage. You can access the swagger documentation
at [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html).

we also provide a `Postman` collection that you can use to test the endpoints.
You can download the collection from [here](docs/Parallels_Desktop_API.postman_collection.json).

### Configuration

The service can be configured using either command-line flags, environment variables
or a config file. The order of precedence is the following, command-line flag,
environment variable, config file.

To use the configuration file just create a yaml file with the environment
variables, if you name it `config.yaml` or `config.yml` the service will
automatically pick it up, you can also add the `.local` before the extension to
make it a local config file, you can also specify the path to the config file
using the `--config` flag.

Here is an example of a config file:

```yaml
environment:
  api_port: 5570
  log_level: DEBUG
```

## Docker Container

We provide a docker file for building the service as a container, but be aware,
this limits the service to work in catalog or orchestrator mode as you will not
have the Parallels Desktop inside the container You can build the container by
running the following command:

```bash
docker build -t pd-api-service .
```

You can then run the container by running the following command:

```bash
docker run -p 8080:8080 pd-api-service
```

We also provide a docker compose file that will run the service and the database
in a container. You can run the compose file by running the following command:

```bash
docker-compose up
```

you can also change the `docker-compose.yml` file to change the port, define mode
or any other environment variable to personalize how you want it to run. you can
also create  a `docker-compose.override.yml` file to override the default values.

## Kubernetes

We provide a helm chart for deploying the service in a kubernetes cluster. You
can find the helm chart in the [helm](helm) directory.

## Security

### Authentication

The service uses JWT (Json Web Tokens) to authenticate the users. By default the
service will start with very harden security settings, you can lower the security
using configuration options, but we do not recommend it. To authenticate with the
service you need to use the `/auth/token` endpoint and pass the username and password
The service by default will be using the `HS256` algorithm to sign the tokens, but
you can change it to use `RS256`

Depending on what type of algorithm you are using, you will need to pass the secrets
or the RSA keys to the service.

If no secret is defined, the service will generate a random secret and use it to
sign the tokens, this will be a random string of characters.

We also will create a default `Super User` with the username `root` and the password
will be based in the hardware ID and your Parallels Desktop license. You can change
the password by passing the `--update-root-password` flag or the `ROOT_PASSWORD`
environment variable.
This will update that password when the service starts.

### Brute Force Attack Protection

The service has a brute force attack protection mechanism that will lock the account
after 5 failed attempts and will use incremental wait periods for each failed attempt.
This can be setup to use a different number of attempts and wait periods.

### Password Complexity

The service has a password complexity pipeline that will check if the password that
the user is trying to set adheres to the complexity requirements. This can be disabled
if required, by default the password complexity is enabled and the complexity is
set to 12 characters, at least one uppercase, one lowercase, one number and one special

### Password Hashing

The service will hash the passwords using the `bcrypt` algorithm by default, but
you can change it to use `sha256` if required.

### Database Encryption

The service uses a file database to persist data and the database is encrypted
at rest using a private key that is defined in the environment variable
`ENCRYPTION_PRIVATE_KEY`. You can generate a new private key by running the
service with `gen-rsa` and it will print the private key to the console.

You can pass the size of the key by passing the `--rsa-key-size` flag, the default
is 2048 bits.

**Attention:** If you do not define a private key your database will be decrypted
and your data be stored in plain text.

**Attention:** If you change the private key, you will not be able to decrypt
the database and you will lose all the data.

## Configuration

You can configure the service by running it with either flags, environment variables
or have a config file. The order of precedence is the following, command-line flag,
environment variable, config file.

This is the list of available flags/environment variables that you can use to
configure the service.

### Variables

| Flag | Description | Default Value |
| ---- | ----------- | ------------- |
| API_PORT | The port that the service will listen on | 8080 |
| API_PREFIX | The prefix that will be used for the API endpoints | /API |
| LOG_LEVEL | The log level of the service | info |
| HMAC_SECRET | The secret that will be used to sign the jwt tokens | |
| ENCRYPTION_PRIVATE_KEY | The private key that will be used to encrypt the database at rest, you can generate one with the `gen-rsa` command | |
| TLS_ENABLED | If the service should use tls | false |
| TLS_PORT | The port that the service will listen on for tls | 8443 |
| TLS_CERTIFICATE | A base64 encoded certificate string | |
| TLS_PRIVATE_KEY | A base64 encoded private key string | |
| ROOT_PASSWORD | The root password that will be used to update the root password of the virtual machine | |
| DISABLE_CATALOG_CACHING | If the service should disable the catalog caching | false |
| MODE | The mode that the service will run in, this can be either `api` or `orchestrator` | API |
| USE_ORCHESTRATOR_RESOURCES | If the service is running in orchestrator mode, this will allow the service to use the resources of the orchestrator | false |
| ORCHESTRATOR_PULL_FREQUENCY_SECONDS | The frequency in seconds that the orchestrator will sync with the other hosts in seconds | 30 |
| DATABASE_FOLDER | The folder where the database will be stored | /User/Folder/.pd-api-service |
| CATALOG_CACHE_FOLDER | The folder where the catalog cache will be stored | /User/Folder/.pd-api-service/catalog |
| Json Web Tokens | | |
| JWT_SIGN_ALGORITHM | The algorithm that will be used to sign the jwt tokens, this can be either `HS256`, `RS256`, `HS384`, `RS384`, `HS512`, `RS512` | HS256 |
| JWT_PRIVATE_KEY | The private key that will be used to sign the jwt tokens, this is only required if you are using `RS256`, `RS384` or `RS512` | |
| JWT_HMACS_SECRET | The secret that will be used to sign the jwt tokens, this is only required if you are using `HS256`, `HS384` or `HS512` | Defaults to random |
| JWT_DURATION | The duration that the jwt token will be valid for, you can use the following format, for example, 5 minutes would be `5m` or 1 hour would be `1h` | 15m |
| Password Complexity | | |
| SECURITY_PASSWORD_MIN_PASSWORD_LENGTH | The minimum length that the password should be, min is 8 | 12 |
| SECURITY_PASSWORD_MAX_PASSWORD_LENGTH | The maximum length that the password should be, max is 40 | 40 |
| SECURITY_PASSWORD_REQUIRE_UPPERCASE | If the password should require at least one uppercase character | true |
| SECURITY_PASSWORD_REQUIRE_LOWERCASE | If the password should require at least one lowercase character | true |
| SECURITY_PASSWORD_REQUIRE_NUMBER | If the password should require at least one number | true |
| SECURITY_PASSWORD_REQUIRE_SPECIAL_CHAR | If the password should require at least one special character | true |
| SECURITY_PASSWORD_SALT_PASSWORD | If the password should be salted | true |
| Brute Force Attack Protection | | |
| BRUTE_FORCE_MAX_LOGIN_ATTEMPTS | The maximum number of login attempts before the account is locked | 5 |
| BRUTE_FORCE_LOCKOUT_DURATION | The duration that the account will be locked for, you can use the following format, for example, 5 minutes would be `5m` or 1 hour would be `1h` | 5s |
| BRUTE_FORCE_INCREMENTAL_WAIT | If the wait period should be incremental, if set to false, the wait period will be the same for each failed attempt | true |

## Catalog Manifests

The Catalog Manifest is an open source implementation of storing and distributing
Parallels Desktop virtual machines in a standardized format. This contains some
security concept and allows for a quick ans secure way of distributing virtual
machines to know more please check the [Catalog Manifests](docs/catalog.md) documentation.

## Orchestrator

The Parallels Desktop Orchestrator Service is a service that can run in a
container or directly in a host and will allow you to orchestrate and manage
multiple Parallels Desktop API Services. This will allow in a simple way to have
a single pane of glass to manage multiple Parallels Desktop API Services and
check their status. To know more please check the
[Orchestrator](docs/orchestrator.md) documentation.

## Contributing

If you want to contribute to this project, please read the [contributing guidelines](CONTRIBUTING.md).

## License

This project is licensed under the [MIT License](LICENSE).
