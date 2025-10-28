---
layout: page
title: Catalog
subtitle: PDFile
menubar: docs_devops_menu
show_sidebar: false
toc: true
---

# PDFile

A PDFile describes how to interact with the catalog service in a single, shareable manifest. It captures connection details, authentication, metadata, and the operation you want to run. PDFile manifests are especially handy for automating catalog pushes, pulls, imports, and scripted validations.

All credentials shown below are placeholders—replace them with your own secrets before running any command.

## Examples

### Create a push PDFile

This example uploads a Parallels Desktop VM to a catalog bucket. Save the content to a file such as `example.push.pdfile`.

```pdfile
TO https://catalog.example.local
INSECURE false

AUTHENTICATE USERNAME demo-user
AUTHENTICATE PASSWORD demo-password

PROVIDER provider=minio;endpoint=https://minio.example.local:9000;bucket=demo-catalog;access_key=demo-access;secret_key=demo-secret;use_ssl=true;ignore_cert=true

CATALOG_ID example-vm
VERSION 1.0
ARCHITECTURE arm64

ROLE ADMINISTRATOR
CLAIM UPDATE_CATALOG
TAG latest,ubuntu
DESCRIPTION Ubuntu 24.04 LTS with developer tooling

FORCE true

MINIMUM_REQUIREMENT CPU 2
MINIMUM_REQUIREMENT MEMORY 4096
MINIMUM_REQUIREMENT DISK 90000
COMPRESS_PACK true
COMPRESS_PACK_LEVEL best_compression

LOCAL_PATH /Volumes/vms/example.pvm
```

Run it:

```prldevops
prldevops push example.push.pdfile
```

Attention: Be careful when sharing PDFiles for **push** has they will contain sensitive authentication information.

### Pull a catalog machine

```prldevops
prldevops pull example.pull.pdfile
```

Use `FROM` when you expect to download a machine. The CLI command decides the actual action, so the PDFile can stay reusable across environments.


```pdfile
FROM catalog.example.local
INSECURE true

AUTHENTICATE API_KEY demo-api-key

CATALOG_ID example-vm
VERSION 1.0
ARCHITECTURE arm64

DESTINATION /Users/demo/Parallels
MACHINE_NAME example-ci-runner
OWNER demo
START_AFTER_PULL true
EXECUTE prlctl exec MachineName --info
CLONE example-ci-runner-copy

RUN pull
```

Run it:

```prldevops
prldevops pull example.pull.pdfile
```

### List catalog metadata

```pdfile
FROM catalog.example.local
INSECURE true

AUTHENTICATE API_KEY demo-api-key

CATALOG_ID example-latest
VERSION 1.0
ARCHITECTURE arm64

RUN list
```

Run it:

```prldevops
prldevops list example.list.pdfile
```

### Import a catalog manifest

```pdfile
TO catalog.example.local
INSECURE false

AUTHENTICATE API_KEY demo-api-key

CATALOG_ID example-vm
VERSION 1.0
ARCHITECTURE arm64

PROVIDER provider=minio;endpoint=https://minio.example.local:9000;bucket=demo-catalog;access_key=demo-access;secret_key=demo-secret;use_ssl=true

RUN import
```

Run it:

```prldevops
prldevops run example.import.pdfile
```

### Import a packaged VM

```pdfile
TO catalog.example.local

AUTHENTICATE API_KEY demo-api-key

CATALOG_ID example-latest
VERSION 1.0
ARCHITECTURE arm64
DESCRIPTION Ubuntu 24.04 image uploaded from object storage
TAG ubuntu,desktop
ROLE DEVELOPER
CLAIM PULL_VM

IS_COMPRESSED true
VM_TYPE pvm
VM_SIZE 25000
VM_REMOTE_PATH s3://demo-catalog/ubuntu/ubuntu-24.04.tar.gz
PROVIDER provider=minio;endpoint=https://minio.example.local:9000;bucket=demo-catalog;access_key=demo-access;secret_key=demo-secret;use_ssl=true
FORCE true

RUN import-vm
```

Run it:

```prldevops
prldevops run example.import-vm.pdfile
```

## Running a PDFile

Use `prldevops run <path>` to respect the `RUN`, `PULL`, `IMPORT`, or `IMPORT-VM` directives embedded in the PDFile. You can also call the sub-commands directly to override the directive:

```prldevops
prldevops push example.push.pdfile
prldevops pull example.pull.pdfile
prldevops list example.list.pdfile
prldevops run example.import.pdfile
prldevops run example.import-vm.pdfile
```

## Available Commands

{: .table .table-bordered .table-striped .table-hover}

