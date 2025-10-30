---
layout: page
title: Architecture
subtitle: Source Code Details
menubar: docs_devops_menu
show_sidebar: false
toc: false
---

# Source Code

The source code for the DevOps service is available on [GitHub](https://github.com/Parallels/prl-devops-service){:target="_blank"}.

## 2 Folder Structure

The main source code resides in the [src](https://github.com/Parallels/prl-devops-service/tree/main/src) directory. Here’s a breakdown of the primary folders and their roles:

Folder | Role in the code-base | Why it matters
api_documentation/ | Generates and hosts the OpenAPI / Swagger spec for every REST and gRPC endpoint. It contains a small code-gen harness plus a packaged HTML bundle for publishing the docs. GitHub | Keeps API contracts explicit and version-controlled for both internal teams and external integrators.
basecontext/ | Shared helpers that enrich Go context.Context with common cross-cutting data (request IDs, auth claims, deadlines). GitHub | Guarantees traceability and consistent cancellation semantics across micro-services.
catalog/ | All logic for image metadata: push, pull, validate, taint/revoke, and a pluggable cache layer. Sub-packages such as cacheservice/ and providers/ implement storage adapters. GitHub | Powers the golden-image workflow so every VM starts from a trusted, versioned manifest.
cmd/ | Entry points for each executable (CLI, orchestrator, catalog, reverse-proxy). Builds are wired through Go’s main packages here. GitHub | Cleanly separates binaries from libraries, simplifying Docker image builds.
common/ | A shared package featuring a global instance of the common-go-logger, with timestamps enabled for enhanced logging accuracy.| Improves logging consistency and debuggability across the service.
compressor/ | Utilities to (de)compress large VM archives before upload or after download. GitHub | Reduces network egress costs and speeds up catalog import/export.
config/ | Strongly-typed structs + YAML/ENV loaders that capture every tunable system parameter. GitHub | Central location for default values, env overrides, and config validation.
constants/ | Service-wide enumerations, default strings, HTTP header keys, etc. GitHub | Avoids magic strings scattered across the code-base.
controllers/ | Thin HTTP / gRPC handler layer that wires routing to the underlying service logic. GitHub | Implements the “C” of MVC, keeping transport concerns out of business code.
data/ | Embedded assets and static seed files (e.g., SQL migrations, default manifests). GitHub | Ensures the service is self-contained at start-up—no external bootstrap scripts needed.
docs/ | Developer-facing design notes and ADRs that accompany the code. GitHub | Provides living documentation right beside the implementation.
errors/ | Centralised error types, gRPC status mapping, and helper wrappers for stack traces. GitHub | Gives callers rich, typed feedback while maintaining uniform logging.
helpers/ | A utility package providing reusable functions for date/time handling, HTTP helpers, integer operations, OS utilities, string manipulation, version parsing, and platform-specific volume management. | Centralizes common helper logic, reducing duplication and simplifying maintenance across the codebase.
install/ | Shell scripts and helper manifests for bootstrapping hosts and registering them with the orchestrator. GitHub | Accelerates first-time setup in lab and CI environments.
logs/ | Custom logger façade that standardises JSON output and hooks into telemetry back-ends. GitHub | Provides consistent, structured logs across every micro-service.
mappers/ | DTO ↔️ persistence translators, shielding business structs from DB schemas. GitHub | Decouples storage evolution from API contracts.
models/ | Core domain structs (VM, Host, Manifest, Token, etc.). Acts as the canonical data model. GitHub | Shared language across services, DB, and API.
notifications/ | Abstraction for outbound webhooks / event streams (Slack, email, Opsgenie). GitHub | Enables alerting and CI feedback loops without polluting core logic.
orchestrator/ | Heart of the platform: host discovery, VM scheduling, health checks, and lifecycle APIs. Individual Go files map 1-to-1 with REST resources (e.g., get_host_virtual_machines.go). GitHub | Ensures the right VM lands on the right macOS host with minimal latency.
pdfile/ | Helpers for parsing the Parallels .pvm / .macvm package format (disk images, config). GitHub | Makes VM introspection and validation platform-agnostic.
restapi/ | OpenAPI-generated client/server glue code plus custom middleware (CORS, auth, metrics). GitHub | Keeps transport boilerplate out of business services.
reverse_proxy/ | Lightweight TCP/HTTP router that exposes VM services (SSH, RDP, HTTPS) on per-VM sub-domains/ports. Includes a small models/ package for runtime config. GitHub | Gives external users a stable URL while hosts and VMs remain private.
security/ | Token validation, password hashing, and RBAC policy helpers. GitHub | Centralises all authZ/authN concerns in one place.
serviceprovider/ | Go interfaces + adapters for external systems (storage, queue, secrets). GitHub | Allows swapping S3 for Azure Blob, or adding a new secret vault, without touching business code.
sql/ | SQL scripts and migration stubs, often embedded by data/. GitHub | Version-controls schema alongside code for repeatable deployments.
startup/ | Wiring for dependency injection: builds service graphs, sets up routing, starts HTTP/gRPC servers. GitHub | Keeps main.go thin and makes unit-testing easier.
telemetry/ | Prometheus metrics, OpenTelemetry traces, and health-check endpoints. GitHub | Gives ops teams deep insight into performance and error rates.
tests/ | Integration and behaviour-driven tests that spin up in-memory services or Docker-compose stacks. GitHub | Protects against regressions across the multi-service surface area.
writers/ | Stream/JSON/XML writers for large HTTP responses (log tails, VM exports). GitHub | Keeps memory usage predictable when returning big payloads.
