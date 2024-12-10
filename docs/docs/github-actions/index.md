---
layout: page
title: Parallels Desktop DevOps GitHub Action
subtitle: How to use the Visual Studio Code Extension
menubar: docs_github_action_menu
show_sidebar: false
is_index: true
badges:
  - type: coverage
    badge: '![coverage](https://raw.githubusercontent.com/Parallels/parallels-desktop-github-action/main/badges/coverage.svg)'
  - type: linter
    badge: '[![Lint Codebase](https://github.com/Parallels/parallels-desktop-github-action/actions/workflows/linter.yml/badge.svg)](https://github.com/Parallels/parallels-desktop-github-action/actions/workflows/linter.yml)'
  - type: CI
    badge: '[![CI](https://github.com/Parallels/parallels-desktop-github-action/actions/workflows/ci.yml/badge.svg)](https://github.com/Parallels/parallels-desktop-github-action/actions/workflows/ci.yml)'
version: 0.3.3
repo: parallels/parallels-desktop-github-action
category: Documentation
---

This action allows you to run Parallels Desktop virtual machines in your GitHub
Actions workflows. You can start, stop, and run commands in a VM, as well as
clone, create, and delete VMs.

## Usage

```yaml
name: Run Parallels Desktop VM
on: [push]

jobs:
  parallels-desktop:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Pull From Catalog
        id: pull
        uses: parallels/parallels-desktop-github-action@v1
        with:
          operation: 'pull'
          username: ${{ secrets.PARALLELS_USERNAME }}
          password: ${{ secrets.PARALLELS_PASSWORD }}
          host_url: devops.example.com
          base_image:
            root:${{
            secrets.CATALOG_ROOT_PASSWORD}}@catalog.example.com/mac-github-runner/v1
      - name: Configure Github Runner
        uses: parallels/parallels-desktop-github-action@v1
        with:
          operation: 'run'
          username: ${{ secrets.PARALLELS_USERNAME }}
          password: ${{ secrets.PARALLELS_PASSWORD }}
          host_url: devops.example.com
          machine_name: ${{ steps.pull.outputs.machine_name }}
          run: |
            echo "Hello, World!"
      - name: Delete VM
        if: always()
        uses: parallels/parallels-desktop-github-action@v1
        with:
          operation: 'delete'
          username: ${{ secrets.PARALLELS_USERNAME }}
          password: ${{ secrets.PARALLELS_PASSWORD }}
          host_url: devops.example.com
          machine_name: ${{ steps.pull.outputs.machine_name }}
```
