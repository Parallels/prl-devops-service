---
layout: page
title: RestApi
subtitle: Overview
menubar: docs_devops_menu
show_sidebar: false
toc: false
---

# What is It?

Managing Parallels Desktop hosts remotely has been a puzzle for a while now, but we've come up with an exciting solution that we're thrilled to share. We understand that not everyone is comfortable working with the command line interface and writing scripts, so we've developed something simpler.

REST API is a widely used protocol that makes it easy for users of all levels to create, delete, manage, configure, and execute commands remotely on any enabled host. Our API is a game-changer for those managing remote installations or deploying new clean VMs for a CI/CD pipeline.

But we didn't stop there. We've also created simple and powerful documentation using Swagger, which allows users to experiment with different endpoints and test the API.

<div class="flex-center"><img src="{{ site.url }}{{ site.baseurl }}/assets/img/restapi_swagger.png"></div>

To ensure the security of your hosts, we've included a robust RBAC system. This system provides granular access to each VM in the host, giving you peace of mind that your security is not compromised.

# How Does It Work?

The RestAPI is built around our powerful command line interface, exposing some of the functionality using the RESTful programming interface. To control the Parallels Desktop host you will need to have the Parallels Desktop Pro or Business installed on the host machine, if you don't have it, the DevOps service can help you with that, check our [installation guide]({{ site.url }}{{ site.baseurl }}/getting-started/installation/).

Once this is done, you can start the DevOps service and by default it will enable the RestAPI. You can then start using the RestAPI to control the host and the VMs.

# How difficult is it to set up?

Out of the gate, the RestAPI is enabled by default when you start the DevOps service. You can then start using the RestAPI to control the host and the VMs by just using the default settings. You then might want to configure the RBAC system to control the access to the RestAPI.
