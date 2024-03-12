---
layout: page
title: Quick Start
subtitle: Quickly spin up a Parallels Desktop DevOps Service
show_sidebar: false
version: 0.5.4
toc: true
side_toc: true
download_ubuntu_tab:
  style: boxed
  icon_size: xl
  items:
    - title: x86_64
      icon: microchip 
      content: |-
        <div class="mt-3">
            <a href="https://releases.ubuntu.com/22.04.4/ubuntu-22.04.4-live-server-amd64.iso" class="button is-primary is-medium">
                Download Ubuntu Server 22.04 LTS for x86_64
            </a>
        </div>
    - title: arm64
      icon: microchip 
      content: |-
        <div class="mt-3">
            <a href="https://cdimage.ubuntu.com/releases/22.04/release/ubuntu-22.04.4-live-server-arm64.iso" class="button is-primary is-medium">
                Download Ubuntu Server 22.04 LTS for arm64
            </a>
        </div>
---

# Securely create and share Golden Master images with your team in minutes

## Introduction

One of the key challenges of managing a large number of virtual machines is ensuring that all of them are running the same software stack. This is especially true when you have a large team of developers, each of whom may have their own unique development environment.

There is also the concern of ensuring that the software stack is secure and up-to-date. This is where Golden Master images come in. A Golden Master image is a virtual machine that has been configured with the software stack that you want to use. This image can then be shared with your team, ensuring that everyone is using the same software stack.


Lets take a look at how you can use the Parallels Desktop DevOps Service to create a Golden Master image and share it with your team. We will be creating an Ubuntu Virtual Machine and then installing the required software stack. We will then use the Parallels Desktop DevOps Service to create a Golden Master image of the virtual machine and share it with our team using the Parallels Desktop DevOps Service and its [Remote Catalog]({{ site.url }}{{ site.baseurl }}/docs/catalog) capability.

**So lets get started!**

## Creating a local Golden Master image

### Requirements

