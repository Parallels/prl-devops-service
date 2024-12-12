---
layout: page
title: Catalog
subtitle: PDFile
menubar: docs_devops_menu
show_sidebar: false
toc: true
---

# PDFile

A PDFile is a file structure similar to Docker manifest files that contains all the necessary information required to push or pull a virtual machine from the catalog. This makes it easy to share or store, allowing for a better automation flow. It can be used to either push or pull from a catalog manifest.

## Example

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

## Available Commands

{: .table .table-bordered .table-striped .table-hover}

| Command | Options | Description | example|
| --- | --- | --- | --- |
| To | {url} | The URL of the catalog to push to. | **TO** localhost:5740 |
| FROM | {url} | The URL of the catalog to pull from. | **FROM** example.com |
| INSECURE | {boolean} | Whether to use an insecure connection. | **INSECURE** `true` |
| AUTHENTICATE | {option} {value} | The username and password or the api key to authenticate with. | |
| | *USERNAME* | The username to authenticate with. | **AUTHENTICATE** **USERNAME** root |
| | *PASSWORD* | The password to authenticate with. | **AUTHENTICATE** **PASSWORD** pass |
| | *API_KEY* | The api key to authenticate with. | **AUTHENTICATE** **API_KEY** somekey |
| DESCRIPTION | {description} | The description of the virtual machine. | **DESCRIPTION** test description |
| TAG | {tag} | The tags of the virtual machine. | **TAG** test,tag |
| CATALOG_ID | {id} | The id of the catalog to push to. | **CATALOG_ID** test-catalog |
| VERSION | {version} | The version of the virtual machine. | **VERSION** 1.0 |
| ARCHITECTURE | {architecture} | The architecture of the virtual machine. | **ARCHITECTURE** arm64 |
| LOCAL_PATH | {path} | The path to the virtual machine. | **LOCAL_PATH** /Users/foobar/Parallels/macOS_Github_actions_runner.macvm |
| ROLE | {role} | The role of the user. | **ROLE** Admin |
| CLAIM | {claim} | The claim of the user. | **CLAIM** UPDATE_CATALOG |
| PROVIDER | {option} {value} or {full connection string} | The provider to use. | |
| | *NAME* | The name of the provider. | **PROVIDER** **NAME** aws-s3 |
| | *PROVIDER_OPTION* | The bucket of the provider. | **PROVIDER** **BUCKET** catalog-example |
| | *full_connection_string* | a connection string containing all the options | **PROVIDER** provider=aws-s3;option=value |
| MACHINE_NAME | {name} | The name of the virtual machine. | **MACHINE_NAME** test_pull_1 |
| OWNER | {owner} | The owner of the virtual machine. | **OWNER** foobar |
| DESTINATION | {path} | The path to store the virtual machine. | **DESTINATION** /Users/foobar/Parallels |
| START_AFTER_PULL | {boolean} | Whether to start the virtual machine after pulling. | **START_AFTER_PULL** false |
| EXECUTE | {command} | The command to execute after pulling. | **EXECUTE** echo "Hello World" |
