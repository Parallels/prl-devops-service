# Use Case Engine

An interactive, guided workflow engine for Jekyll documentation.

## What is the Use Case Engine?

The Use Case Engine transforms static documentation into **interactive guided experiences**. Users progress through numbered steps, make decisions that branch the experience, complete quizzes, edit code, and verify their environment — all within a single page.

Key features:
- **Branching workflows** — Users choose paths (e.g., AWS vs GCP vs Azure)
- **Interactive checkpoints** — Verify environment state via API calls
- **Quizzes** — Knowledge checks with instant feedback
- **Code editors** — Editable code snippets with auto-save
- **Terminal panels** — Display commands with copy-to-clipboard
- **Side quests** — Optional learning modules that link between use cases
- **Progress persistence** — State saved in localStorage, survives page reloads
- **Dark mode** — Automatic (via `prefers-color-scheme`) or manual (`data-theme="dark"`)

## Quick Start

### 1. Create a Data File

Place the use case YAML data in `docs/_data/`:

```yaml
# docs/_data/my-first-use-case.yml
id: my-first-use-case
title: My First Use Case
level: beginner
duration: "~5 min"
steps:
  - id: intro
    kind: narrative
    title: Welcome
    body: "Hello, world!"
  - id: end
    kind: terminal
    title: Complete
    body: "You finished!"
```

### 2. Create a Page File

Create a Markdown page in `docs/use-cases/` that references the data file:

```yaml
---
layout: use_case
title: My First Use Case
level: beginner
duration: "~5 min"
group: Use Cases
category: DevOps
tags:
  - getting-started
scenario: A brief description for the index card.
unlocks:
  - starter-badge
uce_data: my-first-use-case   # <-- key matching the _data filename
---
```

The `uce_data` field tells the layout which file in `_data/` to load.

### 2. Navigate to It

Visit `/docs/use-cases/my-first-use-case` to see the rendered experience.

## YAML Reference

### Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier |
| `title` | string | Yes | Display title |
| `level` | string | No | `beginner`, `intermediate`, `advanced` |
| `duration` | string | No | Estimated time |
| `scenario` | string | No | Experience description |
| `key_outcomes` | array | No | Learning outcomes |
| `prerequisites` | array | No | Requirements (`hard`/`soft`) |
| `tags` | array | No | Tags for filtering |
| `steps` | array | Yes | Step definitions |
| `side_quests` | array | No | Global side quest pool |

### Step Kinds

| Kind | Description | Required Fields |
|------|-------------|-----------------|
| `narrative` | Text/content step | `body` |
| `section_header` | Section divider | `title`, `icon` (optional) |
| `choice` | Branching decision | `question`, `branches` |
| `checkpoint` | Environment verification | `checkpoint.api.url` |
| `quiz` | Multiple-choice question | `question`, `options` |
| `code_editor` | Editable code snippet | `language`, `initial_code` |

See the [example YAML](../_use-case-engine/examples/vm-cicd-pipeline.yml) for a complete demonstration of all step kinds.

## Side Quests

Side quests are optional learning modules linked from narrative steps. They can link to other use cases, and users are automatically redirected back upon completion.

```yaml
# At the top level:
side_quests:
  - id: terraform-basics
    title: Terraform Basics
    description: Learn IaC fundamentals.
    links_to: learn-terraform
    duration: ~15 min

# From a narrative step:
- id: learn-more
  kind: narrative
  side_quests:
    - terraform-basics
```

## Branching

Define decision points with `choice` steps:

```yaml
- id: choose-cloud
  kind: choice
  question: Pick your cloud provider
  branches:
    aws:
      label: AWS
      description: EC2 path
    gcp:
      label: Google Cloud
      description: Compute Engine path
```

The engine automatically filters steps based on the chosen branch and persists the selection in localStorage.

## State Persistence

User progress is saved in `localStorage` under the key `uce_progress:<use_case_id>`. This includes:
- Completed steps
- Branch choices
- Quiz results
- Side quest completion/skip state

State persists across page reloads and browser sessions.

## Schema Validation

A JSON Schema is provided for validating use case YAML files. Use it to check your files:

```bash
npx ajv-cli validate \
  -s docs/_use-case-engine/schema.json \
  -d docs/use-cases/my-use-case.yml
```

## Troubleshooting

### Steps not rendering
- Check that each step has a unique `id`
- Ensure `kind` is one of: `narrative`, `section_header`, `choice`, `checkpoint`, `quiz`, `code_editor`, `terminal`, `walkthrough`
- Check browser console for errors

### Branches not working
- Ensure each branch has a unique key (e.g., `aws`, `gcp`)
- The `next` field in branches should reference a valid step `id`

### Checkpoint always fails
- Verify the API URL is accessible from the browser
- Check that `response_path` matches the actual JSON structure
- Remember: only `localhost` and `127.0.0.1` URLs are allowed for security

### Progress resets on reload
- Check that localStorage is not blocked by browser settings
- Ensure no other script clears `uce_progress:*` keys
