---
layout: page
title: Orchestrator
subtitle: Overview
menubar: docs_devops_menu
show_sidebar: false
toc: true
---

# Overview

## What is It?

Managing CI/CD pipelines is a complex process that requires a team of developers and operations to put in a lot of effort. Even after all the hard work, pipelines can still take a long time to run and often leave behind "residues" that could affect further runs. However, having a process that enables easy and straightforward management of a group of nodes, where you can spawn a virtual machine based on a pre-configured image with all the necessary software for a specific task, can significantly reduce this complexity. Adding a security layer to this process means you can easily revoke access to a specific template anytime, increasing the confidence of the DevOps team.

Our Orchestrator service is an orchestrator system for Parallels Desktop, which allows you to manage and deploy a pool of hosts and their VMs, including different architectures. Managing all of this from a single pane of glass greatly reduces the complexity of managing those build systems at scale, making it possible to reduce maintenance costs. You can pair it with our Catalog Service to cache base images from a secure repository and increase the speed of deploying new images in that same orchestrator.

### Example Architecture for CI/CD Automation

![Catalog Pulling Diagram]({{ site.url }}{{ site.baseurl }}/img/orchestrator/orchestrator_ci_cd_github_scenario_square.drawio.png)

## Architecture

The Parallels Desktop Orchestrator Service is written in Go and uses the same base code as the Parallels Desktop API Service. It is a single executable that behaves differently depending on how it is run, making deployment simpler. The service has a straightforward architecture, consisting of a collection of hosts that connect to a Parallels Desktop API instance. The orchestrator starts by running a background service that monitors the status of each host, recording any changes it detects such as available resources, health state and running virtual machines. You can manage each host individually by creating, starting, stopping and deleting virtual machines, or you can let the orchestrator do the work for you by creating a virtual machine and allowing the orchestrator to choose the host with enough resources to run it.

![Catalog Pulling Diagram]({{ site.url }}{{ site.baseurl }}/img/orchestrator/orchestrator.drawio.png)

## How Does It Work?

The Orchestrator Service is a cloud-based tool that simplifies the management of a group of hosts (pool) in a single location. With this service, the administrator can effortlessly create virtual machines on the farm without worrying about the underlying infrastructure or machine allocation.

When integrated with the catalog, the entire process can be automated, making it similar to the scenarios observed in CI/CD pipelines. This means that you can have a pool of hosts in any location, be it your cloud or current providers. And, by using the Orchestrator Service, you can manage these resources effortlessly, allowing virtual machines to be created and destroyed as per the requirements.
