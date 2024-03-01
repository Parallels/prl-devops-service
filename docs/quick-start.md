---
layout: page
title: Quick Start
subtitle: Quickly spin up a Parallels Desktop DevOps Service
show_sidebar: false
version: 0.5.4
---

{% assign myOtherPost = site.pages | where:"url", "/docs/getting-started/installation/" | first %}

{{ myOtherPost.content }}

### Quick Configuration

By default the **devops service** will run with default values but these can be configured by creating a `config.yaml` file in the same directory as the binary. Below is an example of a configuration file.

```yaml
environment:
  api_port: 5570
  log_level: DEBUG
```

This is the most basic configuration file, you can find more information about the configuration file in [here]({{ site.url }}{{ site.baseurl }}/docs/getting-started/configuration).

### Running the Service

Once you have the binary you can run it by executing the following command.

```powershell
prldevops api
```

This will start the service and you will be able to access the swagger ui at [http://localhost:5570/swagger/index.html](http://localhost:5570/swagger/index.html)

**Note:** The service will run on port 5570 as defined in the above configuration example, you can change this by modifying the `config.yaml` file.

### Checking if the service is running

Once you started it you can then quickly check the health status of the service by running the following command.

```powershell
curl http://localhost:5570/api/health/probe
```

### Getting Help

If you need help with the service you can run the following command to get a list of all the available commands.

```powershell
prldevops --help
```

The help command is also context aware, so if you run `prldevops <command> --help` it will give you more information about that specific command.

For example:

```powershell
prldevops api --help
```

<div class="flex-center"><img alt="prldevops help" src="{{ site.url }}{{ site.baseurl }}/assets/anims/prdevops_help.gif" /></div>

### Installing Tools

The service will require some tools to be installed on the host machine for certain features to work.
If you just want to use the orchestrator and catalog features then this is ready to use but if you want to use the virtual machine management features then you will need to install the Parallels Desktop for Mac application.
You also might need `Hashicorp Packer` and `Vagrant` installed on your machine if you want to create virtual machines from a packer template, a vagrant box or a vagrantfile.

For this we have created a special command that will check if the tools are installed and if not it will install them for you.

```powershell
prldevops install
```

### Next Steps

You now have the service running in your machine, you can start playing with all the available features and start creating virtual machines and managing your hosts.

To do so follow the links below to get more information about the available features.

- [API]({{ site.url }}{{ site.baseurl }}/docs/restapi/overview/)
- [Orchestrator]({{ site.url }}{{ site.baseurl }}/docs/orchestrator/overview/)
- [Catalog]({{ site.url }}{{ site.baseurl }}/docs/catalog/overview/)
- [Reverse Proxy]({{ site.url }}{{ site.baseurl }}/docs/reverse-proxy/overview/)
