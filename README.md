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

## Catalog Manifests

The Catalog Manifest is an open source implementation of storing and distributing Parallels Desktop virtual machines in a standardized format. This contains some security concept and allows for a quick ans secure way of distributing virtual machines to know more please check the [Catalog Manifests](docs/catalog.md) documentation.

## Contributing

If you want to contribute to this project, please read the [contributing guidelines](CONTRIBUTING.md).

## License

This project is licensed under the [MIT License](LICENSE).