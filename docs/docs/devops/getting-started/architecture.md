---
layout: page
title: Getting Started
subtitle: Service Architecture
menubar: docs_devops_menu
show_sidebar: false
toc: false
---

# Service Architecture

Modernize your macOS & Windows CI pipelines with orchestrated, on‑demand VMs powered by Parallels Desktop.

# Parallels DevOps Service

> **Modernize your macOS & Windows CI pipelines with orchestrated, on‑demand VMs powered by Parallels Desktop.**

---

## 1. What Is It?

The **Parallels DevOps Service** is a lightweight, cloud‑native platform that
provisions, scales, and retires virtual machines (VMs) just‑in‑time for your
build and test pipelines. Think of it as a **“Kubernetes for Parallels Desktop”**
—but tuned for macOS, Windows, and mixed‑architecture workloads.

## Purpose

The Parallels DevOps Service exists to **accelerate software delivery** by
providing development and QA teams with an automated, secure, and predictable
way to spin up disposable Parallels Desktop® virtual machines on demand. By
abstracting away the manual steps of image preparation, host selection, and
cleanup, the platform shortens feedback loops, reduces infrastructure toil, and
aligns macOS and Windows workflows with modern DevOps practices.

## 2. High‑Level Architecture

| Layer | Role | Key Tech |
|-------|------|----------|
| **Orchestrator Service** | Schedules VMs across a pool of hosts | Go 1.22, gRPC/REST, Docker/K8s |
| **Catalog Service** | Stores **metadata** (manifest, versions, RBAC) about golden images | Go 1.22, gRPC/REST |
| **Storage Providers** | Hold **VM binaries** (pvm/macvm) | Amazon S3, Azure Blob, Artifactory, … |
| **Reverse Proxy** | Routes requests to the appropriate virtual machine  inside the host | Go 1.22, Rest API |
| **Host Agents** | Run on macOS machines with Parallels Desktop to report VM status and execute commands | Go 1.22, Rest API, Parallels Desktop Cli |

---

## 3. Technology Stack

* **Language:** Go (services & CLI)
* **Packaging:** Multi‑arch Docker images published to **Github** and **Docker Hub**
* **API:** "Open" REST + websocket communication
* **Database:** Json file (pluggable—MySQL support in roadmap)
* **AuthN/Z:** OIDC (JWT) + role‑based access control
* **Deployment:** Helm charts for Kubernetes, Docker Compose for PoC, fully air‑gapped option
