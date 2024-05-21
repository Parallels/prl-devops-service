---
title: Parallels Desktop DevOps Service
subtitle: Put your virtual machines to work on your CI/CD pipeline
layout: page
callouts: examples_callouts
is_home: true
---

# Parallels Desktop DevOps Service

[![License: Fair Source](https://img.shields.io/badge/license-fair-source.svg)](https://fair.io/)
[![Build](https://github.com/Parallels/prl-devops-service/actions/workflows/pr.yml/badge.svg)](https://github.com/Parallels/prl-devops-service/actions/workflows/pr.yml)
[![Publish](https://github.com/Parallels/prl-devops-service/actions/workflows/publish.yml/badge.svg)](https://github.com/Parallels/prl-devops-service/actions/workflows/publish.yml)
[![discord](https://dcbadge.vercel.app/api/server/pEwZ254C3d?style=flat&theme=default)](https://discord.gg/pEwZ254C3d)

This is the Parallels Desktop DevOps Service, a service that will allow you to
manage and orchestrate multiple Parallels Desktop hosts and virtual machines.
It will allow you to create, start, stop and delete virtual machines and will
also allow you to manage the hosts that are running the virtual machines.

## Licensing

You can use the service for free with up to **10 users**, without needing any
Parallels Desktop Business license. However, if you want to continue using the
service beyond 10 users, you will need to purchase a Parallels Desktop Business
license. The good news is that no extra license is required for this service
once you have purchased the Parallels Desktop Business license.

## Architecture

The Parallels Desktop DevOps is a service that is written in Go and it is a very
light height designed to provide some of the missing remote management
tools for virtual machines running remotely in Parallels Desktop. It uses rest
api to execute the necessary steps. It also has RBAC (Role Based Access Control)
to allow for a secure way of managing virtual machines. You can manage most of
the operations for a Virtual Machine Lifecycle.
