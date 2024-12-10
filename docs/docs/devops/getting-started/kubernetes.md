---
layout: page
title: Getting Started
subtitle: Kubernetes
menubar: docs_devops_menu
show_sidebar: false
toc: true
---

If you want to run only the DevOps service [Catalog]({{ site.url }}{{ site.baseurl }}/docs/catalog/overview/), [Orchestrator]({{ site.url }}{{ site.baseurl }}/docs/orchestrator/overview/) or the [Reverse Proxy]({{ site.url }}{{ site.baseurl }}/docs/reverse-poxy/overview/) in a kubernetes cluster we provide a helm chart. This allows you to quickly spin up the service in a cluster by just passing the configuration options as values.

## Prerequisites

- [Helm](https://helm.sh/)
- [Kubernetes](https://kubernetes.io/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

## Running the DevOps Service

### Add the Helm Repository

To add the helm repository run the following command:

```powershell
helm repo add prl-devops https://parallels.github.io/prl-devops-service/charts
```

once the repository is added you can install the chart

### Install the Chart

To install the chart run the following command:

```powershell
helm install prl-devops prl-devops/prl-devops-service
```

you can also pass a chart configuration file to the install command:

```powershell
helm install prl-devops prl-devops/prl-devops-service -f values.yaml
```

## Configuration values.yaml

Below is an example of a configuration file.

```yaml
replicaCount: 1

image:
  repository: cjlapao/prl-devops-service
  pullPolicy: IfNotPresent
  tag: "latest"

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

serviceAccount:
  create: true
  automount: true
  name: ""

podAnnotations: {}
podLabels: {}

podSecurityContext: {}
  # fsGroup: 2000

securityContext: {}
  # capabilities:
  #   drop:
  #   - ALL
  # readOnlyRootFilesystem: true
  # runAsNonRoot: true
  # runAsUser: 1000

storage:
  node_name: ''
  databasePath: '/go/bin/db'

security:
  jwt:
    rsa_private_key: ""
    hmac_secret: ""
    duration: "15m"
    signing_method: "HS256"
  password:
    min_password_length: 12
    max_password_length: 40
    require_uppercase: true
    require_lowercase: true
    require_numbers: true
    require_special_characters: true
    salt_password: true
  brute_force:
    max_login_attempts: 5
    lockout_duration: 10s
    increment_lockout_duration: true
  encryption_private_key: ""
  root_password: ""

service:
  type: ClusterIP
  port: 80
  targetPort: 80

logLevel: info

config:
  mode: "api"
  disableCatalogCaching: false

ingress:
  istio: false
  prefix: parallels
  enabled: true
  annotations: {}
  gateway: ""
  host: ""
  apiPort: 80
  apiPrefix: /api
  tls:
    enabled: false
    port: 443
    certificate:
    privateKey:
    

resources: {}
  # We usually recommend not to specify default resources and to leave this as a conscious
  # choice for the user. This also increases chances charts run on environments with little
  # resources, such as Minikube. If you do want to specify resources, uncomment the following
  # lines, adjust them as necessary, and remove the curly braces after 'resources:'.
  # limits:
  #   cpu: 100m
  #   memory: 128Mi
  # requests:
  #   cpu: 100m
  #   memory: 128Mi

autoscaling:
  enabled: false
  minReplicas: 1
  maxReplicas: 100
  targetCPUUtilizationPercentage: 80

nodeSelector: {}

tolerations: []

affinity: {}
```
