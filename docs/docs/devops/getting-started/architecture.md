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
* **Packaging:** Multi‑arch Docker images published to **GitHub Container Registry**
* **API:** "Open" REST + websocket communication
* **Database:** SQLite (default) or PostgreSQL via GORM — see
  Data Persistence below (#417)
* **AuthN/Z:** OIDC (JWT) + role‑based access control
* **Deployment:** Helm charts for Kubernetes, Docker Compose for PoC, fully air‑gapped option

## Data Persistence

The service persists its records through a GORM data layer (#417):

* **SQLite (default; supported today):** a single file (`prl-devops.db`)
  in the configured database folder, opened in WAL mode with foreign
  keys enforced.
  <!-- engine default + filename: src/database/config.go
  DefaultDatabaseConfig / DefaultDatabaseFilename; WAL pragmas:
  src/database/sqlite.go -->
* **PostgreSQL (enablement in progress):** selected by DSN; the
  data-access layer still relies on SQLite-specific row ordering, so
  running the service end-to-end on PostgreSQL is not yet supported —
  #417 tracks the gap. Settings and setup reference live on the
  PostgreSQL Setup page.
  <!-- driver wiring: src/database/postgres.go; settings parse:
  src/database/config.go FromApplicationConfig -->
* **Legacy JSON store (`data.json`):** deprecated input for the
  automatic first-start migration; afterwards it is a frozen snapshot.
  See the
  [upgrade guide]({{ site.baseurl }}/docs/devops/getting-started/database-upgrade/).
  <!-- migration + marker table: src/database/legacy_migrator/;
  frozen/selection warnings: src/serviceprovider/data.go; the removal
  version lives in the CHANGELOG only -->
* **Backups:** no built-in SQL backup. SQLite: copy the file
  (`sqlite3 <file> ".backup <dest>"`); PostgreSQL: your operations
  runbook. The built-in backup settings covered the legacy file only.
  <!-- posture gap per OQ-421-3; legacy-only backup code:
  src/data/backup.go -->