| Command | Options | Description | Used For | Example |
| --- | --- | --- | --- | --- |
| TO | {url} | Catalog endpoint used when sending data. | push, import, import-vm | `TO catalog.example.local` |
| FROM | {url} | Catalog endpoint to read from. | pull, list, import, import-vm | `FROM catalog.example.local` |
| INSECURE | {boolean} | Force HTTP and skip TLS validation. | push, pull, list, import, import-vm | `INSECURE true` |
| PREFIX | {prefix} | API prefix for the catalog endpoint. | push, pull, list, import, import-vm | `PREFIX /api/v1` |
| AUTHENTICATE | {option} {value} | Authentication credentials (repeat for each option). | push, pull, list, import, import-vm |  |
|  | *USERNAME* | Username for basic authentication. | push, pull, list, import, import-vm | `AUTHENTICATE USERNAME demo-user` |
|  | *PASSWORD* | Password for basic authentication. | push, pull, list, import, import-vm | `AUTHENTICATE PASSWORD demo-password` |
|  | *API_KEY* | API key authentication alternative. | push, pull, list, import, import-vm | `AUTHENTICATE API_KEY demo-api-key` |
| DESCRIPTION | {description} | Human readable manifest description. | push (optional), import-vm (optional) | `DESCRIPTION Ubuntu 24.04 LTS` |
| TAG | {tag} | Comma separated tags saved with the manifest. | push (optional), import-vm (optional) | `TAG latest,ubuntu` |
| CATALOG_ID | {id} | Catalog identifier. | push, pull, list, import, import-vm | `CATALOG_ID ubuntu-latest` |
| VERSION | {version} | Manifest version. | push, pull, list, import, import-vm | `VERSION 24.04-1` |
| ARCHITECTURE | {architecture} | Guest architecture (`x86_64`, `arm64`). | push, pull, list, import, import-vm | `ARCHITECTURE arm64` |
| LOCAL_PATH | {path} | VM bundle to upload. | push | `LOCAL_PATH /Volumes/vms/ubuntu-latest.pvm` |
| ROLE | {role} | Required roles stored on the manifest. | push, import-vm | `ROLE ADMINISTRATOR` |
| CLAIM | {claim} | Required claims stored on the manifest. | push, import-vm | `CLAIM UPDATE_CATALOG` |
| PROVIDER | {option} {value} or connection string | Remote storage connection info. | push, import, import-vm | `PROVIDER provider=minio;endpoint=https://...` |
|  | *NAME* | Provider name when listing attributes separately. | push, import, import-vm | `PROVIDER NAME minio` |
| MACHINE_NAME | {name} | Local VM name after pull. | pull | `MACHINE_NAME ubuntu-ci-runner` |
| OWNER | {owner} | Owner associated with the pulled VM. | pull | `OWNER demo` |
| DESTINATION | {path} | Directory where the VM will be stored. | pull | `DESTINATION ~/Parallels` |
| START_AFTER_PULL | {boolean} | Starts the VM immediately after pulling. | pull | `START_AFTER_PULL true` |
| COMMAND | {pdfile command} | Alternative to RUN, specifies the PDFile command. | push, pull, list, import, import-vm | `COMMAND PUSH` |
| EXECUTE | {command} | Command executed once the VM is available. | pull (optional) | `EXECUTE prlctl exec "MachineName" --info` |
| CLONE | {name} | Clone the pulled VM locally into a new name. | pull (optional) | `CLONE ubuntu-ci-runner-copy` |
| CLONE_TO | {name} | Alternative to CLONE, specifies the target name for the clone operation. | pull (optional) | `CLONE_TO ubuntu-ci-runner-new` |
| CLONE_ID |
| CLIENT | {base64} | Base64 telemetry payload forwarded with the pull. | pull (optional) | `CLIENT eyJldmVudCI6...` |
| MINIMUM_REQUIREMENT | {metric value} | Minimum CPU, memory, or disk requirements saved with the manifest. | push (optional) | `MINIMUM_REQUIREMENT CPU 4` |
| COMPRESS_PACK | {boolean} | Compresses the upload into a `.pdpack`. | push (optional) | `COMPRESS_PACK true` |
| COMPRESS_PACK_LEVEL | {level} | Compression level (`best_speed`, `balanced`, `best_compression`, `default`, `no_compression`). | push (optional) | `COMPRESS_PACK_LEVEL best_compression` |
| IS_COMPRESSED | {boolean} | Indicates the remote machine archive is already compressed. | import-vm | `IS_COMPRESSED true` |
| VM_TYPE | {type} | Remote VM type (for example `parallels-desktop`). | import-vm | `VM_TYPE parallels-desktop` |
| VM_SIZE | {size} | Size of the remote VM in MB. | import-vm | `VM_SIZE 25000` |
| VM_REMOTE_PATH | {uri} | Location of the VM archive in object storage. | import-vm | `VM_REMOTE_PATH s3://demo-catalog/ubuntu/ubuntu-24.04.tar.gz` |
| FORCE | {boolean} | Overwrite existing catalog metadata during import-vm. | import-vm | `FORCE true` |
| RUN | {operation} | Sets the PDFile command (`PUSH`, `PULL`, `LIST`, `IMPORT`, `IMPORT-VM`). | push, pull, list, import, import-vm | `RUN IMPORT-VM` |
