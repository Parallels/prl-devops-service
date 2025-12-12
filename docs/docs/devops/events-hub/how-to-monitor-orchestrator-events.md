---
layout: page
title: Events Hub
subtitle: How to Monitor Orchestrator Events
menubar: docs_devops_menu
show_sidebar: false
toc: true
---
# How to Monitor Orchestrator Events
To monitor orchestrator events in real-time, you can subscribe to the `orchestrator` event channel using the Events Hub WebSocket service. This allows you to receive instant notifications about cluster-level orchestration events, such as host health changes, host VM status updates, etc.

<div class="mermaid">
sequenceDiagram
    participant Client
    participant Server
    Client->>Server: Connect & Subscribe (orchestrator)
    Server-->>Client: Subscription Confirmed
    Note over Server: Host Health Update
    Server-->>Client: Event: HOST_HEALTH_UPDATE
</div>

## Steps to Monitor Orchestrator Events
1. **Obtain Your Token**: Get a Bearer token or API Key from your administrator, or refer to the [API reference]({{ site.url }}{{ site.baseurl }}/docs/devops/restapi/reference/api_keys).
2. **Connect to the Events Hub**: 
   You can subscribe to multiple event channels during connection, like `orchestrator,health`, to receive both orchestrator events and health check events. For this example, we are subscribing to `orchestrator`.
   
   For simplicity, the examples below use `wscat` (a popular WebSocket client) to connect and subscribe to orchestrator events. You can use any WebSocket client of your choice. Refer to the REST API [reference]({{ site.url }}{{ site.baseurl }}/docs/devops/restapi/reference/events/#v1-ws-subscribe-get) if you want to use another client.

    - Install `wscat` (if not already installed):
        ```bash
        brew install node
        npm install -g wscat
        ```
    - Connect and subscribe to orchestrator events on the `orchestrator` channel:
        ```bash 
        wscat -H "Authorization: Bearer <YOUR_TOKEN>" -c "ws://<YOUR_HOST>/api/v1/ws/subscribe?event_types=orchestrator"
        ```
        - Note: Use `wss://` instead of `ws://` if your server is configured for TLS.

3. **Verify Subscription**: Upon successful connection, you will receive a confirmation message indicating that you are subscribed to the `orchestrator` channel.
Example confirmation message:
```json
 {
     "id": "985763c36f1a7f04663353ceb5515afaa497414ebec3226e9fa0d475b7713e34",
     "event_type": "global",
     "timestamp": "2025-12-11T13:04:37.484843Z",
     "message": "WebSocket connection established; subscribed to global by default",
     "body": {
         "client_id": "338a28f7-666e-46e1-8e7d-4feba5edc061",
         "subscriptions": ["orchestrator", "global"]
     }
 }
```

## Example Orchestrator Events
Here is an example orchestrator event when a host VM status changes from resuming to running:
```json
{
    "id":"7cbc843fd99e2965fd4039410a0e20c2984b8b9c8889991954751d66b90cabb1",
    "event_type":"orchestrator",
    "timestamp":"2025-12-12T12:12:14.164146Z",
    "message":"HOST_VM_STATE_CHANGED","body":{
        "host_id":"a16c7dfc2e8e79b927f0d8e786c7ac5fce172d800be06c15cb789eabae1550de",
        "event":{
            "previous_state":"resuming",
            "current_state":"running",
            "vm_id":"ae032783-53d3-49c3-babf-1475589d9a7b"
            }
    }
}
```

Here is an example orchestrator event when a host VM is added:
```json
{
    "id":"8d9e0f1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0t1u2v3w4x5y6z7a8b9",
    "event_type":"orchestrator",
    "timestamp":"2025-12-12T14:20:30.987654Z",
    "message":"HOST_VM_ADDED",
    "body":{
        "host_id":"b27d8e9f00112233445566778899aa4f5e6d7c8b9a0b1c2d3e4f5061728394a",
        "event":{
            "vm_id":"ce143894-64e4-50d4-cbcf-2586690e0b8c",
            "new_vm": {
                "name": "New Host VM",
                "os_type": "windows",
                "state": "stopped"
            }
        }
    }
}
```

Here is an example orchestrator event when a host VM is removed:
```json
{
    "id":"1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0t1u2v3w4x5y6z7a8b9c0d1e2",
    "event_type":"orchestrator",
    "timestamp":"2025-12-12T15:30:45.123456Z",
    "message":"HOST_VM_REMOVED",
    "body":{
        "host_id":"b27d8e9f00112233445566778899aa4f5e6d7c8b9a0b1c2d3e4f5061728394a",
        "event":{
            "vm_id":"ce143894-64e4-50d4-cbcf-2586690e0b8c"
        }
    }
}
```

Here is an example orchestrator event when host health is changed from healthy to unhealthy:
```json
{
    "id":"4f5e6d7c8b9a0b1c2d3e4f5061728394a5b6c7d8e9f00112233445566778899aa",
    "event_type":"orchestrator",
    "timestamp":"2025-12-12T13:15:20.123456Z",
    "message":"HOST_HEALTH_UPDATE",
    "body":{
        "host_id":"b27d8e9f00112233445566778899aa4f5e6d7c8b9a0b1c2d3e4f5061728394a",
        "state":"unhealthy",
    }
}
```

### JSON Object Schemas
On receiving orchestrator events, the `body` field will contain different structures based on the event type. Below are the JSON object schemas for common orchestrator events:

On host health update event:
```go
type HostHealthUpdate struct {
	HostID string `json:"host_id"`
	State  string `json:"state"`
}
```
On host VM state change event:
```go
type HostVmEvent struct {
	HostID string      `json:"host_id"`
	Event  interface{} `json:"event"` //        VmStateChange, VmAdded, or VmRemoved
}
```
### FAQ
#### What types of orchestrator events can I monitor?
You can monitor various orchestrator events such as host health updates and host VM state changes (e.g., started, stopped, suspended, resumed, etc.).
#### How do I filter specific orchestrator events?
When subscribing to the Events Hub, you can specify the `orchestrator` event type to receive only orchestrator-related events.
#### Can I use other WebSocket clients besides `wscat`?
Yes, you can use any WebSocket client that supports custom headers for authentication. Just ensure you include the appropriate authorization header.
#### Is it possible to monitor events for multiple hosts?
Yes, by subscribing to the `orchestrator` event type, you will receive events for all hosts managed by the server.
#### My host is running but I am not receiving any events, why?
Ensure that your Bearer token or API Key is valid and has the necessary permissions to access the Events Hub. Check for typos in the token and verify that the server URL is correct. Also, ensure that your network allows WebSocket connections to the specified host and port.
#### I am not receiving events but REST API calls are working fine, why?
Your host might be running on old version that does not support Events Hub. Ensure that your Host is updated to the latest version that includes Events Hub support.
                        **OR**
There is an issue with the WebSocket connection. Check your network settings, firewall rules, and ensure that the WebSocket endpoint is reachable from your client.
                        **OR**
There might be a temporary issue with the Events Hub service on the server. Check the server logs for any errors related to the Events Hub and consider restarting the service if necessary.
#### Where can I find troubleshooting tips?
Refer to the troubleshooting section in the [overview guide]({{ site.url }}{{ site.baseurl }}/docs/devops/events-hub/overview#troubleshooting--faq) for common issues.

