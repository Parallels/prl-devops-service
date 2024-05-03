---
layout: page
title: Control Multiple Hosts
subtitle: Manage multiple hosts seamless in a single pane
show_sidebar: false
toc: true
side_toc: true
---

# Control Multiple Hosts from a Single Interface

## Introduction

Controlling a single host remotely has many benefits, such as using it as a build machine, test machine, or a production machine. However, managing multiple hosts can be complex and challenging. You need to keep track of each host's resources, health status, and the virtual machines running on them. To simplify this process, you can make use of the Orchestrator service.

The Orchestrator service is an essential part of the DevOps service and starts automatically with the API. It eliminates the complexity of running different binaries and services on the same host. You can use it to add a host to the pool of hosts, which enables you to perform the same operations you would do with a single host, but with the added benefit of automating the selection of the host with enough resources to run the virtual machine and needed architecture.

In this guide, we will show you how to use the Orchestrator service to control multiple hosts from a single interface. We will use a CI/CD pipeline running a simple test on a virtual machine as an example. The pipeline will be triggered by a push to a GitHub repository and will run on a host chosen by the Orchestrator service.

### Architecture

This will show you a small diagram of what we will be setting up and how it will work.

![Github Runner]({{ site.url }}{{ site.baseurl }}/img/examples/orchestrator_ci_cd_github_scenario-simple_github_flow.drawio.png)

**Let's get started!**

## Requirements

