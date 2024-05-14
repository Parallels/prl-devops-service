---
layout: page
title: Quick Start
subtitle: Quickly spin up a Parallels Desktop DevOps Service
show_sidebar: false
version: 0.5.4
---

{% assign installationContent = site.pages | where:"url", "/docs/getting-started/installation/" | first %}

{{ installationContent.content }}




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

<div class="flex flex-center">
  <img src="{{ site.url }}{{ site.baseurl }}/img/examples/prldevops_help.gif" alt="Catalog List"/>
</div>

![DevOps Help]({{ site.url }}{{ site.baseurl }}/img/prldevops_help.gif)

### Installing Tools

The service will require some tools to be installed on the host machine for certain features to work.
If you just want to use the orchestrator and catalog features then this is ready to use but if you want to use the virtual machine management features then you will need to install the Parallels Desktop for Mac application.
You also might need `Hashicorp Packer` and `Vagrant` installed on your machine if you want to create virtual machines from a packer template, a vagrant box or a vagrantfile.

For this we have created a special command that will check if the tools are installed and if not it will install them for you.

```powershell
prldevops install
```

### Next Steps

You now have the service installed and running, you can dive a bit more into our [official documentation]({{ site.url }}{{ site.baseurl }}/docs/) to see what else you can do with the service. 

You can also check our examples and tutorials to see how you can use the service to automate your virtual machine management.

[Run a Catalog service]({{ site.url }}{{ site.baseurl }}/examples/remote-catalog/)  
[Control Multiple Hosts]({{ site.url }}{{ site.baseurl }}/examples/orchestrator/)  
[Github Actions]({{ site.url }}{{ site.baseurl }}/examples/github-actions/)  