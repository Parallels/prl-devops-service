---
layout: page
title: Getting Started
subtitle: Installation
menubar: docs_menu
show_sidebar: false
toc: true
---

# Install Parallels Desktop DevOps

Devops Service is a command line tool that allows you to manage and orchestrate multiple Parallels Desktop hosts and virtual machines. It will allow you to create, start, stop and delete virtual machines and will also allow you to manage the hosts that are running the virtual machines.

You can download the latest version of the **Parallels Desktop DevOps Service** by selecting the platform you are using and clicking the download button for the binary.

{% include inner-tabs.html data="download_tabs" %}


### Quick Configuration

By default the **devops service** will run with default values but these can be configured by creating a `config.yaml` file in the same directory as the binary. Below is an example of a configuration file.

to create a configuration file you can run the following command

```powershell
touch /usr/local/bin/config.yaml
open -a TextEdit /usr/local/bin/config.yaml
```

You can then add the following basic configuration to the file, you can find more information about the configuration file in [here]({{ site.url }}{{ site.baseurl }}/docs/getting-started/configuration){:target="_blank"}.

```yaml
environment:
  api_port: 80
  log_level: DEBUG
  ROOT_PASSWORD: VeryStr0ngPassw0rd
```

#### Start it as a service

We can also install it as a daemon to run in the background.

```sh
sudo prldevops install service /usr/bin/prldevops
launchctl start com.parallels.devops-service
```

This command will start the service, and the REST API can then be accessed at `http://localhost:80`. 

{% include notification.html message="To change the configuration you will need to stop the daemon before doing any changes." status="is-info" icon="info" %}

### Checking if the service is running

Once you started it you can then quickly check the health status of the service by either running the following command.

```powershell
curl http://localhost:80/api/health/probe
```

or to go to the browser and navigate to the [swagger page](http://localhost:80//swagger/index.html){:target="_blank"}.

We also make available a Postman collection that you can import and use to interact with the `DevOps Service`. You can download it [here]({{ site.url }}{{ site.baseurl }}/Parallels_Desktop_API.postman_collection.json){:target="_blank"}.


### Running the service as a container

You can find more about how to run the service as a container in the [Docker documentation]({{ site.url }}{{ site.baseurl }}/docs/getting-started/docker/){:target="_blank"} win the Getting Started section.
