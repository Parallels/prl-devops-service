---
layout: page
title: Events Hub
subtitle: How To Know My Client ID
menubar: docs_devops_menu
show_sidebar: false
toc: true
---
# How to Know My Client ID
When you connect to the Events Hub WebSocket service, each client is assigned a unique Client ID. This Client ID is essential for identifying your connection and managing events. Here’s how you can find out your Client ID.

## Steps to Know Your Client ID
1. First, ensure you have connected to the Events Hub WebSocket service and subscribed to the desired event channels. Refer to the [How to Send Heartbeats]({{ site.url }}{{ site.baseurl }}/docs/devops/events-hub/how-to-send-heartbeat) guide for connection instructions. When you send a ping message to the health channel, you will receive a pong response that includes your Client ID.
2. Once connected, send a heartbeat ping message to the `health` event channel. A heartbeat message is a simple JSON object like this:
    ```json
    {"event_type": "health","message":"ping"}
    ```
3. After sending the ping message, you will receive a pong response from the server. The pong response will contain your Client ID in the `client_id` field. A typical pong response looks like this:
    ```json
    {
        "id":"1495c450ab281318efc6b63482fde92630b66c0bf66acd5073fdc3a44402be67","ref_id":"cbd91850fad5fc286caa89916b081c9d34c4b7bde2ce4ce4392f47527a9cc27",
        "event_type":"health",
        "timestamp":"2025-12-12T15:57:55.570578Z",
        "message":"pong",
        "client_id":"d24c0350-936d-4c7a-8948-36f42343c87c"
    }
    ```
4. Locate the `client_id` field in the pong response to find your unique Client ID.
5. You can now use this Client ID for managing your connection and subscriptions within the Events Hub.

Alternatively, you can find your Client ID in the initial connection confirmation message you receive when you first connect to the Events Hub. This message includes a `body` section that contains the `client_id`.

You can also send a client info request to the server to retrieve your Client ID. Send a JSON message like this:

```json
{"event_type": "system","message":"client-id"}
```

The server will respond with a message containing your Client ID.
```json
{
    "id":"73841562a9753aa06a88623a1d5d47c5e831a6911e4a79d80910d495c4ecedd7","ref_id":"4232a7922d4c21f08f25c0da09b07b74a054d033948cbf53fbaf17808b45069f",
    "event_type":"system",
    "timestamp":"2025-12-15T09:30:57.021027Z",
    "message":"client-id",
    "body":
    {
        "client-id":"26369eda-4802-49f7-a07f-2292721d1724"
    },
    "client_id":"26369eda-4802-49f7-a07f-2292721d1724"
}
```

## FAQ
#### Can I change my Client ID?
No, the Client ID is automatically assigned by the Events Hub service and cannot be changed.
#### Is the Client ID persistent across sessions?
No, the Client ID is unique to each connection session. A new Client ID is assigned each time you connect to the Events Hub.
#### How is the Client ID used?
The Client ID is used if you want to unsubscribe from specific event channels or manage your subscriptions.
#### What if I lose my Client ID?
Send another heartbeat ping message to the `health` channel or a client info request to retrieve your Client ID again.
#### Can I use the Client ID for multiple connections?
No, each connection to the Events Hub has its own unique Client ID. You cannot use the same Client ID for multiple connections.
#### Where can I find more information about unsubscribing using Client ID?
Refer to the [How To Send Unsubscribe Requests for Events channel]({{ site.url }}{{ site.baseurl }}/docs/devops/events-hub/how-to-send-unsubscribe-requests) guide for detailed instructions on how to use your Client ID to unsubscribe from event channels.

Refer to the troubleshooting section in the [overview guide]({{ site.url }}{{ site.baseurl }}/docs/devops/events-hub/overview#troubleshooting--faq) for common issues.
