---
layout: page
title: Getting Started
subtitle: Docker
menubar: docs_menu
show_sidebar: false
toc: true
---

If you want to run only the DevOps service [Catalog]({{ site.url }}{{ site.baseurl }}/docs/catalog/overview/), [Orchestrator]({{ site.url }}{{ site.baseurl }}/docs/orchestrator/overview/) or the [Reverse Proxy]({{ site.url }}{{ site.baseurl }}/docs/reverse-poxy/overview/)  we provide a docker image. This allows you to quickly spin up the service without having to install any dependencies.

## Prerequisites

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)

## Running the DevOps Service

for a quick start the `DevOps` service you can run:

```powershell
docker run -d -p 5570:80 --name prldevops cjlapao/prl-devops-service:latest
```

This will start the service and you will be able to access the swagger ui at [http://localhost:5570/swagger/index.html](http://localhost:5570/swagger/index.html)

You can pass any of the [configuration options]({{ site.url }}{{ site.baseurl }}/docs/getting-started/configuration//#configuration-file) as environment variables to the docker container.

for example:

```powershell
docker run -d -p 8008:80 --name prldevops -e API_PORT=8008 -e LOG_LEVEL=DEBUG cjlapao/prl-devops-service:latest
```

This will start the service on port 8008 and with the log level set to `DEBUG`.

## Docker Compose

We also provide supoort for `docker-compose` to make it easier to manage the service. you can create a `docker-compose.yaml` file with the following content:

```yaml
version: '3.9'
name: api
services:
  api:
    build: .
    ports:
      - "8008:80"
    environment:
      HMAC_SECRET: ''
      LOG_LEVEL: 'info'
      SECURITY_PRIVATE_KEY: ''
      TLS_ENABLED: 'false'
      TLS_PORT: '447'
      TLS_CERTIFICATE: ''
      TLS_PRIVATE_KEY: ''
      API_PORT: '80'
      API_PREFIX: '/api'
      ROOT_PASSWORD: ''
      DISABLE_CATALOG_CACHING: 'false'
      TOKEN_DURATION_MINUTES: 60
      MODE: api
      USE_ORCHESTRATOR_RESOURCES: 'false'
```

Once you have the `docker-compose.yaml` file you can start the service by running:

```powershell
docker-compose up -d
```

This will start the service and you will be able to access the swagger ui at [http://localhost:8080/swagger/index.html](http://localhost:8008/swagger/index.html)