- [Parallels Desktop for Mac](https://www.parallels.com/products/desktop/)
- [Parallels Desktop DevOps Service](https://github.com/Parallels/prl-devops-service)

{% include notification.html message="You can use our free trial to test it before you buy using this [Link](https://www.parallels.com/products/desktop/trial/) " status="is-success" %}

### Install Parallels Desktop DevOps Service

We will be using one of the features of the DevOps service for installing the required tools, in this case the Parallels Desktop for Mac. You will need to download the latest version of the Parallels Desktop DevOps Service. You can do this by selecting the platform you are running and clicking the button to download the binary.

{% include inner-tabs.html data="download_tabs" %}

Once you have downloaded the binary, copy it to a location in your path for example `/usr/local/bin` and make it executable.

```bash
sudo mv prl-devops-service /usr/local/bin
sudo chmod +x /usr/local/bin/prl-devops-service
```

You should now have access to execute the `prldevops` command from your terminal.

### Install Parallels Desktop for Mac

Now that we have the DevOps service installed, we can use it to install Parallels Desktop for Mac. You can do this by running the following command:

```sh
prldevops install parallels-desktop
```

This will download and install the latest version of Parallels Desktop for Mac and install it on your system. 

### Create a Virtual Machine

We now have our virtualization software installed, so we can create a virtual machine. We will be creating an Ubuntu 22.04 Server virtual machine for this example.

#### Download the ISO for your architecture

You can download the ISO from the [Ubuntu website](https://ubuntu.com/download/server) or use the links below:

{% include inner-tabs.html content='download_ubuntu_tab' %}


#### Install Ubuntu

Once you have downloaded the ISO, you can create a new virtual machine using the following command line

```sh
prlctl create "test-vm" -d ubuntu
prlctl set "test-vm" --cpus 2
prlctl set "test-vm" --memsize 2048
prlctl set "test-vm" --device-set hdd0 --size 64G
```

Now we need to set the ISO file as the boot device and start the virtual machine and start it.

```sh
prlctl set "test-vm" --device-set cdrom0 --image /path/to/ubuntu-22.04.4-live-server-amd64.iso --connect
prlctl set "test-vm" --device-bootorder "cdrom0 hdd0"
prlctl start "test-vm"
```

#### Installing Parallels Tools

Some of the functionality required for this example needs the Parallels Tools installed. You can install the tools by following the on screen instructions. Unfortunately we will not have any copy and paste functionality until the tools are installed so you will need to go to the machine, login with the user you setup and type in the following commands:

First lets remove any mounted cdroms and create a mount point for the cdrom0 device
```sh
for dev in /dev/sr0 /dev/cdrom /dev/dvd; do sudo eject $dev; done 2>/dev/null
sudo mkdir /media/cdrom0
```

Now lets add the `Parallels Desktop Tools` iso using the ui

<div class="flex flex-center">
  <img src="{{ site.url }}{{ site.baseurl }}/img/examples/install_tools_step03.gif" alt="Install Parallels Tools step 3"/>
</div>

Finally we can install the tools by running the following command

```sh
sudo mount -o exec /dev/sr0 /media/cdrom0
sudo /media/cdrom0/install
reboot
```

You should now have the *parallels desktop tools* installed and running.

For more information you can follow the [official documentation](https://kb.parallels.com/129740)

#### Customize the Golden Image

Install the operating system by following the on screen instructions. once the installation is complete, you can login to the virtual machine and install the required software stack. In our case we will be installing the following:

- Docker
- Node.js
- Python
- Dotnet SDK

First lets jump into the machine using our terminal and the command line client.

```sh
prlctl enter "test-vm" \
```

now we can install the required software stack

```sh
sudo apt update && sudo apt upgrade -y
sudo apt install -y docker.io nodejs npm python3 git dotnet-sdk-8.0
```

Once the installation is complete, you should now have a virtual machine with the required software stack installed ready for use.

## Configuring and Running our DevOps Remote Catalog

The DevOps Remote Catalog is a service designed to facilitate the sharing of VMs or Golden Master Images in your organization, if you want to know more on how it works you can check the [official documentation]({{ site.url }}{{ site.baseurl }}/docs/catalog/overview/)

{% include notification.html message="On this example we will be setting up **Parallels Desktop DevOps Remote Catalog** service in a docker container running locally, you can deploy it as a docker container in the cloud or as a daemon service in any remote macOS" status="is-info" %}

### Configuring the DevOps Remote Catalog

To configure the service we will be adding a [configuration file]({{ site.url }}{{ site.baseurl }}/docs/getting-started/configuration/) to the service. You can use the following example as a starting point:

Run the following command to create the configuration file:

```sh
touch config.yml
```

then open the file with your favorite terminal text editor  and add the following content:

```yaml
environment:
  api_port: 80
  mode: catalog
```

This will configure the service to run in catalog mode and listen on port 80, with all the default settings and while this is ok for a quick start, we will need to harden the service for production use.
We will be setting up some extra security settings to allow a more secure deployment, you can find more information on how to do this in the [official documentation]({{ site.url }}{{ site.baseurl }}/docs/getting-started/configuration/)

#### Setting up database encryption

By default the service runs with an unencrypted database, this is of course not desirable for a production deployment, to enable this encryption we will need to generate a private key and set it in the configuration file.

We use any type of **RSA** key base64 encoded, if you do not have any, you can generate one using the `prldevops` command line tool.:

```sh
prldevops gen-rsa --file=private.pem
cat private.pem | base64
```

<div class="flex flex-center">
  <img src="{{ site.url }}{{ site.baseurl }}/img/examples/base64_db_key.png" alt="Database base64 RSA key"/>
</div>

you should get a very long string, copy it and add it to the configuration file:

```yaml
environment:
  api_port: 80
  mode: catalog
  encryption_private_key: <base64_encoded_private_key>
```

#### Setting up TLS

Now that we have our database encrypted, we can also enable TLS to secure the communication between the service and the clients, to do this you will need to have an SSL certificate and private key. Once you have them you can base64 encode them and add them to the configuration file:

```yaml
environment:
  api_port: 80
  mode: catalog
  encryption_private_key: <base64_encoded_private_key>
  tls_enabled: true
  tls_port: 443
  tls_certificate: <base64_encoded_certificate>
  tls_private_key: <base64_encoded_private_key>
```

#### Setting up the JWT signing method

The last thing we will be setting up is the JWT signing method, this is used to sign the JWT tokens that the service uses for authentication. We allow different signing methods, and on this example we will be using the `HMACS` method.

This consists of a secret that will be used to sign the tokens, similar to a password but should be really big and kept secret. This is used in the background to sign the tokens so you will not need to memorize it.

```yaml
environment:
  api_port: 80
  mode: catalog
  encryption_private_key: <base64_encoded_private_key>
  tls_enabled: true
  tls_port: 443
  tls_certificate: <base64_encoded_certificate>
  tls_private_key: <base64_encoded_private_key>
  jwt_hmac_secret: VeryStr0ngS3cr3t
  jwt_sign_method: HS256
```

the final step is to copy the file to the service folder:

```sh
cp config.yml /usr/local/bin/config.yml
```

### Changing the Root password

By default the service runs with a random root password and you will need to change this in order to access the service rest api, we provider a simple command line to make this change:

```sh
prldevops update-root-password --password=VeryStr0ngP@ssw0rd
```

### Starting the service

we can easily start the service by running the following command:

```sh
prldevops
```

we can also install it as a daemon to allow it to run in the background:

```sh
sudo prldevops install service /usr/bin/prldevops
launchctl start com.parallels.devops-service
```

This should bring the service up and running, you can now access the service rest api on `http://localhost:80`

## Pushing the Golden Master Image to the Remote Catalog

Now that we have our `DevOps Remote Catalog` service running, we can push the Golden Master image to the service.
but before we start we need to get some requirements out of the way:

The way the `DevOps Remote Catalog` works is by only storing the metadata of the Golden Master image, this means that the actual image will be stored in a remote location, for this example we will be using a S3 bucket to store the image, you can use any other compatible storage service, like `Azure Blob Storage` or `jfrog artifactory`. Most of these providers provide a free tier that you can use to test this feature.

You can find out more about the architecture [here]({{ site.url }}{{ site.baseurl }}/docs/catalog/architecture/)

### Creating the PDFile

We provide a easy way to automate the `push` and `pull` process by using a reusable file called `pdfile`, this file contains all the information required to push and pull vms from our catalog. It is very similar to a dockerfile, but for vms.

create a pdfile with the following content:

```pdfile
TO localhost
INSECURE true

AUTHENTICATE USERNAME root
AUTHENTICATE PASSWORD VeryStr0ngP@ssw0rd
AUTHENTICATE USERNAME1 root

PROVIDER NAME aws-s3
PROVIDER BUCKET <bucket_name>
PROVIDER REGION <bucket_region>
PROVIDER ACCESS_KEY <bucket_access_key>
PROVIDER SECRET_KEY <bucket_secret_key>

LOCAL_PATH /path/to/the/test-vm.pvm

DESCRIPTION Build Machine

TAG ubuntu, build-machine, arm64, docker, python, nodejs

CATALOG_ID ubuntu-22-04-builder
VERSION v1
ARCHITECTURE arm64
```

This would push the vm in the `LOCAL_PATH` to the `localhost` *remote catalog* with everyone having access to it, but as we previously said we can restrict access to specific claims and roles for each catalog entry, to achieve this we simply can add the `ROLE` and/or `CLAIM` into the `pdfile`, for example the same but saying the user needs the `developer` claim to be able to access it

```pdfile
TO localhost
INSECURE true

AUTHENTICATE USERNAME root
AUTHENTICATE PASSWORD VeryStr0ngP@ssw0rd
AUTHENTICATE USERNAME1 root

PROVIDER NAME aws-s3
PROVIDER BUCKET <bucket_name>
PROVIDER REGION <bucket_region>
PROVIDER ACCESS_KEY <bucket_access_key>
PROVIDER SECRET_KEY <bucket_secret_key>

LOCAL_PATH /path/to/the/test-vm.pvm

DESCRIPTION Build Machine

CLAIM developer

TAG ubuntu, build-machine, arm64, docker, python, nodejs

CATALOG_ID ubuntu-22-04-builder
VERSION v1
ARCHITECTURE arm64
```

you can then just run the following command to push the image to the catalog:

```sh
prldevops push --file=<filename>
```

{% include notification.html message="This operation can take a while depending on the size of the vm and your internet connection" status="is-warning" %}

### Pulling the Golden Master Image

Now the last part, we need a way of sharing this Golden Image, we can do this in a similar way like pushing using a `pdfile` designed for this purpose, this file can be shared or even stored in a repository for easy access as it does not contain any sensitive information.

```pdfile
FROM localhost
INSECURE true

CATALOG_ID ubuntu-22-04-builder
VERSION v1
ARCHITECTURE arm64

MACHINE_NAME build-machine

OWNER cjlapao
DESTINATION /Users/foo/Parallels
START_AFTER_PULL false
```

{% include notification.html message="This operation can take a while depending on the size of the vm and your internet connection" status="is-warning" %}

After this you should have a new vm and ready top be used with all of the stack installed.