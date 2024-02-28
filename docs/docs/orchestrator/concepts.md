---
layout: page
title: Orchestrator
subtitle: Concepts
menubar: docs_menu
show_sidebar: false
toc: true
---

## Hosts

A host is a connector to a Parallels Desktop Api Service, this will allow the
orchestrator to connect to the service and manage it. The orchestrator will keep
an eye on the status of the host and will record any changes that it sees, like
for example the available resources, it's health state and the virtual machines
that are running on it.

## Virtual Machines

A virtual machine is a virtual machine that is running on a host, the
orchestrator will keep an eye on the status of the virtual machine and will
record any changes that it sees, like for example the state of the virtual
machine, the host that is running on and the resources that it is using.