---
layout: page
title: Documentation
subtitle: User-friendly hub with guides, API docs, and tutorials
menubar: docs_devops_menu
show_sidebar: false
---

# Parallels Desktop DevOps Service

This is the Parallels Desktop DevOps Service, a service that will allow you to
manage and orchestrate multiple Parallels Desktop hosts and virtual machines.
It will allow you to create, start, stop and delete virtual machines and will
also allow you to manage the hosts that are running the virtual machines.

## Licensing

You can use the service for free with up to **10 users**, without needing any
Parallels Desktop Business license. However, if you want to continue using the
service beyond 10 users, you will need to purchase a Parallels Desktop Business
license. The good news is that no extra license is required for this service
once you have purchased the Parallels Desktop Business license.

## Architecture

The Parallels Desktop DevOps is a service that is written in Go and it is a very
light height designed to provide some of the missing remote management
tools for virtual machines running remotely in Parallels Desktop. It uses rest
api to execute the necessary steps. It also has RBAC (Role Based Access Control)
to allow for a secure way of managing virtual machines. You can manage most of
the operations for a Virtual Machine Lifecycle.
