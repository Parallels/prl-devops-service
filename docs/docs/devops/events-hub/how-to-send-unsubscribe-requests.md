---
layout: page
title: Events Hub
subtitle: How To Send Unsubscribe Requests for Events channel
menubar: docs_devops_menu
show_sidebar: false
toc: true
---

# How to Send Unsubscribe Requests for Events channel
The Events Hub WebSocket service allows clients to unsubscribe from specific event channels when they no longer wish to receive events from those channels. This is done by sending an unsubscribe request to the server. Importantly, the global channel is subscribed by default and cannot be unsubscribed.

> Unsubscribe requests must be sent by the same user who made the subscription.

## Steps to Send Unsubscribe Requests
1. **Obtain Your Token**: Get a Bearer token or API Key from your administrator, or refer to the [API reference]({{ site.url }}{{ site.baseurl }}/docs/devops/restapi/reference/api_keys).
2. **Send a REST Request**: For unsubscribing from specific event channels, you need to send a POST request to the unsubscribe endpoint [/api/v1/ws/unsubscribe]({{ site.url }}{{ site.baseurl }}/docs/devops/restapi/reference/events/#_api_v1_ws_unsubscribe_post). You can use tools like `curl` or Postman to send this request. 
   
    - Install `curl` (if not already installed):

        ```bash
        brew install curl
        ```
    - Use the following `curl` command to send an unsubscribe request:        
        ```bash
        curl -X POST "http://localhost:5680/api/v1/ws/unsubscribe" \
        -H "Authorization: Bearer {your_token_here}" \
        -H "Content-Type: application/json" \
        -d '{
          "client_id": "3ce40e66-fd3c-479c-a755-5a4e1917a715",
          "event_types": ["pdfm", "orchestrator"]
        }'
        ```
        - Replace `{your_token_here}` with your actual Bearer token or API Key.
        - Replace the `client_id` value with your actual Client ID. You can find your Client ID by following  the steps in the [How to Know My Client ID]({{ site.url }}{{ site.baseurl }}/docs/devops/events-hub/how-to-know-my-client-id) guide.
        - Replace the `event_types` array with the list of event channels you want to unsubscribe from.
   
3. **Verify Unsubscription**: After sending the unsubscribe request, you should receive a return code of `200 OK` indicating that the request was successful. You can also monitor your WebSocket connection to ensure that you no longer receive events from the unsubscribed channels.
4. **Handle Errors**: If the unsubscribe request fails, you will receive an error response with details about the failure. Common errors include invalid Client ID or attempting to unsubscribe from channels that you are not subscribed to. Review the error message and adjust your request accordingly.
5. **Reconnect if Necessary**: If you need to subscribe to different event channels after unsubscribing, you need to disconnect and reconnect to the Events Hub WebSocket service with the new subscription parameters.
6. By following these steps, you can effectively manage your event subscriptions in the Events Hub WebSocket service by sending unsubscribe requests as needed.

### FAQ
#### What if I lose my Client ID?
You can find your Client ID by following the steps in the [How to Know My Client ID]({{ site.url }}{{ site.baseurl }}/docs/devops/events-hub/how-to-know-my-client-id) guide.
#### Can I unsubscribe from multiple event channels in a single request?
Yes, you can specify multiple event channels in the `event_types` array with commas when sending the unsubscribe request.
#### What happens if I try to unsubscribe from a channel I am not subscribed to?
You will receive an error response indicating that you are not subscribed to that channel. Ensure you only unsubscribe from channels you are currently subscribed to.
#### Do I need to reconnect after unsubscribing?
No, unsubscribing does not require you to disconnect from the WebSocket service. However, if you want to subscribe to new channels, you will need to reconnect with the updated subscription parameters.
#### How can I confirm that I have successfully unsubscribed?
You can monitor your WebSocket connection to ensure that you no longer receive events from the unsubscribed channels. Additionally, a successful unsubscribe request will return a `200 OK` status code.
#### Is there a limit to the number of event channels I can unsubscribe from at once?
There is no specific limit mentioned.
#### Can I unsubscribe to global channel?
No, the global channel is subscribed by default and cannot be unsubscribed.

Refer to the troubleshooting section in the [overview guide]({{ site.url }}{{ site.baseurl }}/docs/devops/events-hub/overview#troubleshooting--faq) for common issues.

