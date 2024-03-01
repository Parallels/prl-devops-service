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

<div class="tabs is-boxed">
  <ul>
    <li id="download_tab_mac" class="tab" onclick="openTab(event, 'download_tab_mac_content')">
      <a>
        <span class="icon is-medium">
          <i class="fa-brands fa-apple fa-xl"></i>
        </span>
        <span>Mac</span>
      </a>
    </li>
    <li id="download_tab_linux" class="tab" onclick="openTab(event,'download_tab_linux_content')">
      <a>
        <span class="icon is-medium">
          <i class="fa-brands fa-linux fa-xl"></i>
        </span>
        <span>Linux</span>
      </a>
    </li>
    <li id="download_tab_windows" class="tab" onclick="openTab(event, 'download_tab_windows_content')">
      <a>
        <span class="icon is-medium">
          <i class="fa-brands fa-windows fa-xl"></i>
        </span>
        <span>Windows</span>
      </a>
    </li>
  </ul>
</div>
<div class="container tab-container">
    <div id="download_tab_mac_content" class="content-tab" style="display:none">
      <div class="test">
        <span>
          <a href="https://github.com/Parallels/prl-devops-service/releases/download/release-v{{ page.version }}/prldevops--darwin-arm64.tar.gz" class="m-1 button is-primary">
            Parallels Desktop DevOps Service for Mac with Apple Silicon
          </a>
        </span>
        <span>
          <a href="https://github.com/Parallels/prl-devops-service/releases/download/release-v{{ page.version }}/prldevops--darwin-amd64.tar.gz" class="m-1 button is-primary">
            Parallels Desktop DevOps Service for Mac with Intel chip
          </a>
        </span>
      </div>
    </div>
    <div id="download_tab_windows_content" class="content-tab" style="display:none">
      <p>
        {% include notification.html message="At the moment we do not provide any binaries for windows" status="is-warning" %}
      </p>
    </div>
    <div id="download_tab_linux_content" class="content-tab" style="display:none">
      <p>
        {% include notification.html message="Please be aware that running this service in windows you will only have access to the orchestrator and catalog features" status="is-warning" %}
      </p>
      <div class="test">
        <span>
          <a href="https://github.com/Parallels/prl-devops-service/releases/download/release-v{{ page.version }}/prldevops--linux-amd64.tar.gz" class="m-1 button is-primary">
            Parallels Desktop DevOps Service for intel chips
          </a>
        </span>
        <span>
          <a href="https://github.com/Parallels/prl-devops-service/releases/download/release-v{{ page.version }}/prldevops--linux-arm64.tar.gz" class="m-1 button is-primary">
            Parallels Desktop DevOps Service for arm chips
          </a>
        </span>
      </div>
    </div>
</div>

### Build from source

You will need to clone the repository from github and then build the project using the following commands.

```shell
git clone https://github.com/Parallels/prl-devops-service
cd prl-devops-service
make build
```

this will create the binary in the `out/binaries` folder. you can then execute the binary from there, or move it to a location in your path.
