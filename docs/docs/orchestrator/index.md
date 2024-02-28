---
layout: page
title: Orchestrator
subtitle: Manage and orchestrate multiple Parallels Desktop Api Services
show_sidebar: false
toc: false
hero_link: /docs/orchestrator/overview/
hero_link_text: Documentation
---

# What is It?

Our Orchestrator service is an orchestrator system for Parallels Desktop,
allowing you to manage and deploy a pool of hosts and their VMs,
including different architectures.

Managing all of this from a single pane of glass greatly reduces the complexity
of managing build systems at scale, allowing it to reduce maintenance costs.

You can pair it with our Catalog Service to cache base images from a secure
repository and increase the speed of deploying new images in that same orchestrator.

# How Does It Work?

<div class="flex-center"><img src="../../../img/devtools_service-orchestrator.drawio.png"></div>
The Orchestrator Service is a cloud-based tool that helps manage a group of
hosts (pool) in a single location. This service allows an administrator to
create virtual machines on the farm with ease, without worrying about the
underlying infrastructure or where the machines can be created.

When paired with the catalog, the entire process can be automated, similar to
the scenarios we see in the CI/CD pipelines. This means you can have a pool of
hosts in any location, whether in your cloud or current providers. You can then
use the Orchestrator Service to manage the resources, allowing virtual machines
to be created and destroyed as needed.

# How difficult is it to set up?

To set up the orchestrator, you simply need to start your regular API, which
will become accessible. There is no need for any additional setup. Once it is
operational, you can easily include the necessary hosts in your system. However,
please note that each node in the system must have the RestAPI running to be
added to the farm.

# What are we trying to solve?

Managing CI/CD pipelines is a complicated process requiring a multi-disciplined
team of developers and operations to put in much effort. Even after all the hard
work, pipelines can still take a long time to run and often leave behind
"residues" that could impact further runs. However, having a process that allows
for easy and straightforward management of a group of nodes, where you can spawn
a virtual machine based on a pre-configured image with all the necessary software
for a specific task, can significantly reduce this complexity. Adding a security
layer to this process means you can easily revoke access to a specific template
anytime, increasing the DevOps teams' confidence.

# Example Architecture for CI/CD Automation

<div class="flex-center"><img src="../../../img/devtools_service-orchestrator_ci_cd_github_scenario_square.drawio.png"></div>