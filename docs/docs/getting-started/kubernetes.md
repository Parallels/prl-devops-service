---
layout: page
title: Getting Started
subtitle: Kubernetes
menubar: docs_menu
show_sidebar: false
toc: true
---

If you want to run only the DevOps service [Catalog]({{ site.url }}{{ site.baseurl }}/docs/catalog/overview/), [Orchestrator]({{ site.url }}{{ site.baseurl }}/docs/orchestrator/overview/) or the [Reverse Proxy]({{ site.url }}{{ site.baseurl }}/docs/reverse-poxy/overview/) in a kubernetes cluster we provide a helm chart. This allows you to quickly spin up the service in a cluster by just passing the configuration options as values.

## Prerequisites

- [Helm](https://helm.sh/)
- [Kubernetes](https://kubernetes.io/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

## Running the DevOps Service

for a quick start the `DevOps` service you can run:

```powershell
helm repo add prl https://cjlapao.github.io/prl-devops-helm/
```