- [Parallels Desktop for Mac](https://www.parallels.com/products/desktop/)
- [Parallels Desktop DevOps Service](https://github.com/Parallels/prl-devops-service)
- [GitHub account and repository](https://github.com)
- Multiple hosts with **Parallels Desktop** and **Parallels Desktop DevOps Service** installed
- Rest Client, like [Postman](https://www.postman.com/)
- Docker (Optional)

{% include notification.html message="You can try **Parallels Desktop** for free before purchasing by clicking this [link](https://www.parallels.com/products/desktop/trial/) and downloading our trial version" status="is-success" icon="comments-dollar" %}

<div style="margin-top:10px">
{% include notification.html message="While you can follow this guide with only one host, it is best to have multiple hosts to see the full potential" status="is-warning" icon="info" %}
</div>

## Step 1: Setting up the Orchestrator Service

The orchestrator service can be run from any device and is not restricted to a macOS host, as it is not used to create virtual machines. There are multiple ways to run the orchestrator service, and in this guide, we will be utilizing a docker container. Additionally, we offer a `Helms Chart` that can be used to deploy the orchestrator service on a Kubernetes cluster if you have one.

### Running the Orchestrator Service

To run the orchestrator service you can use the following command:

```bash
docker run --name pd-devops-orchestrator -p 8080:8080 -e ROOT_PASSWORD=VeryStr0ngPassw0rd -e API_PORT=8080 -e MODE=orchestrator -d cjlapao/prl-devops-service
```

This command will start the orchestrator service on port 8080 with the root password set to `VeryStr0ngPassw0rd`. You can change the password by changing the `ROOT_PASSWORD` environment variable.

### Accessing the Orchestrator Service

Once the docker container is working you can access the orchestrator using any rest client, like Postman. The orchestrator service has a REST API that can be used to interact with it. The API documentation can be found at `http://localhost:8080/swagger`.

We also make available a Postman collection that you can import and use to interact with the orchestrator service. You can download it [here]({{ site.url }}{{ site.baseurl }}/Parallels_Desktop_API.postman_collection.json).

## Step 2: Installing the DevOps Service on the Hosts

For each host we want to add to the orchestrator service, we need to install the DevOps service. The DevOps service is responsible for creating and managing virtual machines on the host.

### Install Parallels Desktop DevOps Service

To install the required tools, such as **Parallels Desktop for Mac**, we'll be using one of the features of the DevOps service. You can download the latest version of the Parallels Desktop DevOps Service by selecting the platform you are using and clicking the download button for the binary.

{% include inner-tabs.html data="download_tabs" %}

After downloading the binary, copy it to a directory in your path, such as /usr/local/bin, and make it executable.

```bash
sudo mv prldevops /usr/local/bin
sudo chmod +x /usr/local/bin/prldevops
```

You now have the necessary access to execute the `prldevops` command directly from your terminal.

Please check the configuration documentation for the DevOps service [here]({{ site.url }}{{ site.baseurl }}/docs/getting-started/configuration) to configure the service correctly. For the rest of the steps is important to set the **root password** to a known value as we will need it to add the host to the orchestrator service. We also will need to know the **API port** that the DevOps service is running on.

### Install Parallels Desktop for Mac

After installing the DevOps service, we can use it to install **Parallels Desktop for Mac** by running this command:

```sh
prldevops install parallels-desktop
```

This will download and install the latest version of **Parallels Desktop for Mac** and install it on your system.

## Step 3: Adding Hosts to the Orchestrator Service

Now that we have the orchestrator service running and the DevOps service installed on the hosts, we can add the hosts to the orchestrator service.

### Adding Hosts

To add a host to the orchestrator service, we need to make a POST request to the `/api/v1/orchestrator/hosts` endpoint. The request body should contain following information:

```json
{
    "host": "host1.example.com",
    "description": "Example Host 1",
    "tags": [
        "Builder"
    ],
    "authentication": {
        "username": "root@localhost",
        "password": "VeryStr0ng"
    }
}
```

You can repeat this process for each host you want to add to the orchestrator service. The orchestrator service will automatically detect the resources available on each host.

### Listing Hosts

To list the hosts added to the orchestrator service, you can make a GET request to the `/api/v1/orchestrator/hosts` endpoint. and this should return a list of hosts with their resources.

### Resources List

You can also request how much resources are available in the orchestrator service by making a GET request to the `/api/v1/orchestrator/overview/resources` endpoint. This will return a list of resources available in the orchestrator service.

## Step 4: Creating a CI/CD Pipeline

Now that we have our hosts added to the orchestrator service, we can create a CI/CD pipeline that will run a simple test on a virtual machine. We will use GitHub Actions to create the pipeline. This will just be a simple example to show how the orchestrator service can be used.

```yaml
name: Test Pipeline

on:
  workflow_dispatch:

jobs:
  create_runner:
    # This will be the name of the job
    name: Deploy Self Hosted Runner
    # This job will be running on the ubuntu_builder runner self hosted runner
    runs-on: 
    - self-hosted
    - ubuntu-latest
    # This will be the steps that will be executed on this job
    steps:
        # First we will be cloning the generic VM that we created before, in this case we will be cloning
        # a macOs Sonoma VM using our devops service, you can check the documentation on how to do this
        # in the official repository of the GitHub action 
        - name: Clone VM
          id: clone
          uses: parallels/parallels-desktop-github-action@v1
          with:
            operation: 'clone'
            # The username to access the devops service
            username: ${{ secrets.PARALLELS_USERNAME }}
            # The password to access the devops service
            password: ${{ secrets.PARALLELS_PASSWORD }}
            # Set this to true if you are using a self signed certificate
            insecure: true
            # the orchestrator for the devops service
            orchestrator_url: example.com:8080
            # The name for the VM that we will be cloning
            base_vm: macOS_action_runner_builder
            # If we should start the VM after cloning
            start_after_op: true
        # This step we will be setting up the runner on the VM that we cloned
        - name: Configure Github Runner
          id: configure  
          uses: parallels/parallels-desktop-github-action@v1
          with:
            # How many retries we should do if the operation fails on each execution
            max_attempts: 5
            # What will be the timeout after each failed attempt and before the next one
            timeout_seconds: 2
            operation: 'run'
            username: ${{ secrets.PARALLELS_USERNAME }}
            password: ${{ secrets.PARALLELS_PASSWORD }}
            orchestrator_url: example.com:8080
            insecure: true
            machine_name: ${{ steps.pull.outputs.vm_id }}
            # These will be running some scripts we have created to install the runner in an automated way
            # you can copy this step and change only the secrets to use the scripts that we have created
            run: |
                curl -o /Users/install-runner.sh https://raw.githubusercontent.com/Parallels/prlctl-scripts/main/github/actions-runner/mac/install-runner.sh
                curl -o /Users/configure-runner.sh https://raw.githubusercontent.com/Parallels/prlctl-scripts/main/github/actions-runner/mac/configure-runner.sh
                curl -o /Users/remove-runner.sh https://raw.githubusercontent.com/Parallels/prlctl-scripts/main/github/actions-runner/mac/remove-runner.sh
                chmod +x /Users/install-runner.sh
                chmod +x /Users/configure-runner.sh
                chmod +x /Users/remove-runner.sh

                /Users/install-runner.sh -u parallels -p /Users
                /Users/configure-runner.sh -u parallels -p /Users/action-runner -o Locally-build -t ${{ secrets.GH_PAT }} -n ${{ github.run_id }}_builder -l ${{ github.run_id }}_builder
   # The GitHub action will output the ID and the name of the VM that we cloned, this will be used later on
   # for the cleanup job
   outputs:
        vm_id: ${{ steps.pull.outputs.vm_id }}
        vm_name: ${{ steps.pull.outputs.vm_name }}
```