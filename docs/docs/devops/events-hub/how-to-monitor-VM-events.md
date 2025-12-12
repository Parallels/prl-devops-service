---
layout: page
title: Events Hub
subtitle: How to Monitor VM Events
menubar: docs_devops_menu
show_sidebar: false
toc: true
---
# How to Monitor VM Events

To monitor VM events in real-time, you can subscribe to the `pdfm` event channel using the Events Hub WebSocket service. This allows you to receive instant notifications about VM lifecycle changes, such as when a VM is started, stopped, or deleted.

<div class="mermaid">
sequenceDiagram
    participant Client
    participant Server
    Client->>Server: Connect & Subscribe (pdfm)
    Server-->>Client: Subscription Confirmed
    Note over Server: VM State Changes (e.g. Running -> Stopped)
    Server-->>Client: Event: VM_STATE_CHANGED
</div>

## Steps to Monitor VM Events

1. **Obtain Your Token**: Get a Bearer token or API Key from your
    administrator, or refer to the
    [API reference]({{ site.url }}{{ site.baseurl }}/docs/devops/restapi/reference/api_keys).
2. **Connect to the Events Hub**:
    You can subscribe to multiple event channels during connection, like `pdfm,health`, to receive both VM events and health check events. For this example, we are subscribing to `pdfm`.
    
    For simplicity, the examples below use
    `wscat` (a popular WebSocket client) to connect and subscribe to VM events.
    You can use any WebSocket client of your choice. Refer to the REST API
    [reference]({{ site.url }}{{ site.baseurl }}/docs/devops/restapi/reference/events/#v1-ws-subscribe-get)
    if you want to use another client.

     - Install `wscat` (if not already installed):

         ```bash
         brew install node
         npm install -g wscat
         ```

     - Connect and subscribe to VM events on the `pdfm` channel:

         ```bash
         wscat -H "Authorization: Bearer <YOUR_TOKEN>" -c "ws://<YOUR_HOST>/api/v1/ws/subscribe?event_types=pdfm"
         ```

         - Note: Use `wss://` instead of `ws://` if your server
             is configured for TLS.
3. **Verify Subscription**: Upon successful connection, you will receive a
    confirmation message indicating that you are subscribed to the `pdfm`
    channel.

    Example confirmation message:

    ```json
     {
         "id": "985763c36f1a7f04663353ceb5515afaa497414ebec3226e9fa0d475b7713e34",
         "event_type": "global",
         "timestamp": "2025-12-11T13:04:37.484843Z",
         "message": "WebSocket connection established; subscribed to global by default",
         "body": {
             "client_id": "338a28f7-666e-46e1-8e7d-4feba5edc061",
             "subscriptions": [
                 "pdfm",
                 "global"
             ]
         }
     }
    ```

4. **Monitor VM Events**: You will start receiving real-time VM events. Each
    event contains details about VM state changes (e.g., when a VM starts,
    stops, or is deleted).

    Example VM state change event:

    ```json
     {
         "id": "60adab5b984894719ecb83a0ad7da3d2027a63932c65d0f0585f3762083aa4a1",
         "event_type": "pdfm",
         "timestamp": "2025-12-11T13:21:21.514683Z",
         "message": "VM_STATE_CHANGED",
         "body": {
             "previous_state": "suspended",
             "current_state": "resuming",
             "vm_id": "fb83295e-0ccb-46b1-8205-b03997f82b00"
         }
     }
    ```

### JSON Object Schemas

On VM state change event:

```go
type VmStateChange struct {
    PreviousState string `json:"previous_state"`
    CurrentState  string `json:"current_state"`
    VmID          string `json:"vm_id"`
}
```

On VM added event:

```json
{
    "id": "7cbc843fd99e2965fd4039410a0e20c2984b8b9c8889991954751d66b90cabb1",
    "event_type": "pdfm",
    "timestamp": "2025-12-12T12:12:14.164146Z",
    "message": "VM_ADDED",
    "body": {
        "vm_id": "ae032783-53d3-49c3-babf-1475589d9a7b",
        "new_vm": {
            "name": "My New VM",
            "os_type": "ubuntu",
            "state": "running"
        }
    }
}
```

On VM removed event:

```json
{
    "id": "4f5e6d7c8b9a0b1c2d3e4f5061728394a5b6c7d8e9f00112233445566778899aa",
    "event_type": "pdfm",
    "timestamp": "2025-12-12T13:15:20.123456Z",
    "message": "VM_REMOVED",
    "body": {
        "vm_id": "ae032783-53d3-49c3-babf-1475589d9a7b"
    }
}
```

### FAQ
#### What types of VM events can I monitor?
You can monitor various VM events such as state changes (started, stopped, suspended, resumed..etc), additions, and deletions.
#### Can I use other WebSocket clients besides `wscat`?
Yes, you can use any WebSocket client that supports custom headers for authentication. Just ensure you include the appropriate authorization header.
#### Is it possible to monitor events for multiple VMs?
Yes, by subscribing to the `pdfm` event type, you will receive events for all VMs managed by the server.

Refer troubleshooting section in the [overview guide]({{ site.url }}{{ site.baseurl }}/docs/devops/events-hub/overview#troubleshooting--faq) for common issues.
