---
layout: page
title: Events Hub
subtitle: How to Send Heartbeats
menubar: docs_devops_menu
show_sidebar: false
toc: true
---
# How to Send Heartbeats
The Events Hub WebSocket service supports heartbeat messages to keep connections alive and monitor client activity. Heartbeats are periodic messages sent by the client to the server to indicate that the connection is still active.

<div class="mermaid">
flowchart TD
    A[Start Connection] --> B{Connected?}
    B -- Yes --> C[Wait 30 Seconds]
    C --> D[Send Ping]
    D --> E{Receive Pong?}
    E -- Yes --> C
    E -- No --> F[Reconnect]
    B -- No --> F
</div>

## Steps to Send Heartbeats
1. **Obtain Your Token**: Get a Bearer token or API Key from your administrator, or refer to the [API reference]({{ site.url }}{{ site.baseurl }}/docs/devops/restapi/reference/api_keys).
2. **Connect to the Events Hub**: For simplicity, the examples below use `wscat` (a popular WebSocket client) to connect and subscribe to the `health` event channel. You can use any WebSocket client of your choice. Refer to the REST API [reference]({{ site.url }}{{ site.baseurl }}/docs/devops/restapi/reference/events/#v1-ws-subscribe-get) if you want to use another client.

   - Install `wscat` (if not already installed):

     ```bash
     brew install node
     npm install -g wscat
     ```

   - Connect and subscribe to the `health` channel:

     ```bash
     wscat -H "Authorization: Bearer <YOUR_TOKEN>" -c "ws://<YOUR_HOST>/api/v1/ws/subscribe?event_types=health"
     ```
     - Note: Use `wss://` instead of `ws://` if your server is configured for TLS.
3. **Send Heartbeat Messages**: Once connected, you can send heartbeat messages at regular intervals. A heartbeat message is a simple JSON object with the following structure:
    ```json
    {"event_type": "health","message":"ping"}
    ```
    You can send this message using your WebSocket client. For example, in `wscat`, you can type the JSON message and press Enter to send it.

4. **Receive Pong Responses**: The server will respond to heartbeat messages with a pong message, confirming that the connection is still active. A typical pong response looks like this:
    ```json
    {
        "id":"1495c450ab281318efc6b63482fde92630b66c0bf66acd5073fdc3a44402be67","ref_id":"cbd91850fad5fc286caa89916b081c9d34c4b7bde2ce4ce4392f47527a9cc27",
        "event_type":"health",
        "timestamp":"2025-12-12T15:57:55.570578Z",
        "message":"pong",
        "client_id":"d24c0350-936d-4c7a-8948-36f42343c87c"
    }
    ```
5. **Automate Heartbeats**: To ensure consistent heartbeat messages, consider automating the sending of heartbeats using a script or a WebSocket client that supports automatic ping/pong functionality.
6. **Monitor Connection Health**: Regularly monitor the connection health by checking for pong responses. If you do not receive a pong response within a specified timeout period, consider reconnecting to the Events Hub.
   
By following these steps, you can effectively send heartbeat messages to the Events Hub WebSocket service, ensuring that your connection remains active and healthy.

### FAQ
#### How often should I send heartbeat messages?
It is recommended to send heartbeat messages every 30 to 60 seconds, but this can vary based on your application's requirements.
#### What happens if I don't send heartbeat messages?
If heartbeat messages are not sent, the server may consider the connection inactive and close it after a timeout period.
#### Can I customize the heartbeat message?
No, the heartbeat message format should remain consistent to ensure proper communication with the server.
#### I am not receiving pong responses. What should I do?
Check your subscribe request, You must subscribe to the `health` event channel to receive pong responses. If the issue persists, consider reconnecting to the Events Hub.

Refer to the troubleshooting section in the [overview guide]({{ site.url }}{{ site.baseurl }}/docs/devops/events-hub/overview#troubleshooting--faq) for common issues.
