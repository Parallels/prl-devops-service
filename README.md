# Parallels Desktop Rest API Service

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) 
[![Build](https://github.com/Parallels/pd-api-service/actions/workflows/pr.yml/badge.svg)](https://github.com/Parallels/pd-api-service/actions/workflows/pr.yml) 
[![Publish](https://github.com/Parallels/pd-api-service/actions/workflows/publish.yml/badge.svg)](https://github.com/Parallels/pd-api-service/actions/workflows/publish.yml) 
[![](https://dcbadge.vercel.app/api/server/pEwZ254C3d?style=flat&theme=default)](https://discord.gg/pEwZ254C3d)

## Description

This service is a wrapper for the Parallels Desktop Rest API. It allows you to start, stop, pause, resume, and reset a virtual machine. It also allows you to get the status of a virtual machine.

## Architecture

The Parallels Desktop Rest API Service is a service that is written in Go and it is a very light height designed to provide some of the missing remote management tools for virtual machines running remotely in Parallels Desktop. It uses rest api to execute the necessary steps. It also has RBAC (Role Based Access Control) to allow for a secure way of managing virtual machines. You can manage most of the operations for a Virtual Machine Lifecycle.

## Installation

If you want to run the service locally, you can either build the source code by cloning this repository or you can use the already provided binaries in the github releases.

### Build from source

To build the source code, you need to have Go installed on your machine. You can download it from [here](https://golang.org/dl/). Once you have Go installed, you can clone this repository and run the following command:

```bash
go build ./src
```

This will create a binary called `pd-api-service` in the root directory of the repository. You can then run this binary to start the service.

### Download the binary

You can download the binary from the [releases](https://github.com/Parallels/pd-api-service/releases) page. Once you have downloaded the binary, you can run it to start the service.

## Usage

Once you start the api it will by default be listening on [localhost:8080](http://localhost:8080), you can change the port by passing the `--port=<port_number>` flag.

Once the service is running, you can use the inbuilt swagger documentation to see the available endpoints and their usage. You can access the swagger documentation at [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html).

we also provide a `Postman` collection that you can use to test the endpoints. You can download the collection from [here](docs/Parallels_Desktop_API.postman_collection.json).

## Docker Container

We provide a docker file for building the service as a container. You can build the container by running the following command:

```bash
docker build -t pd-api-service .
```

You can then run the container by running the following command:

```bash
docker run -p 8080:8080 pd-api-service
```

We also provide a docker compose file that will run the service and the database in a container. You can run the compose file by running the following command:

```bash
docker-compose up
```

you can also change the `docker-compose.yml` file to change the port, define mode or any other environment variable to personalize how you want it to run. you can also create  a `docker-compose.override.yml` file to override the default values.

## Kubernetes

We provide a helm chart for deploying the service in a kubernetes cluster. You can find the helm chart in the [helm](helm) directory.

## Security

### Authentication
The service uses JWT tokens to authenticate the users. The service will generate a JWT token that will be valid for 60 minutes and it will be signed with a secret that is defined in the environment variable `HMAC_SECRET`. You can then use this token to authenticate with the service.

### Database Encryption

The service uses a database to store the virtual machines and the database is encrypted at rest using a private key that is defined in the environment variable `SECURITY_PRIVATE_KEY`. You can generate a new private key by running the service with the flag `--gen-rsa` and it will print the private key to the console.

**Attention:** If you do not define a private key your database will be decrypted and your data be stored in plain text.

**Attention:** If you change the private key, you will not be able to decrypt the database and you will lose all the data.

## Configuration

You can configure the service by passing the following flags:
The format will be --flag=value

| Flag | Description | Default Value |
| ---- | ----------- | ------------- |
| --port | The port that the service will listen on | 8080 |
| --log-level | The log level of the service | info |
| --update-root-password | The root password that will be used to update the root password of the virtual machine | |
| --gen-rsa | Generate a new rsa key pair, this also takes the --file flag for the output | |
| --install | Install the service as a service in the host | |
| --uninstall | Uninstall the service from the host | |
| --mode | The mode that the service will run in, this can be either `api` or `orchestrator` | api |
| --use-orchestrator-resources | If the service is running in orchestrator mode, this will allow the service to use the resources of the orchestrator | false |

### Environment Variables

some of the flags can also be defined as environment variables to set it up and there are more configuration that can only be defined as environment variables.

| Flag | Description | Default Value |
| ---- | ----------- | ------------- |
| API_PORT | The port that the service will listen on | 8080 |
| LOG_LEVEL | The log level of the service | info |
| HMAC_SECRET | The secret that will be used to sign the jwt tokens | |
| SECURITY_PRIVATE_KEY | The private key that will be used to encrypt the database at rest, you can generate one with the flag `--gen-rsa` | |
| TLS_ENABLED | If the service should use tls | false |
| TLS_PORT | The port that the service will listen on for tls | 8443 |
| TLS_CERTIFICATE | A base64 encoded certificate string | |
| TLS_PRIVATE_KEY | A base64 encoded private key string | |
| API_PREFIX | The prefix that will be used for the api endpoints | /api |
| ROOT_PASSWORD | The root password that will be used to update the root password of the virtual machine | |
| DISABLE_CATALOG_CACHING | If the service should disable the catalog caching | false |
| TOKEN_DURATION_MINUTES | The duration of the jwt token in minutes | 60 |
| MODE | The mode that the service will run in, this can be either `api` or `orchestrator` | api |
| USE_ORCHESTRATOR_RESOURCES | If the service is running in orchestrator mode, this will allow the service to use the resources of the orchestrator | false |

## Catalog Manifests

The Catalog Manifest is an open source implementation of storing and distributing Parallels Desktop virtual machines in a standardized format. This contains some security concept and allows for a quick ans secure way of distributing virtual machines to know more please check the [Catalog Manifests](docs/catalog.md) documentation.

## Orchestrator

The Parallels Desktop Orchestrator Service is a service that can run in a container or directly in a host and will allow you to orchestrate and manage multiple Parallels Desktop Api Services. This will allow in a simple way to have a single pane of glass to manage multiple Parallels Desktop Api Services and check their status. To know more please check the [Orchestrator](docs/orchestrator.md) documentation.


## Contributing

If you want to contribute to this project, please read the [contributing guidelines](CONTRIBUTING.md).

## License

This project is licensed under the [MIT License](LICENSE).