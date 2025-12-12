---
layout: page
title: Events Hub
subtitle: Overview
menubar: docs_devops_menu
show_sidebar: false
toc: true
---

# Events Hub Guide

The **Events Hub** is a real-time WebSocket service that enables DevOps engineers and SREs to subscribe to critical infrastructure events. Instead of polling REST APIs, you can receive instant updates about host health, VM lifecycle changes, and system alerts.

## Overview

- **Protocol**: WebSocket (RFC 6455)
- **Subscribe Endpoint**: `GET /v1/ws/subscribe?event_types=pdfm,health,orchestrator` [Subscribe API Reference]({{ site.url }}{{ site.baseurl }}/docs/devops/restapi/reference/events/#v1-ws-subscribe-get)
- **Unsubscribe Endpoint**: `POST /v1/ws/unsubscribe` [Unsubscribe API Reference]({{ site.url }}{{ site.baseurl }}/docs/devops/restapi/reference/events/#v1-ws-unsubscribe-post)
- **Auth**: Bearer Token (`Authorization: Bearer <token>`) or API Key (`X-Api-Key: <key>`)
- **Latency**: Sub-millisecond event propagation
- **Connection Limit**: `One active connection per IP address (enforced)`

## Quick Start

Connect to the global channel (default) to start receiving events immediately.

### 1. Get your Token
Obtain a Bearer token or API Key from your administrator or refer to [API Reference](/docs/devops/restapi/reference/api_keys).

### 2. Connect
```bash
wscat -H "Authorization: Bearer <YOUR_TOKEN>" -c "ws://<YOUR_HOST>/api/v1/ws/subscribe?event_types=pdfm,health,orchestrator"
```

### 3. Verify
You will receive a confirmation message:
```json
{
  "id":"985763c36f1a7f04663353ceb5515afaa497414ebec3226e9fa0d475b7713e34","event_type":"global",
  "timestamp":"2025-12-11T13:04:37.484843Z",
  "message":"WebSocket connection established subscribed to global by default",
  "body":{
    "client_id":"338a28f7-666e-46e1-8e7d-4feba5edc061",
    "subscriptions":["pdfm","health","orchestrator","global"]
  }
}
```

## Core Concepts

### Event Flow
<div class="mermaid">
sequenceDiagram
    participant Client
    participant Server
    Client->>Server: Connect (Bearer Token)
    Server-->>Client: Connection Established (Global Channel)
    Client->>Server: Subscribe (pdfm, health)
    Server-->>Client: Subscription Confirmed
    loop Heartbeat (Every 30s)
        Client->>Server: Ping {"event_type": "health"}
        Server-->>Client: Pong
    end
    Server-->>Client: Event: VM_STATE_CHANGED
</div>

### Event Channels

Channels are topics you can subscribe to. You can specify multiple channels 
during connection.

| Channel | Description | Use Case |
| :--- | :--- | :--- |
| **`global`** | System-wide broadcasts. **Auto-subscribed**; cannot be removed. | Critical maintenance alerts |
| **`pdfm`** | Host & VM events | VM lifecycle monitoring |
| **`orchestrator`** | Cluster-level orchestration events | Task assignments, cluster scaling |
| **`health`** | Ping/Pong heartbeat channel | Connection liveness checks |
| **`system`** | System & metadata info | Retrieving your unique `client_id` |

### Connection Rules

- **Single Connection per IP**: To prevent resource exhaustion, only one 
  active WebSocket connection is allowed per client IP address. New connection 
  attempts from the same IP will be rejected with `409 Conflict`.


## Troubleshooting & FAQ

**Q: I get a `409 Conflict` error when connecting.**

A: You likely have another active connection from the same IP. Close the 
existing connection or check for "zombie" processes.

**Q: Can I unsubscribe from `global`?**

A: No, the `global` channel is mandatory for system-wide alerts.

**Q: How to keep the connection alive for long time?**

To keep the WebSocket connection alive for an extended period, consider implementing a heartbeat mechanism where the client periodically sends ping messages to the server. This helps prevent timeouts and ensures that the connection remains active. Refer to the [How to Send Heartbeats]({{ site.url }}{{ site.baseurl }}/docs/devops/events-hub/how-to-send-heartbeat) guide for detailed instructions.

**Q: I'm not receiving pongs.**
A: Ensure you have subscribed to the `health` channel in your connection URL 
(`?event_types=health`).

**Q: My connection drops after 60 seconds.**

A: Check if your client is sending heartbeats. Some load balancers drop idle 
WebSocket connections.

**Q: I'm not receiving VM state change events.**

A: Ensure you're subscribed to the `pdfm` channel and that VMs are actually 
changing state. Check the VM exists and you have proper permissions.

**Q: The unsubscribe request fails with 403.**

A: The unsubscribe request must be sent from the same authenticated user/login 
that created the WebSocket subscription. Additionally, make sure the `client_id` 
in your unsubscribe request matches your actual client ID. Request your client 
ID first using the system channel if needed.

