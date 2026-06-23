# Writing Use Cases

This guide explains how to write use cases for the Parallels DevOps Service documentation.

---

## Table of Contents

1. [File Structure](#file-structure)
2. [Top-Level Properties](#top-level-properties)
3. [Introduction Section](#introduction-section)
4. [Steps](#steps)
5. [Step Kinds](#step-kinds)
6. [Layout Modes](#layout-modes)
7. [Conditional Rendering](#conditional-rendering)
8. [Resources Section](#resources-section)
9. [Checklist](#checklist)
10. [Side Quests](#side-quests)
11. [What's Next](#whats-next)
12. [Walkthrough Panels](#walkthrough-panels)
13. [Expression Resolution](#expression-resolution)
14. [Hidden Flag](#hidden-flag)
15. [Complete Working Example](#complete-working-example)
16. [Tips and Best Practices](#tips-and-best-practices)

---

## File Structure

Each use case consists of two files:

1. **Data file** (`docs/_data/<use-case-id>.yml`) — Contains all structured data: steps, panels, side quests, etc.
2. **Markdown page** (`docs/use-cases/<use-case-id>.md`) — Contains front matter that references the data file.

The data file is the primary source of truth. The page front matter uses `uce_data` to point to the data file key.

---

## Top-Level Properties

The top-level properties define the use case metadata.

```yaml
id: use-case-id
title: Use Case Title
level: beginner | intermediate | advanced
duration: "~5 min"
group: Use Cases
category: DevOps
order: 1
hidden: true
tags:
  - tag1
  - tag2
unlocks:
  - achievement-id
depends_on:
  - prerequisite-use-case
```

---

## Introduction Section

The `introduction` section appears in the intro card at the top of the use case page.

```yaml
introduction:
  scenario: >
    Plain text scenario description.
  markdown_scenario: >
    **Bold** and *italic* supported via markdownify.
  key_outcomes:
    - Outcome 1
    - Outcome 2
  prerequisites:
    hard:
      - Mandatory requirement
    soft:
      - Recommended knowledge
  architecture_diagram: |
    flowchart LR
      A[Start] --> B[End]
  image: /assets/images/my-image.png
```

| Property | Type | Description |
|---|---|---|
| `scenario` | string | Plain text scenario description |
| `markdown_scenario` | string | Markdown scenario (takes precedence over `scenario`) |
| `key_outcomes` | array | Bulleted list of outcomes the user will achieve |
| `prerequisites.hard` | array | Mandatory prerequisites |
| `prerequisites.soft` | array | Recommended but not required |
| `architecture_diagram` | string | Mermaid diagram syntax (renders as flowchart) |
| `image` | string | Image path (takes priority over `architecture_diagram` when both present) |

**Notes:**

- `markdown_scenario` is rendered via Jekyll's `markdownify` filter, so it supports Markdown syntax including bold, italics, links, and blockquotes.
- The `architecture_diagram` uses Mermaid syntax and renders as a flowchart in the intro card's right column.
- The `image` field accepts a path (e.g., `/assets/images/architecture-overview.png`) and renders as a clean framed image with box-shadow and rounded corners.
- When both `image` and `architecture_diagram` are present, `image` takes priority.
- When neither is present, the right column is omitted entirely (single-column layout).

### Example: Full Introduction

```yaml
introduction:
  markdown_scenario: >
    **Parallels DevOps Service** is a command-line tool for managing and
    orchestrating multiple Parallels Desktop hosts and virtual machines. In this
    use case you will install the service on your machine, configure it, start it
    as a background daemon, and verify it is healthy and ready to use.
  key_outcomes:
    - Install the prldevops binary for your platform
    - Create a basic configuration file
    - Run the service as a system daemon
    - Verify the service health via the REST API
    - Understand what tools are needed for advanced features
  prerequisites:
    hard:
      - A Mac, Linux, or Windows machine
      - Access to a terminal with admin/sudo privileges
    soft:
      - Basic familiarity with command-line tools
      - Understanding of REST APIs
  architecture_diagram: |
    flowchart LR
        A[Developer] --> B[Git Push]
        B --> C[CI Runner]
        C --> D[VM/Cloud]
```

---

## Steps

Each use case has a `steps` array. Each step has a `kind` that determines how it renders.

Common step properties:

```yaml
- id: step-id
  kind: narrative
  title: Step Title
  duration: "~3 min"
  tag: API
  layout: columns
  if: choose-platform.value == 'mac'
  body: >
    HTML or plain text body.
  markdown_body: |
    ## Heading
    Markdown content parsed via marked.js.
  resources:
    - title: Resource Name
      url: https://example.com
  checklist:
    - Checklist item 1
    - Checklist item 2
  side_quests:
    - side-quest-id
```

| Property | Type | Description |
|---|---|---|
| `id` | string | Unique step identifier |
| `kind` | string | Step type: `narrative`, `choice`, `quiz`, `checkpoint`, `code_editor`, `markdown`, `section_header` |
| `title` | string | Step title displayed in the step header |
| `duration` | string | Estimated time for this step |
| `tag` | string | Badge shown on certain step types (e.g., `API`, `EXEC`) |
| `layout` | string | Layout mode: `columns`, `rows`, `auto`, `single` |
| `if` | string | Conditional expression for rendering |
| `body` | string | Plain text or HTML body content |
| `markdown_body` | string | Markdown body content (parsed via marked.js) |
| `resources` | array | External resource links |
| `checklist` | array | Checklist items rendered at the bottom of the step |
| `side_quests` | array | Side quest IDs referenced in this step |

### Step Kinds Overview

| Kind | Description | Panels Supported |
|---|---|---|
| `narrative` | Main step type with panels and text | Yes |
| `choice` | Branching user selection | No |
| `quiz` | Knowledge check with multiple choice | No |
| `checkpoint` | API verification with status tracking | No |
| `code_editor` | Monaco code editor | No |
| `markdown` | Raw Markdown step | No |
| `section_header` | Divider with label | No |

---

## Step Kinds

### 1. `narrative` — Main step type with panels

The `narrative` step is the most versatile step type. It supports a left column for body text and a right column for panels. Panels can be singular (one panel) or plural (multiple tabbed panels).

**Single panel:**

```yaml
- id: install-step
  kind: narrative
  title: Install the CLI
  layout: rows
  markdown_body: |
    Follow the instructions for your platform.
  panel:
    kind: terminal
    name: Install
    cmd: |
      curl -fsSL https://example.com/install.sh | bash
    output: |
      Installing... done!
    copy_button: true
  resources:
    - title: Docs
      url: /docs/install/
```

**Multiple panels (tabbed):**

```yaml
- id: multi-panel-step
  kind: narrative
  title: Platform-Specific Setup
  layout: rows
  panels:
    - kind: terminal
      id: install-mac
      name: macOS
      if: choose-platform.value == 'mac'
      markdown_body: |
        Instructions for Mac.
      cmd: |
        brew install tool
      copy_button: true
    - kind: terminal
      id: install-linux
      name: Linux
      if: choose-platform.value == 'linux'
      markdown_body: |
        Instructions for Linux.
      cmd: |
        sudo apt install tool
      copy_button: true
```

**Important rules:**

- Panel-level `markdown_body` takes precedence over step-level `markdown_body`.
- Panel-level `side_quests` overrides step-level `side_quests`.
- When `panels` (plural) has more than one visible panel, a tab bar appears above the panel content.
- When only one panel is visible, the tab bar is hidden and the panel renders directly.

### 2. `choice` — Branching user selection

The `choice` step presents cards that the user can select. The selected value is stored and can be referenced in conditions and expressions throughout the use case.

```yaml
- id: choose-platform
  kind: choice
  title: Choose Your Platform
  question: Which platform are you installing on?
  icon: fa-desktop
  branches:
    mac:
      label: macOS
      value: mac
      description: Apple Silicon or Intel Mac with full VM support
      icon: fa-apple
      icon_type: fa-brands
    linux:
      label: Linux
      value: linux
      description: Orchestrator and Catalog features only
      icon: fa-linux
      icon_type: fa-brands
    windows:
      label: Windows
      value: windows
      description: Orchestrator and Catalog features only
      icon: fa-windows
      icon_type: fa-brands
    source:
      label: From Source
      value: source
      description: Build from the GitHub repository
      icon: fa-github
      icon_type: fa-brands
```

Branch values can be referenced in conditions and expressions:

- `${{ choose-platform.value }}` — The selected branch value (e.g., `mac`, `linux`)
- `${{ choose-platform.label }}` — The selected branch label (e.g., `macOS`, `Linux`)

### 3. `quiz` — Knowledge check

The `quiz` step presents a multiple-choice question with immediate feedback.

```yaml
- id: quiz-features
  kind: quiz
  title: Quick Knowledge Check
  duration: "~1 min"
  question: |
    Which features are available when running the service on Linux or Windows?
  options:
    - text: "Full VM management, Orchestrator, and Catalog"
      correct: false
    - text: "Orchestrator and Catalog only"
      correct: true
    - text: "Catalog only"
      correct: false
    - text: "None -- the service only runs on macOS"
      correct: false
  feedback_correct: |
    Correct! Since virtual machine management requires Parallels Desktop for Mac,
    running the service on Linux or Windows limits it to the Orchestrator and Catalog features.
  feedback_incorrect: |
    Not quite. Virtual machine management requires Parallels Desktop for Mac.
    Therefore, Linux and Windows installations only support the Orchestrator and Catalog features.
```

**Properties:**

| Property | Type | Required | Description |
|---|---|---|---|
| `question` | string | Yes | The question text (supports Markdown) |
| `options` | array | Yes | Array of answer options |
| `options[].text` | string | Yes | Answer text |
| `options[].correct` | boolean | Yes | Whether this is the correct answer |
| `feedback_correct` | string | No | Message shown when the correct answer is selected |
| `feedback_incorrect` | string | No | Message shown when an incorrect answer is selected |

### 4. `checkpoint` — API verification

The `checkpoint` step lets users verify their environment by making API calls. It supports both single-task and multi-task modes.

**Single-task mode:**

```yaml
- id: verify-env
  kind: checkpoint
  title: Verify Your Environment
  duration: "~3 min"
  body: |
    Before proceeding, verify that Docker and Git are installed correctly.
    Run the verification below to confirm your environment is ready.
  checkpoint:
    api:
      url: "http://localhost:80/api/health/probe"
      method: GET
      timeout_ms: 10000
      expect_status_code: 200
      response_path: '$.status == "OK"'
      response_type: json
    on_failure: retry-or-skip
```

**Multi-task mode (tasks array):**

```yaml
- id: multi-check
  kind: checkpoint
  title: Verify Everything
  duration: "~5 min"
  body: |
    Multiple verification tasks to confirm your setup.
  checkpoint:
    api:
      tasks:
        - id: api-health
          title: Verify API Connectivity
          url: "http://localhost:5475/api/health/probe"
          method: GET
          timeout_ms: 10000
          expect_status_code: 200
          response_path: '$.status == "OK"'
          response_type: json
        - id: vm-state
          title: Check VM State
          url: "http://localhost:5475/api/vm/status"
          method: GET
          timeout_ms: 10000
          response_path: '$.state == "running"'
          response_type: json
    on_failure: retry-or-skip
```

**Properties:**

| Property | Type | Description |
|---|---|---|
| `checkpoint.api.url` | string | API endpoint URL |
| `checkpoint.api.method` | string | HTTP method (GET, POST, PUT, DELETE) |
| `checkpoint.api.timeout_ms` | integer | Timeout in milliseconds (default: 30000) |
| `checkpoint.api.expect_status_code` | integer | Expected HTTP status code |
| `checkpoint.api.response_path` | string | JSONPath expression to validate against response |
| `checkpoint.api.response_type` | string | Response type: `json` or `text` |
| `checkpoint.api.tasks` | array | Multi-task mode: array of individual API tasks |
| `checkpoint.on_failure` | string | Behavior on failure: `retry-only` or `retry-or-skip` |

### 5. `code_editor` — Monaco code editor

The `code_editor` step provides an in-browser code editor using Monaco Editor (the engine behind VS Code).

**Single file:**

```yaml
- id: write-config
  kind: code_editor
  title: Write Your Configuration
  duration: "~3 min"
  markdown_body: |
    Copy the configuration below into your config.yaml file.
  files:
    - name: config.yaml
      language: yaml
      initial_code: |
        environment:
          api_port: 80
          log_level: DEBUG
          ROOT_PASSWORD: VeryStr0ngPassw0rd
```

**Multiple files (tabbed editor):**

```yaml
- id: multi-file
  kind: code_editor
  title: Project Files
  files:
    - name: Dockerfile
      language: dockerfile
      initial_code: |
        FROM node:18
        COPY . .
        RUN npm install
    - name: package.json
      language: json
      initial_code: |
        {
          "name": "myapp",
          "version": "1.0.0"
        }
```

**Properties:**

| Property | Type | Description |
|---|---|---|
| `language` | string | Default language (can be overridden per file) |
| `files` | array | Array of files to edit (tabbed if more than one) |
| `files[].name` | string | Display name for the file tab |
| `files[].language` | string | Syntax highlighting language |
| `files[].initial_code` | string | Initial code content |
| `validation_cmd` | string | Optional command to validate the code |

Supported languages: `yaml`, `json`, `javascript`, `typescript`, `python`, `bash`, `shell`, `dockerfile`, `html`, `css`, `markdown`, and more.

### 6. `markdown` — Raw Markdown step

The `markdown` step renders raw Markdown content via marked.js.

```yaml
- id: intro-markdown
  kind: markdown
  title: Introduction
  markdown_body: |
    ## Welcome
    This is **bold** and *italic* content.
    
    > A blockquote.
    
    - Item 1
    - Item 2
```

### 7. `section_header` — Divider with label

The `section_header` step renders a horizontal divider with a section label.

```yaml
- id: section-start
  kind: section_header
  title: Installation
  icon: fa-download
```

---

## Layout Modes

The `layout` property on a step controls how the body text and panels are arranged.

| Value | Description |
|---|---|
| `columns` (default) | Left column for body text, right column for panels |
| `rows` | Stacked layout — full-width panels below body text |
| `auto` | Resolves to `columns` if panels exist, `rows` otherwise |
| `single` | Forces single column (used when no panel is present) |

**Example: Rows layout**

```yaml
- id: health-check
  kind: narrative
  title: Verify the Service is Running
  duration: "~2 min"
  tag: API
  layout: rows
  markdown_body: |
    Now that the service is running, let's test the connection.
  panel:
    kind: api_call
    title: Health Probe
    url: "http://localhost:80/api/health/probe"
    method: GET
    expect_status: 200
    copy_button: true
  checklist:
    - The request uses the standard GET method.
    - The service returns a 200 OK status response.
    - The response includes a JSON body with a status field.
```

---

## Conditional Rendering

Steps and panels can be conditionally rendered using the `if` property. The condition references values from `choice` steps.

```yaml
- id: install-binary
  kind: narrative
  title: Install the Binary
  panels:
    - kind: terminal
      name: Install
      if: choose-platform.value == 'mac' || choose-platform.value == 'linux'
      cmd: |
        curl -fsSL https://example.com/install.sh | bash
    - kind: terminal
      name: Install Windows
      if: choose-platform.value == 'windows'
      cmd: |
        wget https://example.com/install.exe
```

Conditions support:

- Equality: `choose-platform.value == 'mac'`
- Inequality: `choose-platform.value != 'mac'`
- Logical OR: `choose-platform.value == 'mac' || choose-platform.value == 'source'`
- Logical AND: `choose-platform.value == 'mac' && choose-platform.value != 'source'`

---

## Resources Section

The `resources` array appears at the bottom of steps as a collapsible section with external links.

```yaml
resources:
  - title: API Reference
    url: /docs/api/
  - title: Download Page
    url: https://github.com/releases
```

**Rules:**

- Both `title` and `url` are required for each resource entry.
- Entries without a title or URL are silently ignored.
- Resources are rendered as a collapsible flyout panel with a "Go" button for each link.
- Resources are optional at the step level.

---

## Checklist

The `checklist` array renders as a bulleted list at the bottom of the step body, spanning the full width of both columns in a two-column layout.

```yaml
checklist:
  - The request uses the standard GET method.
  - The service returns a 200 OK status response.
  - The response includes a JSON body containing a status field.
```

**Rendering:**

- Checklist items are rendered with green check icons.
- Items support Markdown syntax (bold, italic, inline code).
- The checklist spans the full width of the narrative grid (both columns).
- Checklist is optional and can appear on any step kind.

---

## Side Quests

Side quests are optional sub-tasks that users can explore at any point. They appear as inline expandable cards within the step's left column.

**At the step level:**

```yaml
- id: infra-intro
  kind: narrative
  title: Setting Up Infrastructure
  side_quests:
    - terraform-basics
```

**At the panel level (overrides step-level):**

```yaml
panels:
  - kind: terminal
    name: Install
    side_quests:
      - docker-networking
```

**Top-level side quest definitions:**

```yaml
side_quests:
  - id: docker-start
    title: Run the Web Dashboard with Docker
    expanded: true
    description: |
      The DevOps Service includes a web-based dashboard for visual management.
      This side quest walks you through running the dashboard as a Docker container.
    links_to: devops-ui-docker-start
    requires: []
    duration: "~5 min"
  - id: terraform-basics
    title: Terraform Basics
    expanded: true
    description: |
      Learn the fundamentals of Infrastructure as Code using Terraform.
    links_to: vm-cicd-pipeline
    requires: []
    duration: "~15 min"
```

**Side quest properties:**

| Property | Type | Description |
|---|---|---|
| `id` | string | Unique identifier, referenced in `side_quests` arrays |
| `title` | string | Display title in the side quest card |
| `expanded` | boolean | If true, opens automatically when rendered |
| `description` | string | Description text shown in the card body |
| `links_to` | string | Slug of the linked use case (hidden use cases are filtered out) |
| `requires` | array | Prerequisite side quest IDs that must be completed first |
| `duration` | string | Estimated time for the side quest |

**Filtering rules:**

- Side quests linked to hidden use cases are automatically filtered out.
- The filter checks if `links_to` contains `://` (external URL). If it does, the side quest is always shown.
- If `links_to` is a use case slug and that use case has `hidden: true`, the side quest is hidden.

---

## What's Next

The `whats_next` section appears at the bottom of the use case after completion. It shows recommended next use cases.

```yaml
whats_next:
  title: Journey Complete
  body: >
    You've successfully installed, configured, and verified the Parallels DevOps Service.
    Here's how to take your fleet management to the next level.
  recommendations:
    - use_case: devops-ui-docker-start
      icon: fa-box-archive
    - use_case: orchestrator-use-case
      title: Fleet Orchestration
      description: Pool many hosts into one schedulable fleet
      icon: fa-network-wired
```

**Properties:**

| Property | Type | Description |
|---|---|---|
| `title` | string | Title shown in the What's Next section |
| `body` | string | Descriptive text (supports Markdown) |
| `recommendations` | array | List of recommended use cases |
| `recommendations[].use_case` | string | Slug of the recommended use case |
| `recommendations[].title` | string | Override title for the recommendation |
| `recommendations[].description` | string | Custom description for the recommendation |
| `recommendations[].icon` | string | Font Awesome icon class |

**Filtering rules:**

- Hidden use cases are automatically filtered from recommendations.
- Recommendations reference use case IDs that correspond to other data files.

---

## Walkthrough Panels

Walkthrough panels render interactive hotspot-based image tours. They reference a separate data file in `docs/_data/`.

**Step-level panel definition:**

```yaml
- id: catalog-walkthrough
  kind: narrative
  panel:
    kind: walkthrough
    items_ref: catalog_walkthrough
    highlight_ids:
      - search
      - manifest
    auto_advance: false
```

**Referenced walkthrough data file structure:**

```yaml
main_image: catalog-main.png
title: Catalog Interface
show_background_effect: false
items:
  - id: header
    type: step
    title: Catalog Service Information
    description: Check the catalog information such as name, description, and number of manifests available.
    hotspot:
      x: 35
      y: 12.5
      radius: 25
  - id: search
    type: step
    title: Search and Filter
    description: Find any VM in your catalog using powerful search and category filters.
    hotspot:
      x: 75
      y: 12.5
      radius: 30
  - id: nested-scene
    type: scene
    title: Manifest Details
    description: View full manifest information including version history, taints, and security claims.
    image: catalog-main-side-menu.png
    steps:
      - id: rbac-overview
        type: step
        title: RBAC Layer
        description: Every catalog enforces role-based access control.
        hotspot:
          x: 40
          y: 50
          radius: 30
          color: "#3B82F6"
```

**Walkthrough data properties:**

| Property | Type | Description |
|---|---|---|
| `main_image` | string | Primary image filename (stored in `docs/assets/img/ui/`) |
| `title` | string | Walkthrough title shown in the info panel |
| `show_background_effect` | boolean | Whether to show a background blur effect |
| `items` | array | Array of hotspot items |
| `items[].id` | string | Unique identifier for the item |
| `items[].type` | string | `step` for standard hotspot, `scene` for nested image |
| `items[].title` | string | Title shown in the info panel |
| `items[].description` | string | Description text (supports HTML) |
| `items[].hotspot.x` | float | Horizontal position as percentage (0-100) |
| `items[].hotspot.y` | float | Vertical position as percentage (0-100) |
| `items[].hotspot.radius` | integer | Hotspot circle radius in pixels |
| `items[].hotspot.color` | string | Optional accent color for sub-scenes |
| `items[].image` | string | Nested image filename (for `scene` type items) |
| `items[].steps` | array | Sub-steps for `scene` type items |

**Walkthrough panel properties:**

| Property | Type | Description |
|---|---|---|
| `kind` | string | Must be `walkthrough` |
| `items_ref` | string | Key referencing a walkthrough data file |
| `highlight_ids` | array | IDs of items to highlight on load |
| `auto_advance` | boolean | Whether to auto-advance between steps |

---

## Expression Resolution

Use the `${{ variable }}` syntax throughout body text, panel `markdown_body`, and step titles to reference choice selections dynamically.

**Available variables:**

- `${{ choose-platform.value }}` — The selected branch value (e.g., `mac`, `linux`)
- `${{ choose-platform.label }}` — The selected branch label (e.g., `macOS`, `Linux`)

**Usage examples:**

```yaml
title: "Install the prldevops Binary on ${{ choose-platform.label }}"
markdown_body: |
  On ${{ choose-platform.value }}, run the following command.
panels:
  - kind: terminal
    name: Install
    cmd: |
      curl -fsSL https://example.com/install.sh | bash
    if: choose-platform.value == 'mac'
```

Expressions are resolved server-side by Jekyll and client-side by the expression resolver in the engine.

---

## Hidden Flag

Use cases can be hidden from the index page by setting `hidden: true` at the top level of the YAML data file:

```yaml
id: vm-cicd-pipeline
title: Building a CI/CD Pipeline on a Virtual Machine
hidden: true
```

Hidden use cases are:

- Excluded from the use case index page listing.
- Filtered from "What's Next" recommendations.
- Cause any side quests linked to them to be hidden as well.

The hidden flag can also be set in the front matter of the corresponding markdown page, but setting it in the data file is the recommended approach.

---

## Complete Working Example

Here is a minimal but complete use case with all major features:

```yaml
id: my-first-use-case
title: My First Use Case
level: beginner
duration: "~10 min"
group: Use Cases
category: Getting Started
order: 1
tags:
  - getting-started
  - tutorial

scenario: >
  Learn the basics of using the Parallels DevOps Service.

introduction:
  markdown_scenario: >
    **Getting Started** with Parallels DevOps Service is simple.
  key_outcomes:
    - Install the CLI
    - Run your first command
  prerequisites:
    hard:
      - A terminal window
      - Internet connection

steps:
  - id: welcome
    kind: markdown
    title: Welcome
    markdown_body: |
      ## Hello!
      
      This tutorial will walk you through the basics.

  - id: choose-type
    kind: choice
    title: What do you want to do?
    branches:
      install:
        label: Install CLI
        value: install
        description: Get the CLI on your machine
      config:
        label: Configure
        value: config
        description: Set up your configuration

  - id: install-step
    kind: narrative
    title: Install the CLI
    layout: rows
    markdown_body: |
      Run the command for your platform.
    panels:
      - kind: terminal
        name: Install
        cmd: |
          curl -fsSL https://example.com/install.sh | bash
        output: |
          Installed successfully!
        copy_button: true

  - id: verify
    kind: checkpoint
    title: Verify Installation
    body: |
      Check that the CLI is working.
    checkpoint:
      api:
        url: "http://localhost:80/api/health/probe"
        method: GET
        timeout_ms: 10000
        expect_status_code: 200
        response_path: '$.status == "OK"'
        response_type: json
    checklist:
      - The CLI responds to health checks
      - The status is OK

side_quests:
  - id: explore-more
    title: Explore More Features
    description: |
      Discover advanced features and configurations.
    links_to: advanced-features
    requires: []
    duration: "~15 min"

whats_next:
  title: Great Job!
  body: >
    You've completed the basics. Continue learning with these use cases.
  recommendations:
    - use_case: advanced-tutorial
      icon: fa-arrow-right
```

---

## Tips and Best Practices

1. **Keep steps small and focused.** Each step should represent a single, atomic action. Long steps overwhelm users.

2. **Use `markdown_body` for rich content.** The `markdown_body` field supports full Markdown syntax and is rendered via marked.js. Use it for headings, lists, blockquotes, and inline code.

3. **Prefer `panel` over `panels` when possible.** A single panel renders more simply (no tab bar). Use `panels` only when you need platform-specific branches.

4. **Use `layout: rows` for terminal-heavy steps.** When the step is primarily about running commands, the rows layout gives panels full width.

5. **Add `checklist` items for verification steps.** Checklists help users track what they've accomplished and serve as a mental model for the step's purpose.

6. **Use `side_quests` for optional deep-dives.** Side quests let curious users explore related topics without blocking the main flow.

7. **Test conditions thoroughly.** Use `if` conditions to hide irrelevant steps for each platform. Test each branch path individually.

8. **Keep side quest descriptions concise.** The description is shown inline in the step — aim for 2-3 sentences.

9. **Use meaningful step IDs.** Step IDs should be descriptive and consistent (e.g., `install-binary`, `verify-health`, `write-config`).

10. **Document your walkthrough hotspot coordinates.** Keep a spreadsheet of hotspot x/y positions so you can reference them when creating or updating walkthrough images.

11. **Use `duration` estimates that are realistic.** Users rely on duration estimates to plan their time. Be conservative.

12. **Add `resources` for external references.** Link to documentation, API references, and other helpful materials at the bottom of relevant steps.



