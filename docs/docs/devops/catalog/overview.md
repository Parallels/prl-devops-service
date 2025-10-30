---
layout: page
title: Catalog
subtitle: Overview
menubar: docs_devops_menu
show_sidebar: false
toc: true
---

# Overview

## What is It?

Our **DevTools Catalog Service** is a metadata repository that stores virtual machines as packs in a remote storage. This approach enables users to securely store and retrieve virtual machines from a cloud service that is scalable and easy to use. The metadata is separated from the virtual machine package, making the service lightweight and deployable on any operating system, allowing it to be used in a variety of environments.

The service is designed to enable operations on an organization to easily create, store, maintain and share virtual machines on their environment with a secure and scalable solution.

## Architecture

The Catalog Repository is a service provided by the Parallels Desktop DevOps, written in GO. It is an extremely lightweight service that can be deployed in multiple ways, such as in a container or on a virtual machine. Additionally, we offer a Helm chart to install it in a Kubernetes cluster. The Catalog Repository takes advantage of the built-in Rest API capabilities of the Parallels Desktop DevOps service, making use of its security features to provide a strong and secure way of distributing virtual machines in a centralized manner.

![Catalog Pulling Diagram]({{ site.url }}{{ site.baseurl }}/img/catalog/catalog_manifest.drawio.png)

## How does it work?

There are two main types of operations used with virtual machine catalogs are called *pushing* and *pulling*. A *pushing* operation is when a virtual machine is uploaded to the catalog, while a *pulling* operation is when a virtual machine is downloaded from the catalog.

### Pushing

![Catalog Pushing Diagram]({{ site.url }}{{ site.baseurl }}/img/catalog/catalog_manifest_pushing.drawio.png)

Here is a [push example]({{ site.url }}{{ site.baseurl }}/docs/devops/catalog/pdfile/#create-a-push-pdfile) of how a flow would typically work for a pushing operation:

First, the user creates a virtual machine.

Next, the user can use either the REST API or the CLI to push the virtual machine to the catalog. This involves providing details about the virtual machine, such as the catalog name, version, and description. This step may also include any required security claims and roles.

Once the process is started, the client checks the metadata to ensure it is valid and correct. The client then pushes the packaged virtual machine to the storage and the metadata to the database.

The client then receives a response with the status of the operation.

Once complete, the virtual machine is available for download by others who have access to the catalog.

### Pulling

![Catalog Pulling Diagram]({{ site.url }}{{ site.baseurl }}/img/catalog/catalog_manifest_pulling.drawio.png)

Below is a [pull example]({{ site.url }}{{ site.baseurl }}/docs/devops/catalog/pdfile/#pull-a-catalog-machine) of how a flow typically works for a pulling operation:

1. The user requests a virtual machine from the catalog using the REST API or the CLI.
2. The service then checks the user's security claims and roles to ensure they have access to the virtual machine. It also verifies if the image is not either tainted or revoked.
3. Once the user is authenticated and authorized, the service retrieves metadata and sends it back to the client.
4. The client connects to the storage provider and downloads the virtual machine package.
5. The client then unpacks the package and registers the virtual machine in the local environment.
6. If caching is enabled, the client stores the virtual machine in the local cache, making it faster for future downloads.
