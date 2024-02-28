---
layout: page
title: Orchestrator
subtitle: Getting Started
menubar: docs_menu
show_sidebar: false
toc: true
---

## Running the Orchestrator

The orchestrator can be run in two different ways, you can run it as a container
or you can run it directly in the host.
It will be the same binary that will be used in both cases, the only difference
is how you run it. If you run it as a container you will need to set the MODE
environment variable to orchestrator and if you run it directly in the host you
will need to set the MODE environment variable to api.

## Managing Hosts and Virtual Machines

Once the orchestrator is running you can start managing the hosts and virtual
machines, you can do this by using the swagger ui that is available at
[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
or by using the cli tool that is available in the [cli](./cli) folder.

Once you added a host, you can start managing it, you can create, start, stop
and delete virtual machines. You can also let the orchestrator create the
virtual machine for you by using the `Create Virtual Machine` endpoint and
passing the necessary parameters. The orchestrator will then choose the host
with enough resources on the same platform to run the virtual machine and will
create it for you.
