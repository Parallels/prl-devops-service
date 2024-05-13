---
layout: page
title: Getting Started
subtitle: Installation
menubar: docs_menu
show_sidebar: false
toc: true
---

Devops Service is a command line tool that allows you to manage and orchestrate multiple Parallels Desktop hosts and virtual machines. It will allow you to create, start, stop and delete virtual machines and will also allow you to manage the hosts that are running the virtual machines.

To install the required tools, such as **Parallels Desktop for Mac**, we'll be using one of the features of the DevOps service. You can download the latest version of the Parallels Desktop DevOps Service by selecting the platform you are using and clicking the download button for the binary.

{% include inner-tabs.html data="download_tabs" %}

After downloading the binary, copy it to a directory in your path, such as /usr/local/bin, and make it executable.

```bash
sudo mv prldevops /usr/local/bin
sudo chmod +x /usr/local/bin/prldevops
```
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
  api_port: 5570
  log_level: DEBUG
```

#### Start it as a service

We can also install it as a daemon to run in the background.

```sh
sudo prldevops install service /usr/bin/prldevops
launchctl start com.parallels.devops-service
```

This command will start the service, and the REST API can then be accessed at `http://localhost:80`. 

{% include notification.html message="To change the configuration you will need to stop the daemon before doing any changes." status="is-info" icon="info" %}