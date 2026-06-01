# Use Case Engine Feature Plan

Status: Design phase awaiting approval by user
Format: Dev-map Feature Issue template (local markdown instead of GitHub)
Date: 2025-12-27

---

## Overview

Build an interactive "use case" engine for our Jekyll docs that lets users follow guided multi-step workflows with branching paths side quests checkpoints and progress persistence --- all driven by YAML definitions.

This mirrors the existing `walkthrough.html` pattern (data-driven JS component in `_includes/`) but is more sophisticated: it supports step kinds beyond hotspot-based UI tours user choices that alter the path side quests (linked use cases) quizzes code editors terminal panels and persistent state via localStorage.

---

## User Story

As a developer reading our docs,
I want interactive guided workflows that adapt to my choices (AWS vs GCP Terraform first OR jump in directly) show checkpoints with real API verification let me take detours into prerequisite learning without losing progress and save my state locally
So that I can complete complex operational procedures efficiently without getting lost or starting over.

---

## Acceptance Criteria

1. A YAML file per use case defines its steps branching logic side quests and panels
2. An intake/introduction card renders from metadata (scenario key outcomes prerequisites architecture diagram)
3. The engine supports these step kinds: narrative choice checkpoint terminal quiz code_editor section_header api_call and integrates the existing walkthrough panel type
4. Branching (choice steps) creates visible decision points; selecting a branch routes to the correct next step
5. Side quests are globally defined linkable from any narrative step auto-save completion state and redirect back to the originating step
6. Progress persists in localStorage and survives browser restart; marked as completed on re-entry
7. Breadcrumbs render from declared category/group metadata
8. All styling follows existing Bulma-based theme variables (`_variables.scss`) and matches screenshot appearance
9. No server-side dependencies -- pure static Jekyll output + client-side JS
10. Unit-test-like validation schema for YAML files (documented spec; runtime warnings)

---

## Definition of Done

- [ ] All components implemented in `_use-case-engine/` (YAML loader renderer localStorage manager SCSS)
- [ ] Example use case YAML produces expected rendered page
- [ ] Integration test: user visits page completes steps refreshes sees saved state
- [ ] JSON Schema validation document exists at `docs/_use-case-engine/schema.json`
- [ ] README at `docs/_use-case-engine/README.md`: "How to author a use case"
- [ ] CSS scoped under `.uce-` prefix to avoid conflicts with bulma/walkthrough styles
- [ ] Review against screenshots before implementation begins

---

## FILE STRUCTURE (Proposed)

```
docs/_use-case-engine/
+-- PLAN.md                                        # This file
+-- README.md                                      # Author guide
+-- schema.json                                    # JSON Schema for validation
+-- layouts/
|   +-- use_case.html                              # New layout for use case pages
+-- includes/
|   +-- uce-intro.html                             # Intro card partial
|   +-- uce-flow.html                              # Main flow renderer partial
|   +-- uce-step-renderer.html                     # Step dispatch partials
+-- assets/
|   +-- css/
|   |   +-- use-case-engine.scss                   # Component styles (SCSS->CSS)
|   +-- js/
|   |   +-- uce-state.js                           # State management
|   |   +-- uce-engine.js                          # State machine + renderer
|   |   +-- uce-panels/
|   |       +-- panel-terminal.js                  # Terminal panel component
|   |       +-- panel-walkthrough.js               # Walkthrough embedding
|   |       +-- panel-api.js                       # API call verification
|   |       +-- panel-quiz.js                      # Quiz assessment
|   |       +-- panel-codeedit.js                  # Code editor
+-- data/
|   +-- use_cases_index.yml                        # Auto-generated index
+-- examples/
    +-- vm-cicd-pipeline.yml                       # Full reference example
```

Where use case YAML files live: `docs/use-cases/` directory (already created).
Each use case is a `.yml` file that Jekyll can render via the custom layout.

---

## FULL YAML SPEC -- Complete Example

