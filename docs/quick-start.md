---
layout: page
title: Quick Start
subtitle: Quickly spin up a Parallels Desktop DevOps Service
show_sidebar: false
version: 0.5.4
---

Devops Service is a command line tool that allows you to manage and orchestrate multiple Parallels Desktop hosts and virtual machines. It will allow you to create, start, stop and delete virtual machines and will also allow you to manage the hosts that are running the virtual machines.

You can get the latest version by downloading the pre compiled binaries from below or by building it from the source code.

### Download and Install

<div class="tabs is-boxed">
  <ul>
    <li class="tab" onclick="openTab(event, 'mac')">
      <a>
        <span class="icon is-medium">
          <i class="fa-brands fa-apple fa-xl"></i>
        </span>
        <span>Mac</span>
      </a>
    </li>
    <li class="tab" onclick="openTab(event,'linux')">
      <a>
        <span class="icon is-medium">
          <i class="fa-brands fa-linux fa-xl"></i>
        </span>
        <span>Linux</span>
      </a>
    </li>
    <li class="tab" onclick="openTab(event, 'windows')">
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
    <div id="mac" class="content-tab" style="display:none">
      <p>
        You can get the latest version of the Parallels Desktop DevOps Service by clicking the button below.
        This will download the latest version of the Parallels Desktop DevOps Service for Mac.
      </p>
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
    <div id="windows" class="content-tab" style="display:none">
      <p>
        {% include notification.html message="At the moment we do not provide any binaries for windows" status="is-warning" %}
      </p>
    </div>
    <div id="linux" class="content-tab" style="display:none">
      <p>
        You can get the latest version of the Parallels Desktop DevOps Service by clicking the button below.
        This will download the latest version of the Parallels Desktop DevOps Service.
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

### Quick Configuration

By default the **devops service** will run with default values but these can be configured by creating a `config.yaml` file in the same directory as the binary. Below is an example of a configuration file.

```yaml
environment:
  api_port: 5570
  log_level: DEBUG
```

This is the most basic configuration file, you can find more information about the configuration file in the [configuration section of the README.md]

### Running the Service

Once you have the binary you can run it by executing the following command.

```shell
./prldevops
```

This will start the service and you will be able to access the swagger ui at [http://localhost:5570/swagger/index.html](http://localhost:5570/swagger/index.html)

**Note:** The service will run on port 5570 as defined in the above configuration example, you can change this by modifying the `config.yaml` file.

### Checking if the service is running

Once you started it you can then quickly check the health status of the service by running the following command.

```shell
curl http://localhost:5570/api/health/probe
```

### Next Steps

You now have t