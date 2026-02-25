# Reverse Proxy UI Agent Specification

This specification outlines the requirements and backend functionality for implementing a User Interface (UI) to manage the Reverse Proxy feature in the DevOps Service project.

Your objective as the UI agent is to build a sleek, intuitive frontend to interact with the backend APIs described below. The user should be able to globally manage the proxy engine, and gracefully control multiple complex Host routing mechanics.

## 1. System Context and Architecture

The backend is built in Go. It uses `net/http/httputil` for HTTP multiplexing and forwarding, and raw `net.Listen` sockets for TCP proxying. 

All configurations are persisted to a SQLite Database and then synced into a running state matrix in memory.
When a user updates a proxy host or route via the API, the backend automatically performs a live restart of the routing engine to apply the changes. 

**Dynamic Parallels Desktop VM Resolution**
A major feature of this proxy is its deep integration with the hypervisor. In any route (HTTP or TCP), instead of manually specifying a hardcoded IP address in `target_host`, the user can supply a `target_vm_id` string representing a Parallels Virtual Machine. The backend automatically looks up the active `InternalIpAddress` of the running VM and routes traffic there.

## 2. API Data Models

The UI will primarily send and list JSON payloads matching these struct definitions.

### `ReverseProxyConfig`
Controls the global state of the entire proxy engine.
```json
{
  "enabled": true,
  "host": "0.0.0.0", // Global fallback host
  "port": "8080"     // Global fallback port
}
```

### `ReverseProxyHost`
Defines a listening endpoint for incoming traffic. A host binds a domain/IP (`host`) and a `port` to specific routes.
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "host": "api.domain.local",
  "port": "80", 
  "tls": null, // Currently reserved / unused
  "cors": {    // Applies only if using HTTP routes
    "enabled": true,
    "allowed_origins": ["*"],
    "allowed_methods": ["GET", "POST", "OPTIONS"],
    "allowed_headers": ["Content-Type", "Authorization"]
  },
  "http_routes": [], // Array of HTTP Routes
  "tcp_route": null  // Object representing a TCP Route
}
```

### `ReverseProxyHostHttpRoute`
Defines an HTTP forwarding rule (Path -> Target).
```json
{
  "id": "abc...",
  "path": "/api/v1",       // String pattern or regex string
  "pattern": "^/api/v1/",  // Optional explicit regex
  "schema": "http",        // "http" or "https"
  "target_host": "10.0.0.5", // Static IP (ignored if target_vm_id is set)
  "target_port": "3000",
  "target_vm_id": "UUID-OF-PARALLELS-VM", // Preferred dynamic targeting
  "response_headers": {
    "X-Custom-Header": "Injected Value"
  }
}
```

### `ReverseProxyHostTcpRoute`
Defines a pure socket forwarding rule.
```json
{
  "target_host": "10.0.0.5", 
  "target_port": "22",
  "target_vm_id": "UUID-OF-PARALLELS-VM"
}
```

## 3. Required Functionality & Workflows

### 3.1 Global Proxy State Management
The UI must provide a global toggle to turn the entire proxy engine on or off, and perhaps a manual refresh/restart button.

- **GET `/v1/reverse-proxy`** - Fetch current global config.
- **PUT `/v1/reverse-proxy/enable`** - Enables the proxy.
- **PUT `/v1/reverse-proxy/disable`** - Disables the proxy.
- **PUT `/v1/reverse-proxy/restart`** - Hot reloads the proxy instance.

### 3.2 Hosts Management
The UI should feature a dashboard or table listing all configured Proxy Hosts.

- **GET `/v1/reverse-proxy/hosts`** - Returns `[]ReverseProxyHost`
- **POST `/v1/reverse-proxy/hosts`** - Create a new host.
- **PUT `/v1/reverse-proxy/hosts/{id}`** - Update an existing host (host, port, cors settings).
- **DELETE `/v1/reverse-proxy/hosts/{id}`** - Delete host.

**UI Actions needed on a Host:**
- Create/Edit Host (Fields: Host Name/IP, Listen Port).
- Configure CORS toggle and arrays for Origins, Methods, Headers.

### 3.3 Route Management (Inside a Host context)
When drilling down into a specific Host, the user needs the ability to assign routes.

**Critical Constraint:** The UI must enforce that a single Host can **either** have multiple HTTP routes **or** a single TCP route, but NEVER BOTH.
If a user adds a TCP route, the Add HTTP Route buttons should disappear or yield a warning warning them that they must delete the TCP route first, and vice versa.
Furthermore, if a TCP route is active, CORS and TLS options must be disabled/hidden at the Host level.

**Managing HTTP Routes:**
- **POST `/v1/reverse-proxy/hosts/{id}/http_routes`** - Add/Update route.
  - *Note on upsert behavior:* Provide an ID if updating an existing route, otherwise omit to create a new one. (The backend matches by Route Path internally if needed, but standard logic applies).
- **DELETE `/v1/reverse-proxy/hosts/{id}/http_routes/{route_id}`** - Remove.

**Managing TCP Routes:**
- **POST `/v1/reverse-proxy/hosts/{id}/tcp_route`** - Updates the singular TCP route available on the Host. Setting this requires tearing down any HTTP routes.

**Target Selection Dropdown (For Routes):**
When the user is configuring a target destination for any route (HTTP or TCP), they should be presented with a modern toggle or Radio Button choice for "Target Type":
- Option A: **Static Destination** -> Renders input boxes for `target_host` (IP or FQDN) and `target_port`.
- Option B: **Virtual Machine Destination** -> Renders an auto-filling dropdown of VM names/IDs available from the local Parallels Desktop instance (you may assume this comes from `/v1/orchestrator/vms` or a similar endpoint), locking `target_host` as empty, and allowing a `target_port` input. 

## 4. UI/UX Design Guidelines

The target aesthetic for this feature should feel premium, fluid, and reactive. Please adhere to the overarching modern aesthetic guidelines defined for this project:

- **Visual Style:** Use dark mode gradients, glassmorphism on modals/panels (blur backdrops), and vibrant but distinct accent colors to denote routes vs. global settings vs. VM endpoints.
- **Micro-interactions:** Actions like enabling the global proxy, adding a route, or switching between HTTP/TCP mode should feature smooth CSS transitions. For example, clearing out HTTP routes to switch to TCP mode should elegantly fade elements out of the DOM.
- **Responsiveness**: Ensure the dashboard displays gracefully regardless of screen size. 
- **Validation**: Surface backend constraints beautifully. If the user tries to save a TCP route when CORS is enabled, show an elegant inline alert or toast rather than just waiting for an HTTP 400 rejection from the API.
