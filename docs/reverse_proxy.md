# Reverse Proxy

The **Reverse Proxy** feature in the DevOps Service provides a powerful and dynamic way to route HTTP(S) and TCP traffic to various internal services or running virtual machines (VMs).

This is especially helpful for exposing specific ports of running VMs to the external network through a unified host interface, securely and efficiently.

## Core Concepts

The reverse proxy functionality revolves around three main components: **Service Configuration**, **Hosts**, and **Routes**.

### 1. Service Configuration

The global service configuration determines whether the reverse proxy engine is active and sets default global settings.

- **Status**: The reverse proxy can be enabled, disabled, or restarted at a global level. Disabling it stops all routing functions.
- **Default Host / Port**: Defines the default listening interface if an individual Host does not specify one.

**Relevant Endpoints:**
- `GET /v1/reverse-proxy` - Retrieve global proxy configurations.
- `PUT /v1/reverse-proxy/enable` - Start and enable the proxy engine.
- `PUT /v1/reverse-proxy/disable` - Stop and disable the proxy engine.
- `PUT /v1/reverse-proxy/restart` - Restart the proxy engine (useful to forcefully flush connections or reload configs).

### 2. Hosts

A **Host** represents a distinct listening interface (a combination of a domain name/IP and a port address). You can define multiple hosts within the reverse proxy to manage different streams of traffic.

- **Host**: The domain name or IP address the proxy listens on for this specific configuration.
- **Port**: The port to listen on.
- **CORS Requirements**: Cross-Origin Resource Sharing (CORS) can be configured globally per HTTP Host.
- **TLS Configuration**: (*Upcoming*) Specifies SSL certificates to secure communication for the Host.

**Note on Constraints:** A single Host can either act as an HTTP Proxy (handling a list of HTTP Routes) OR a TCP Proxy (handling a single TCP Route). It cannot do both simultaneously.

**Relevant Endpoints:**
- `GET /v1/reverse-proxy/hosts` - List all configured hosts.
- `POST /v1/reverse-proxy/hosts` - Create a new proxy host.
- `GET /v1/reverse-proxy/hosts/{id}` - Get details of a specific host.
- `PUT /v1/reverse-proxy/hosts/{id}` - Update a host configuration.
- `DELETE /v1/reverse-proxy/hosts/{id}` - Delete a host.

### 3. Routes

Routes define how traffic reaching a specific Host should be directed. They are divided into two types: **HTTP Routes** and **TCP Routes**.

#### HTTP Routes

HTTP Routes inspect incoming traffic based on URL Paths and forward requests intelligently.

- **Path Matching**: Routes can match literal strings or complex **Regular Expressions (Regex)** to catch precise subsets of URLs.
- **Targeting By Host/Port**: You can forward traffic strictly to a specific IP address and port (e.g., `10.0.0.5:8080`).
- **Dynamic VM Resolution (`target_vm_id`)**: Instead of providing an IP address that might change, you can provide the core ID of a Parallels Desktop VM. The DevOps Service will automatically query the running VM's `InternalIpAddress` and forward traffic there dynamically. If the VM stops or its IP changes, the logic updates seamlessly.
- **Header Injection**: HTTP Routes can inject custom response headers back to the client.

**Relevant Endpoints:**
- `POST /v1/reverse-proxy/hosts/{id}/http_routes` - Add or update an HTTP route for the given host.
- `DELETE /v1/reverse-proxy/hosts/{id}/http_routes/{route_id}` - Remove an HTTP route.

#### TCP Routes

TCP Routes provide raw, lower-level socket forwarding. This is useful for forwarding database connections, SSH traffic, or specialized protocols that do not adhere to standard HTTP structures.

- **Functionality**: Once a client connects to the Host's listening interface, the proxy opens a direct TCP tunnel to the target destination.
- **Targeting**: Similar to HTTP Routes, TCP Routes can forward statically (via IP/Port) or dynamically (via `target_vm_id` and a target port).
- **Restrictions**: 
  - A TCP Route handles *all* traffic hitting the Host. Therefore, you can only define **one** TCP Route per Host.
  - Because it operates below the HTTP level, features like CORS and Header Injection cannot be applied to Hosts using TCP routing.

**Relevant Endpoints:**
- `POST /v1/reverse-proxy/hosts/{id}/tcp_route` - Set or update the TCP route for a host.

## Common Operations Workflow

A typical scenario for exposing a web server running inside a VM:

1. **Start the Proxy Engine**: Ensure the reverse proxy is enabled via `PUT /v1/reverse-proxy/enable`.
2. **Create a Host**: Define the external listening interface (e.g., `proxy.local:80`) via `POST /v1/reverse-proxy/hosts`.
3. **Attach an HTTP Route**: Post an HTTP route mapping `Path: "/app"` to a `target_vm_id` running your containerized web app, specifying the app's internal container port (e.g., `TargetPort: 8080`).
4. **Access the App**: The outside world reaches `http://proxy.local/app`, and the service dynamically tunnels it to the internal VM running the web server.