```yaml
# === Top-level metadata ===
id: vm-cicd-pipeline                    # unique identifier
title: "CI/CD Pipeline with VMs"        # display title
level: intermediate                     # beginner | intermediate | advanced
duration: 30m                           # human-readable duration
group: Foundations                      # breadcrumb group
category: VMs                           # breadcrumb sub-category
tags: [vm, cicd, terraform, aws, gcp]   # searchable tags
icon: pipeline                          # optional icon for index page

# === Introduction section ===
introduction:
  scenario: >-
    Automating a full CI/CD pipeline using ephemeral VMs for isolated testing.
    This workflow demonstrates how to dynamically spin up compute resources
    execute a test suite and tear everything down automatically.
  key_outcomes:
    - "Zero-manual setup"
    - "Parallel test execution"
    - "Automatic infrastructure cleanup"
  prerequisites:
    - name: "Parallels Desktop"
      check: "parallels-desktop --version"          # hard prereq (CLI check)
    - name: "CI Provider"
      check: null                                   # soft prereq (no CLI check)
  architecture_diagram: ""                          # optional img asset path

# === Global unlocks ===
unlocks: [multi-region-deploys, blue-green]

# === Global side quests pool ===
side_quests:
  terraform-basics:
    id: terraform-basics
    title: "Terraform Basics"
    description: >-
      Learn the fundamentals of Terraform before provisioning infrastructure.
    links_to: learn-terraform                       # links to another use case
    requires: []                                    # additional prereqs
    returns_to: introduce-the-workspace             # where to jump back

# === Steps ===
steps:
  # Section divider
  - kind: section_header
    id: part1-init
    title: "Part 1: Initialization"

  # Narrative step with terminal panel
  - kind: narrative
    id: introduce-the-workspace
    title: "Initialize the workspace"
    duration: 2m
    body: >-
      Start by initializing your project with Orchestra.
      This sets up the working directory and configures default settings.
    panel:
      kind: terminal
      cmd: "orchestra init --project my-app"
      copy_button: true
    side_quests:
      - terraform-basics
    next: provisioning

  # Choice/branching step
  - kind: choice
    id: choose-cloud-provider
    title: "Choose your cloud provider"
    question: "Which cloud provider do you use for VMs?"
    branches:
      aws:
        label: "AWS"
        description: "Use EC2 instances with Application Load Balancer"
        icon: aws
        next: provision-aws-vm
      gcp:
        label: "GCP"
        description: "Use Compute Engine instances"
        icon: gcp
        next: provision-gcp-vm
      azure:
        label: "Azure"
        description: "Use Virtual Machines with Load Balancer"
        icon: azure
        next: provision-azure-vm
    auto_select: false

  # Narrative step with walkthrough panel (reuse existing component)
  - kind: narrative
    id: provision-aws-vm
    title: "Provision the AWS VM"
    duration: 8m
    body: >-
      Create an EC2 instance using our Terraform module.
      The module handles security groups IAM roles and key pair generation.
    panel:
      kind: walkthrough
      main_image: "catalog-main.png"
      items_ref: catalog_walkthrough    # from _data/catalog_walkthrough.yml
      highlight_ids: [pull]
      auto_advance: true

  # Checkpoint step with API verification
  - kind: checkpoint
    id: verify-deployment
    title: "Verify deployment"
    body: "Ensure the VM is running and accessible."
    panel:
      kind: api_call
      url: "{{ site.api_base }}/status"
      method: GET
      headers:
        Authorization: "Bearer {{ api_token }}"
      expect:
        status_code: 200
        timeout_ms: 30000
        response_path: $.status == "healthy"
    on_failure: retry-or-skip

  # Quiz step
  - kind: quiz
    id: knowledge-check
    title: "Quick knowledge check"
    question: >-
      Which port does the VM orchestrator expose by default?
    options:
      - id: a
        text: "Port 8080"
      - id: b         # correct answer
        text: "Port 443"
      - id: c
        text: "Port 80"
      - id: d
        text: "Port 9090"
    correct: b
    feedback_correct: >-
      Correct! Port 443 is exposed for HTTPS.
    feedback_incorrect: >-
      Not quite — it's actually port 443. Try reviewing configuration docs.

  # Code editor step
  - kind: code_editor
    id: write-config
    title: "Write your configuration"
    body: "Create a basic `orchestra.yaml` configuration file:"
    language: yaml
    initial_code: |
      project: my-app
      region: us-east-1
      vm_type: t3.micro
    validation_cmd: "orchestra validate orchestra.yaml"
```

