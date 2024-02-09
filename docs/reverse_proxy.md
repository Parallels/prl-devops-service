# Parallels Desktop DevOps Service Reverse Proxy

The Parallels Desktop DevOps Service Reverse Proxy is a service in our DevOps Suite
that allows traffic to be forwarded to a running Parallels Desktop Virtual Machine
from one external port to the internal port of the virtual machine.

This simplifies the process of managing networking traffic to and from the virtual
machines and allows for a more secure and controlled way of managing the traffic.

By default no external traffic is allowed to reach the virtual machines, the reverse
proxy will allow you to control which virtual machines can receive traffic and which
ports are allowed to be used.

## Architecture

The Parallels Desktop DevOps Service Reverse Proxy is written in go and uses the
same base code as the Parallels Desktop API Service. One single executable that depending
on how you run it it will behave in a different way, this allows for a simpler way
for deploying it.

The service itself is a very simple packet forwarder, it will listen for incoming
traffic and will forward it to the virtual machine that is running on the host.

It does distinguish between TCP traffic on a raw communication protocol for example
SSH, and HTTP traffic, for example a web server.

We also added some extra features to the HTTP traffic, like for example the ability
to have a custom CORS policy, and the ability to add a custom header to the request.

This is all controlled by a simple configuration file that will allow you to control
which virtual machines can receive traffic and which ports are allowed to be used.

## Configuration

We tried to make the configuration as simple as possible, the configuration for the
reverse proxy is a simple yaml object in the current configuration file.
If you do not use a configuration file, simple create one by looking at the example
in the [configuration section of the README.md](../README.md#configuration).

```yaml
reverse_proxy:
  port: 80
  host: localhost
  ssl:
    enabled: false
    cert: /etc/ssl/certs/localhost.crt
    key: /etc/ssl/private/localhost.key
  cors:
    enabled: true
    allowed_origins:
      - '*'
    allowed_methods:
      - GET
      - POST
      - PUT
      - DELETE
      - OPTIONS
    allowed_headers:
      - '*'
  hosts:
    - host: localhost
      port: 2202
      tcp_route:
        target_port: 22
        target_host: 10.0.0.1
    - host: localhost
      port: 8081
      http_routes:
        - path: /some_api
          scheme: http
          target_host: 10.0.0.2
          target_port: 8080
          response_headers:
            X-Frame-Options: DENY
```

## Getting Started

To start the reverse proxy you just need to start the binary and add the
`reverse_proxy` as the command-line argument.

```bash
./prldevops reverse_proxy
```
