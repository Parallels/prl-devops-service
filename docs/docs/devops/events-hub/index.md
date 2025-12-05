---
layout: page
title: Events Hub Guide
show_sidebar: false
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

## Real-world Use Cases

### 1. Monitoring VM Lifecycle

Subscribe to `pdfm` to track VM states in real-time.

- **Event**: VM state changes (stopped → running)
- **Action**: Update dashboard status, trigger automation

### 2. Connection Liveness Checks

Subscribe to `health` to implement a heartbeat. If the server stops responding 
to your pings, you can trigger a reconnection alert.

### 3. Cluster Health Monitoring

Subscribe to `orchestrator` to receive host health updates and cluster-level 
VM state changes in real-time.

### 4. System Information

Use `system` channel to retrieve your client ID for unsubscribe operations.

## Event Flows and Examples

### Health Check Flow

The connection is bidirectional. To ensure the link is alive, your client 
**must** subscribe to the `health` channel and send periodic pings.

#### Complete Flow: Request → Response

**1. Client sends ping:**

```json
{
  "event_type": "health",
  "id": "ping-1",
  "message": "ping"
}
```

**2. Server responds with pong:**

```json
{
  "id":"8f01616715bc598b8b326d77212ee5cc589fb8d900a2dba7b0a7e2179144ae08",
  "ref_id":"ping-1",
  "event_type":"health","timestamp":"2025-12-11T13:15:58.302153Z","message":"pong",
  "timestamp":"2025-12-11T13:15:58.302153Z",
  "message":"pong",
  "client_id":"338a28f7-666e-46e1-8e7d-4feba5edc061"
}
```

**Expected behavior:** Send ping every 30 seconds to maintain connection 
liveness.

### VM State Change Flow

Subscribe to `pdfm` to receive real-time VM state notifications.

#### Complete Flow: Action → Event → Response

**1. Action:** User starts a virtual machine

**2. Server broadcasts state change:**

```json
{
  "id":"60adab5b984894719ecb83a0ad7da3d2027a63932c65d0f0585f3762083aa4a1",
  "event_type":"pdfm",
  "timestamp":"2025-12-11T13:21:21.514683Z",
  "message":"VM_STATE_CHANGED",
  "body":
  {
    "previous_state":"suspended",
    "current_state":"resuming",
    "vm_id":"fb83295e-0ccb-46b1-8205-b03997f82b00"
  }
}
```

**3. Client receives:** Event notification in real-time

**Expected behavior:** Update dashboard, trigger automation based on state 
changes.

### Orchestrator Event Flow

Subscribe to `orchestrator` to receive cluster-level events including host 
health updates and VM state changes across the cluster.

#### Complete Flow: Cluster Event → Broadcast → Response

**1. Event:** Host health check or VM state change in cluster

**2. Server broadcasts orchestrator event:**

**Host Health Update:**
```json
{
  "id":"db756a1c56b14cfff4c9be3ca87b42f72b508077bbd63b792f1b04e66fcf3379",
  "event_type":"orchestrator",
  "timestamp":"2025-12-11T17:00:06.910809Z",
  "message":"HOST_HEALTH_UPDATE",
  "body":
  {
    "host_id":"445fb9aefa6914be7d7575ba8fe96d7cebfe2b3c0001442b7628a9ae5fa588d5",
    "state":"healthy"
  }
}
```

**Host VM State Change:**
```json
{
  "id":"1995bc76d90d27471dc03313d34e54dd30ff56ea07fc351e729a4ff213d4dceb",
  "event_type":"orchestrator",
  "timestamp":"2025-12-11T16:24:04.271385Z",
  "message":"HOST_VM_STATE_CHANGED",
  "body":
  {
    "host_id":"445fb9aefa6914be7d7575ba8fe96d7cebfe2b3c0001442b7628a9ae5fa588d5",
    "event":{"previous_state":"suspending",
    "current_state":"suspended",
    "vm_id":"fb83295e-0ccb-46b1-8205-b03997f82b00"}
  }
}
```

**3. Client receives:** Cluster-level event notifications

**Expected behavior:** Monitor cluster health, track VM distribution across 
hosts, implement cluster-wide automation.

### Client ID Request Flow

After connecting, you might need your unique `client_id` for unsubscribe 
operations.

#### Complete Flow: Request → Response

**1. Client requests ID:**

```json
{
  "event_type": "system",
  "message": "client-id"
}
```

**2. Server responds with client ID:**

```json
{
  "event_type": "system",
  "message": "client-id",
  "body": {
    "client-id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Expected behavior:** Store the client ID for future unsubscribe requests.

## Endpoints

### Subscribe Endpoint

**GET** `/v1/ws/subscribe`

For complete API details, see: 
[Subscribe API Reference](/docs/devops/restapi/reference/events/#v1-ws-subscribe-get)

Upgrades the HTTP connection to WebSocket and subscribes to event notifications.

**Quick Example:**
```bash
wscat -H "Authorization: Bearer <TOKEN>" \
  -c "wss://api.example.com/v1/ws/subscribe?event_types=pdfm,health,orchestrator"
```

### Unsubscribe Endpoint

**POST** `/v1/ws/unsubscribe`

For complete API details, see: 
[Unsubscribe API Reference](/docs/devops/restapi/reference/events/#v1-ws-unsubscribe-post)

Unsubscribe active WebSocket client from specific event types without 
disconnecting. **Note**: You cannot unsubscribe from `global`.

**Example Request:**
```json
{
  "client_id": "550e8400-e29b-41d4-a716-446655440000",
  "event_types": ["pdfm"]
}
```

**Example Response:**
```json
{
  "message": "Successfully unsubscribed from event types",
  "remaining_subscriptions": ["global", "health"]
}
```

## Troubleshooting & FAQ

**Q: I get a `409 Conflict` error when connecting.**

A: You likely have another active connection from the same IP. Close the 
existing connection or check for "zombie" processes.

**Q: Can I unsubscribe from `global`?**

A: No, the `global` channel is mandatory for system-wide alerts.

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

