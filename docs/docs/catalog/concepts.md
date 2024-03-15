---
layout: page
title: Catalog
subtitle: Concepts
menubar: docs_menu
show_sidebar: false
toc: true
---

# Concepts


## Catalog Manifest Metadata

The Catalog Manifest Metadata is a JSON file that includes all the essential information required to recreate a virtual machine in the same state as it existed earlier. This file also contains details about the Storage Provider where the virtual machine will be stored. It's important to note that the Service doesn't store any parts of the Virtual Machine. Instead, it passes all the relevant information to the [storage provider](#storage-providers), making it a lightweight approach for distributing the API.

## Storage Providers

This is the designated storage location for all large files. The service can have multiple manifests that store machines in different storage providers. For example, you might have a manifest that stores the machine in an `AWS S3 bucket`, and another one that is stored in `Azure Blob Storage`. This allows for a flexible way of storing virtual machines. When a user or service tries to pull a machine, the service will check the manifest to see where the machine is stored, and then it will retrieve it from the storage provider. Currently, we support the following storage providers:

* [AWS S3](https://aws.amazon.com/s3/)
* [Azure Blob Storage](https://azure.microsoft.com/en-us/services/storage/blobs/)
* [Jfrog Artifactory](https://jfrog.com/artifactory/)

## Connection String

 The connection string for accessing different storage providers requires specific variables, starting with the provider's name. 

Some examples of connection strings are:

### AWS S3

```bash
provider=aws-s3;bucket=<bucket-name>;region=<bucket-region>;access_key=<access_key>;secret_key=<access_secret>
```

### Azure Blob Storage

```bash
provider=azure-storage-account;storage_account_name=<storage-account-name>;container_name=<storage-account-container>;storage_account_key=<storage-account-key>
```

### Jfrog Artifactory

```bash
provider=artifactory;url=<artifactory_url_without_artifact>;repo=<repo_name>;access_key=<api_access_key>
```

# Catalog Manifest and Versions

Each Catalog Manifest in the system has a unique identifier called an id, as well as a version number. The id is used to distinguish between different manifests, while the version number is used to track changes made to a particular manifest. Whenever a virtual machine undergoes an update, a new version of the virtual machine must be defined. You can use version semantics to define the version, which is a free field that works similarly to the tags in docker. Each version represents a complete version of the virtual machine, meaning that to update a virtual machine, you must create a new version of the virtual machine, and then update the manifest to point to the new version.

# Taint

Taint is a feature that allows you to mark a particular version of a virtual machine as tainted. This means that the virtual machine should not be used anymore as it is not usable. This feature comes in handy when you want to remove a version of a virtual machine, or when you have security concerns about a particular version. By marking it as tainted, the service will prevent users from using it, and this will help administrators manage a large number of deployments more efficiently. Tainting a virtual machine is not permanent, you can untaint a version of a virtual machine, and it will be usable again.

{% include notification.html message="Note: The system will only check taints if a pull request is made. If a virtual machine is already running, the system will not check if the version is tainted." status="is-warning" %}

# Revoke

Revoke is a feature that allows you to revoke a version of a virtual machine, rendering it unusable. This is particularly helpful when you want to discontinue a specific version of a virtual machine. By revoking it, the service prevents users from accessing that version of the virtual machine. This is a permanent action, and once a version is revoked, it cannot be used again. This is an excellent way to address security concerns related to a particular version of a virtual machine.

{% include notification.html message="Note: The system will only check revokes if a pull request is made. If a virtual machine is already running, the system will not check if the version is revoked." status="is-warning" %}

# RBAC

Each Catalog Manifest has the feature to specify a mandatory Claim or Role that a user must possess to be able to view it. This allows for a precise way of controlling access to virtual machines. It is particularly useful when you want to provide virtual machines to a specific group of users. For instance, you may want to distribute a virtual machine to your developers and another to your designers. You can create two manifests and define the required claim for each of them. This will enable you to manage access to the virtual machines in a very precise manner.

# Importing a Virtual Machine

Our service supports the import of a metadata file that contains information about a virtual machine. This feature allows you to import a virtual machine that was not created using our service. It's particularly useful when you want to import a virtual machine that was created using a different tool or manually. With this feature, you'll be able to manage all of your virtual machines in one centralized location.

# Push

"Push" refers to the process of uploading a virtual machine to the catalog manifest service. To push a virtual machine, the client follows these steps:

1. Calls the remote catalog service to check if the request is valid.
2. Calls the remote storage service to validate the connection.
3. Checks for any security requirements.
4. Uploads the virtual machine package to the remote storage service.
5. Creates the metadata entry in the remote catalog service.

By separating the responsibilities of the remote catalog and the storage of virtual machine packages, a more dynamic implementation is possible without too many constraints.

# Pull

Pulling a virtual machine from the catalog manifest service allows clients to do heavy lifting themselves instead of the service doing it for them. To pull a virtual machine, the client follows these steps: first, it calls the remote catalog service using the provided connection string. Next, it checks for the virtual machine's metadata and its storage location. Once it finds this information, the client pulls the virtual machine from the storage provider. Afterward, it creates the virtual machine using the metadata and starts it up. This approach is not only more efficient but also provides a secure connection between the client and the storage provider.

## Pulling a Virtual Machine Diagram

![Pulling a Virtual Machine](../../../img/devtools_service-catalog_manifest_pulling.drawio.png)

# Caching

A Virtual Machine (VM) file can be very large, and it can take a lot of time to pull it every time you want to use it. To solve this problem, we have implemented a caching mechanism that allows you to cache the VM locally and then use it from the cache. The mechanism works by checking if the content checksum matches the one in the cache, and if it does, the client will use the cached version. This will significantly reduce the time it takes to pull the VM and make the process much faster.


# PDFile

We have developed a file structure similar to Docker manifest files which we call PDFile. This file contains all the necessary information required to push or pull a virtual machine from the catalog that makes it easy to share or store, allowing for a better automation flow. 

Here is an example of a PDFile:

```pdfile
TO example.com
INSECURE true

AUTHENTICATE USERNAME root
AUTHENTICATE PASSWORD someverylongpassword

CATALOG_ID test-catalog
VERSION 1.0
ARCHITECTURE arm64
LOCAL_PATH /Users/foobar/Parallels/macOS_Github_actions_runner.macvm
ROLE Admin
ROLE User
CLAIM UPDATE_CATALOG

PROVIDER NAME aws-s3
PROVIDER BUCKET catalog-example
PROVIDER REGION us-east-2
PROVIDER ACCESS_KEY SOMEACCESSKEY
PROVIDER SECRET_KEY SOMESECRETKEY
```

You can run this command directly from the command line to push:

```prldevops
prldevops push ./example.pdfile 
```

For pulling, it will be a similar process:

```pdfile
FROM example.com
INSECURE true

AUTHENTICATE USERNAME root
AUTHENTICATE PASSWORD someverylongpassword

CATALOG_ID test_push_1
VERSION v1
ARCHITECTURE arm64

MACHINE_NAME test_pull_1
OWNER foobar
DESTINATION /Users/foobar/Parallels
START_AFTER_PULL false
```

And then run this command:

```prldevops
prldevops pull ./example.pdfile 
```