---

## Step Kind Reference Table

| Kind          | Purpose                            | Required fields            | Optional fields               | Panel rendered?          |
|---------------|------------------------------------|----------------------------|-------------------------------|--------------------------|
| narrative     | Show content + optional panel      | title body                 | id duration panel side_quests | Yes (optional)           |
| choice        | Present branch decision            | title question branches    | auto_select                   | No                       |
| checkpoint    | Validate environment/state         | title                      | body panel on_failure         | Yes (required for api)   |
| section_header| Visual divider/group label         | title                      | icon (divider style)          | No                       |
| terminal      | Execute/show CLI command           | cmd                        | copy_button                   | Yes (terminal panel)     |
| quiz          | Knowledge assessment               | question options correct   | feedback_correct incorrect    | No                       |
| code_editor   | Editable code snippet              | initial_code language      | validation_cmd hint title     | Yes (editor panel)       |
| walkthrough   | Reuse existing hotspot tour        | main_image OR items_ref    | highlight_ids auto_advance    | Yes (existing tw comp.)  |
| api_call      | Verify with external endpoint      | (panel.kind = api_call)    | method headers expect url     | Yes                      |

---

## STATE MANAGEMENT

Key format: `uce_progress:<use_case_id>`
Example: `uce_progress:vm-cicd-pipeline`

Value structure:
```json
{
  "current_step": "choose-cloud-provider",
  "completed_steps": ["introduce-the-workspace"],
  "branches_chosen": {
    "choose-cloud-provider": "aws"
  },
  "side_quests_completed": ["terraform-basics"],
  "updated_at": "2025-12-27T10:30:00Z"
}
```

Behavior rules:
1. **On load**: merge saved state into available step list.
   - Completed steps shown as checked (green checkmark)
   - Current step highlighted with border + "CURRENT" badge
   - Unreachable steps (behind unchosen branches) hidden from view
2. **On step complete**: push to completed_steps, advance current_step
3. **On branch choice**: save in branches_chosen[branch_id]
4. **On side quest complete**: add to side_quests_completed[], mark done on parent AND linked pages (cross-use-case persistence)
5. **Auto-save**: every user action writes immediately; debounced at 100ms

Side quest tagging model for user's exact request (save ones taken and ones skipped but allow re-taking):
```json
{
  "side_quest_tags": {
    "terraform-basics": {
      "status": "skipped",      // or "taken"
      "view_count": 0           // increments if user later goes back into it
    }
  }
}
```
When a user skips a side quest it gets tag status:simple"skipped" so the narrative step marks it as covereD But they can always go back and take it later view_count tracks revisit.

---

## RENDER ENGINE FLOW

Page loads (Jekyll renders use_case.html layout)
```
    |
    |---> Reads YAML from frontmatter or _data path
    |---> Injects into uce-intro.html (partial)
    |     Renders: scenario, key_outcomes, prerequisites, progress bar
    |
    +---> Loads uce-engine.js which:
          |---> Parses step definitions from data object
          |---> Resolves localStorage state by use_case_id
          |---> Filters out unreachable steps (based on branches not taken)
          |---> Renders each step sequentially
          |     +---- For each step:
          |           |--- if kind=narrative > show body (markdown-HTML) + optional panel
          |           |--- if kind=choice > render branch cards (hide until selected)
          |           |--- if kind=checkpoint > validate with api_call
          |           |--- if kind=terminal > render code block with copy button
          |           |--- if kind=quiz > render multiple-choice assessment
          |           |--- if kind=code_editor > render editable textarea
          |           |--- if kind=section_header > render divider group label
          |
          +---> Updates localStorage on every interaction
```

