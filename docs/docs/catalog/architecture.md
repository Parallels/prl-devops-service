---
layout: page
title: Catalog
subtitle: Architecture
menubar: docs_menu
show_sidebar: false
toc: false
---

The Catalog Manifest is a service that is written in Go and it is a very light
height service that can be deployed in a container or on a virtual machine. It
uses rest api to execute the necessary steps. It also has RBAC (Role Based
Access Control) to allow for a secure way of distributing virtual machines. We
also use the concept of **Taint** and **Revoke** to allow for a secure way of
distributing virtual machines.
This will make the management of a global catalog of virtual machines very easy
and secure and in s centralized way.

![Catalog Manifest Architecture](../../../img/devtools_service-catalog_manifest.drawio.png)