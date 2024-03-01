---
layout: page
title: RestApi
subtitle: Overview
menubar: docs_menu
show_sidebar: false
toc: false
---

# What is It?

Managing Parallels Desktop hosts remotely has been a puzzle for a while now, but we've come up with an exciting solution that we're thrilled to share. We understand that not everyone is comfortable working with the command line interface and writing scripts, so we've developed something simpler. 

REST API is a widely used protocol that makes it easy for users of all levels to create, delete, manage, configure, and execute commands remotely on any enabled host. Our API is a game-changer for those managing remote installations or deploying new clean VMs for a CI/CD pipeline. 

<div class="flex-center"><img src="../../../assets/img/restapi_swagger.png"></div>

But we didn't stop there. We've also created simple and powerful documentation using Swagger, which allows users to experiment with different endpoints and test the API. To ensure the security of your hosts, we've included a robust RBAC system. This system provides granular access to each VM in the host, giving you peace of mind that your security is not compromised.

It also comes with a robust RBAC system, allowing the users to manage their hosts with security in mind, giving granular access to each VM in the host.

# How Does It Work?

The RestAPI is built around our powerful command line interface, exposing some of the functionality using the RESTful programming interface. To start it you need to download it from our public repo and start it.

If you don’t have Parallels Desktop installed, it can also help you with that, allowing you to have a single starting point when you want to perform a brand-new installation.
to be created and destroyed as needed.

# How difficult is it to set up?

You just need the RestAPI binary and a Parallels Desktop installation. Once you have both, you can start the RestAPI and start using it. It is that simple.

# What are we trying to solve?

Remote management of Virtual Machines has always been challenging in the market and non-existent in Parallels Desktop, making some integration with 3rd parties difficult.

It also hindered us from providing solutions to some of the pain points for using Parallels Desktop as a tool for developers and DevOps flows, where remote management of resources is a must in today’s CI/CD environments.

With the RESTful API, we open the doors to more advanced tools and integrations, for example, with Terraform, with a Terraform provider or with GitHub Actions with a custom action, allowing us to target markets we could not beforehand and new partnerships with 3rd party suppliers.

The most significant growing market now would be Continuous Integration and Continuous Deployment (CI/CD). Our solution will provide seamless integration and automation functionalities through our API, allowing developers and operations to streamline workflows and accelerate release cycles and testability of their software. This market has seen a substantial rise in the last few years.