Jekyll integration point:
- Use case YAML files live in docs/use-cases/
- Jekyll reads them as page frontmatter or site.data resolution
- Example in layouts/use_case.html:
  ```liquid
  {% assign uc_data = site.data.use_cases[page.slug] %}
  {% include 'uce-intro.html' data=uc_data.introduction %}
  ```

---

## ASCII WIREFRAMES

### Full page layout

```
+==========================================================================+
| Navbar: VM Orchestrator API | Docs | GitHub | Status ...                |
+------------------------------------------------------------------------+-+'
| Breadcrumb: Skill tree > Foundations > VMs > CI/CD pipeline            |
+------------------------------------------------------------------------+-+'
|                                                                        |
|  CI/CD pipeline with VMs                                               |||
|  Intermediate 30 min unlocks 2 more                                    |||
|                                                                        |||
+------------------------------------------------------------------------+-++
| PROGRESS: [--========-- ] 1 of 6                                      |
+-----------------------------------------------------------------------+||
| INTRODUCTION (CURRENT)                                        CURRENT ||
|int |\                                                                    ||
| --------------------------------------------------------------------- ||
| THE SCENARIO                                                           ||
| Automating a full CI/CD pipeline using ephemeral VMs for isolated  |||
| testing. This workflow demonstrates how to dynamically spin up |||   
| compute resources execute a test suite and tear everything down ||
| automatically.                           [ ARCHITECTURE DIAGRAM ] |||
| KEY OUTCOMES                                                          |||
| v Zero-manual setup       +-------------------------------------------|||| 
| v Parallel test exec      |                                           |||
| v Auto infra cleanup      | (img/svg placeholder)                     |||
|                                              ----^----                |||
| PREREQUISITES                       | NEXT ->  |                     |
| o Parallels Desktop                 |__________|                   
| o CI Provider (e.g., GitHub Actions)
| o Basic CLI knowledge
+---------------------------------------------------------------+-+-
| 1. Initialize the workspace              2 min [done-checkmark]
| 2. Provision the VM                       8 min [circle-not-done]
| 3. Define your infrastructure             15 min [circle-not-done]
| 4. Choose your cloud                        -- [circle-not-done]
+---------------------------------------------------------------+|
| CHOOSE YOUR CLOUD                                     BRANCH ||
|int |                                                               ||
| ----------------------------------------------------------------- ||
| Pick one - the next steps adapt to your choice.                  ||
| +--------------------------+  +--------------------------+        \||
| | AWS                      |  | GCP                      |        \||
| | EC2 + ELB path           |  | Compute Engine         |        \||
| | VM Instances             |  | Compute Engines        |        \\||
| | [Select now]             |  | [Select now]           |        \\||
| +--------------------------+  +--------------------------+        \\||
+---------------------------------------------------------------+|   |
|                                                               ||   |
| 4. Define your infrastructure                    15 min       ||   |
| 5. Verify deployment                               3 min      ||   |
|                                                               ||   |
+---------------------------------------------------------------||---+
| <-- Previous                  Progress saved automatically     Next -->
+=======================================================================+
| FOOTER                                                                |
+=======================================================================+
```

### Side quest fly-out (appears below narrative step)

```
+--------------------------------------------------------+
| SIDE QUEST AVAILABLE                                   |
|                                                        |
| Terraform Basics                                       |
| Learn the fundamentals of Terraform before provisioning.|
|                                                        |
| Estimated: 15 minutes                                  |
| Prerequisite for this step                             |
|                                                        |
+----------------------+      +---------------------+    +--+
| Take now!            |      | Skip for now        |    |
+----------------------+      +---------------------+    +--+
```

