---
layout: page
title: Catalog
subtitle: Manage an remote Vms Catalog for ease sharing
show_sidebar: false
toc: false
hero_link: /docs/catalog/overview/
hero_link_text: Documentation
---

# What is It?

Our DevTools Catalog Service is a repository that stores only metadata about a specific virtual machine. This metadata is saved within the catalog and the virtual machine package in a remote storage provider, reducing traffic between the client and the service. The service can also leverage different storage providers simultaneously. 

# How does it work?

The Service is responsible for managing the metadata and granting access to it. Once the details are confirmed, the service sends the necessary information to the client. 

The client then uses this information to upload or download the virtual machine package from where it is stored, making the service lightweight.

For a download or pulling operation, this is how the flow would normally look like:

Open catalog_diagram-pulling_diagram.drawio.png
catalog_diagram-pulling_diagram.drawio.png
For an upload or pushing operation, this is how the flow would normally look like:

Open catalog_diagram-pushing_diagram.drawio.png
catalog_diagram-pushing_diagram.drawio.png
The Catalog Service can be run on any operating system. Additionally, we provide a Docker container.