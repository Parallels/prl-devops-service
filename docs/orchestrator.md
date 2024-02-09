# Parallels Desktop Orchestrator Service

The Parallels Desktop Orchestrator Service is a service that can run in a
container or directly in a host and will allow you to orchestrate and manage
multiple Parallels Desktop API Services. This will allow in a simple way to have
a single pane of glass to manage multiple Parallels Desktop API Services and
check their status.
You will also be able to create virtual machines automatically and let the
orchestrator choose the host with enough free resources

## Architecture

The Parallels Desktop Orchestrator Service is written in go and uses the same
base code as the Parallels Desktop API Service. One single executable that
depending on how you run it it will behave in a different way, this allows for
a simpler way for deploying it.
The service itself has a simple architecture where it will have a collection of
hosts, these hosts are connectors to a Parallels Desktop API instance, the
orchestrator then will start by having a background service that keeps an eye on
the status of each host and records any changes that it sees, like for example
the available resources, it's health state and the virtual machines that are
running on it.
You can then manage each host individually by creating, starting, stopping and
deleting virtual machines or you can let the orchestrator do it for you by
creating a virtual machine and letting the orchestrator choose the host with
enough resources to run it.

![Orchestrator Architecture](./images/devtools_service-orchestrator.drawio.png)

## Concepts

### Hosts

A host is a connector to a Parallels Desktop API Service, this will allow the
orchestrator to connect to the service and manage it. The orchestrator will keep
an eye on the status of the host and will record any changes that it sees, like
for example the available resources, it's health state and the virtual machines
that are running on it.

### Virtual Machines

A virtual machine is a virtual machine that is running on a host, the
orchestrator will keep an eye on the status of the virtual machine and will
record any changes that it sees, like for example the state of the virtual
machine, the host that is running on and the resources that it is using.

## Getting Started

### Running the Orchestrator

The orchestrator can be run in two different ways, you can run it as a container
or you can run it directly in the host.
It will be the same binary that will be used in both cases, the only difference
is how you run it. If you run it as a container you will need to set the MODE
environment variable to orchestrator and if you run it directly in the host you
will need to set the MODE environment variable to API.

## Managing Hosts and Virtual Machines

Once the orchestrator is running you can start managing the hosts and virtual
machines, you can do this by using the swagger ui that is available at
[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
or by using the command-line tool that is available in the [cli](./cli) folder.

Once you added a host, you can start managing it, you can create, start, stop
and delete virtual machines. You can also let the orchestrator create the
virtual machine for you by using the `Create Virtual Machine` endpoint and
passing the necessary parameters. The orchestrator will then choose the host
with enough resources on the same platform to run the virtual machine and will
create it for you.