After taking the side quest and returning:

```
+--------------------------------------------------------+
| DONE: Terraform Basics               [+ View again]    |
| You learned how to provision infrastructure.            |
|                                                        |
| Continue to: Provision the VM --NEXT-->                |
+--------------------------------------------------------+
```

Branch selected state:

```
+--------------------------------------------------------+
| CHOOSE YOUR CLOUD                              SELECTED|
| Pick one - the next steps adapt to your choice.        |
|                                                        |
| X AWS            [SELECTED]   GCP        [not selected]|
|   EC2 + ELB path                              Compute Engine|
|   You chose this                         Not chosen yet  |
+--------------------------------------------------------+
Then shows only the AWS-specific next steps hiding GCP/Azure path
+--------------------------------------------------------+
```

---

## IMPLEMENTATION PHASES

Phase 1 Core Foundation
- YAML data model and JSON Schema definition
- use_case.html layout (extends default adds uce-specific classes)
- uce-state.js: state manager plus localStorage CRUD
- uce-engine.js: state machine plus step router
- Basic narrative step rendering (markdown to HTML via Jekyll markdownify)
- Step sequencing (previous/next navigation, progress bar)
- Intro card partial with scenario/key_outcomes/prerequisites display

Phase 2 Interactive Features
- Choice steps with branch routing UI
- Branch visibility logic (hide unseen branches until selection)
- Checkpoint step with API call verification
- Progress bar with completion tracking
- Side quest trigger from narrative steps

Phase 3 Advanced Panels
- Terminal panel (copy-to-clipboard formatted code blocks)
- Walkthrough panel embedding (reuse walkthrough.html via include)
- Quiz step (multiple choice with scoring/validation)
- Code editor panel (editable textarea with language hint)
- Section header divider rendering

Phase 4 Side Quests
- Global side quest pool definition support
- Side quest modal/slide-out from narrative steps
- Completion tracking and redirect-back logic
- Cross-use-case linking (side quest links to another use case)

Phase 5 Polish and Developer Experience
- Full SCSS component library (.uce-prefixed)
- Responsive breakpoints matching _variables.scss
- Dark mode via CSS custom properties
- Index page generator (all use cases grouped by tag)
- schema.json documentation
- Example YAML file

---

## DEPENDENCIES AND REUSE

| What                                | From                                        |
|-------------------------------------|---------------------------------------------|
| Markdown rendering                  | Jekyll default markdownify filter           |
| Breadcrumbs                         | Existing _breadcrumb.html + page group      |
| Walkthrough embedding               | _includes/walkthrough.html                  |
| Theming/breakpoints                 | _sass/_variables.scss                       |
| Navbar/footer                       | _layouts/default.html                       |
| SCSS compile                        | Existing build pipeline                     |

Style conventions:
- CSS class prefix: uce- for all new classes (no conflicts with bulma or walkthrough)
- JS naming: camelCase, IIFE or ES module wrapped
- SCSS nesting: match patterns in _walkthrough.scss
- Data attributes: data-uce-* for element hooks in JS
- No inline styles: all styling via scoped CSS classes

## OPEN QUESTIONS AND RISKS

1. API calls in checkpoints: Only allow local calls (to services users run locally). Since these are calls to the service we are asking them to run locally, it makes sense to allow these even if the endpoints are not public.
   Recommendation: local-only calls (localhost/127.0.0.1 addresses only)

2. Quiz scoring: Client-side only for now.
   Recommendation: client-side MVP expand later if needed

3. Code editor capabilities: Using a plugin to do real syntax highlighting.
   Recommendation: integrate Prism.js via CDN for real syntax highlighting from the start

4. Performance at scale: Skip for now. We are expecting each use case to be small.
   Low risk at current doc scale skip for MVP

