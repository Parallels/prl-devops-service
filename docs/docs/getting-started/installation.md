---
layout: page
title: Getting Started
subtitle: Installation
menubar: docs_menu
show_sidebar: false
toc: true
---

Devops Service is a command line tool that allows you to manage and orchestrate multiple Parallels Desktop hosts and virtual machines. It will allow you to create, start, stop and delete virtual machines and will also allow you to manage the hosts that are running the virtual machines.

You can get the latest version by downloading the pre compiled binaries from below or by building it from the source code.

### Download and Install

To download the latest version of the Parallels Desktop DevOps Service, simply select the platform you are running and click the button to download the binary.

{% include inner-tabs.html content="download_tabs" %}

### Build from source

You will need to clone the repository from github and then build the project using the following commands.

```powershell
git clone https://github.com/Parallels/prl-devops-service
cd prl-devops-service
make build
```

this will create the binary in the `out/binaries` folder. you can then execute the binary from there, or move it to a location in your path.
