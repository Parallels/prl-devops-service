---
layout: page
title: Share Golden Images
subtitle: Securely create and share Golden Master images with your team in minutes
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
create_vm_tab:
  style: boxed
  icon_size: xl
  items:
    - title: With Control Center
      default: true
      file: tabs/create_vm_control_center
    - title: With Command Line Interface
      file: tabs/create_vm_cli
---

# Securely create and share Golden Master images with your team in minutes

## Introduction

One of the key challenges of managing many virtual machines is ensuring that all are running the same software stack. This is especially true when you have a large team of developers, each of whom may have their unique development environment.
There is also the concern of ensuring the software stack is secure and up-to-date. This is where Golden Master images come in. A Golden Master image is a virtual machine that has been configured with the software stack that you want to use. This image can then be shared with your team, ensuring everyone uses the same software stack.

Let's look at how you can use the Parallels Desktop DevOps Service to create a Golden Master image and share it with your team. We will create an Ubuntu Virtual Machine and install the required software stack. We will then use the Parallels Desktop DevOps Service to make a Golden Master image of the virtual machine and share it with our team using the Parallels Desktop DevOps Service and its [Remote Catalog]({{ site.url }}{{ site.baseurl }}/docs/catalog) capability.

**So lets get started!**

## Creating a local Golden Master image

### Requirements

- [Parallels Desktop for Mac](https://www.parallels.com/products/desktop/)
- [Parallels Desktop DevOps Service](https://github.com/Parallels/prl-devops-service)

{% include notification.html message="You can try it for free before purchasing by clicking this [link](https://www.parallels.com/products/desktop/trial/) and downloading our trial version" status="is-success" icon="comments-dollar" %}

### Install Parallels Desktop DevOps Service

To install the required tools, such as **Parallels Desktop for Mac**, we'll be using one of the features of the DevOps service. You can download the latest version of the Parallels Desktop DevOps Service by selecting the platform you are using and clicking the download button for the binary.

{% include inner-tabs.html data="download_tabs" %}

After downloading the binary, copy it to a directory in your path, such as /usr/local/bin, and make it executable.

```bash
sudo mv prldevops /usr/local/bin
sudo chmod +x /usr/local/bin/prldevops
```

You now have the necessary access to execute the `prldevops` command directly from your terminal.

### Install Parallels Desktop for Mac

After installing the DevOps service, we can use it to install **Parallels Desktop for Mac** by running this command:

```sh
prldevops install parallels-desktop
```

This will download and install the latest version of **Parallels Desktop for Mac** and install it on your system.

### Create a Virtual Machine

We have installed the virtualization software and can now create a virtual machine. For this example, we will create a virtual machine of Ubuntu 22.04 Server.

#### Download the ISO for your architecture

You can download the ISO from the [Ubuntu website](https://ubuntu.com/download/server) or use the links below:

{% include inner-tabs.html content='download_ubuntu_tab' %}

#### Install Ubuntu

After downloading the ISO, you can create a new virtual machine using either **Parallels Desktop Control Center** or the **Command Line Interface**.

{% include inner-tabs.html content="create_vm_tab" %}

#### Installing Parallels Tools

To enable certain features in this example, the Parallels Tools must be installed. You can do this by following the instructions that appear on the screen. Unfortunately, until the tools are installed, you won't be able to use the copy and paste function. In order to proceed, please go to the machine, login with the user you created, and enter the following commands:

Let's start by unmounting any currently mounted cdroms and creating a new mount point for the `cdrom0` device.
```sh
for dev in /dev/sr0 /dev/cdrom /dev/dvd; do sudo eject $dev; done 2>/dev/null
sudo mkdir /media/cdrom0
```

Now lets add the `Parallels Desktop Tools` iso using the ui

<div class="flex flex-center">
  <img src="{{ site.url }}{{ site.baseurl }}/img/examples/install_tools_step03.gif" alt="Install Parallels Tools step 3"/>
</div>

To install the tools, run the command given below: 

```sh
sudo mount -o exec /dev/sr0 /media/cdrom0
sudo /media/cdrom0/install
reboot
```

You should have **Parallels Desktop Tools** installed and running. For more information, follow the [official documentation](https://kb.parallels.com/129740)

#### Customize the Golden Image

Follow the on-screen instructions to install the operating system. Once the installation is complete, log in to the virtual machine and install the required software stack. In our case, we will install the following:

- Docker
- Node.js
- Python
- Dotnet SDK

To start, we need to access the machine through the terminal using the command line client. Use the following command:

```sh
prlctl enter "test-vm"
```

Once inside, we can proceed with installing the necessary software stack by running the following commands: 

```sh
sudo apt update && sudo apt upgrade -y
sudo apt install -y docker.io nodejs npm python3 git dotnet-sdk-8.0
```

After the installation is complete, you will have a virtual machine with the required software stack installed, and it will be ready for use.

{% include notification.html message="You can now shut down the virtual machine as we do not need it running anymore." status="is-success" icon="lightbulb" %}

## Configuring and Running our DevOps Remote Catalog

The DevOps Remote Catalog is a service designed to facilitate the sharing of VMs or Golden Master Images in your organization, if you want to know more on how it works you can check the [official documentation]({{ site.url }}{{ site.baseurl }}/docs/catalog/overview/)

{% include notification.html message="For this example, we will set up **Parallels Desktop DevOps Remote Catalog** service as a locally running daemon. You can deploy it as a docker container in the cloud or as a daemon service in any remote macOS." status="is-info" %}

### Configuring the DevOps Remote Catalog

To set up the service, we will need to create a [configuration file]({{ site.url }}{{ site.baseurl }}/docs/getting-started/configuration/) . You can use the following example as a reference:

1. Run the command below to create the config file:

    ```sh
    touch config.yml
    ```

2. Open the file with your preferred terminal text editor and add the following content:

    ```yaml
    environment:
    api_port: 80
    mode: catalog
    ```

This will configure the service to run in catalog mode and listen on port 80 with all the default settings. However, this setup is only suitable for quick testing. For production use, we need to implement additional security measures to ensure a more secure deployment. You can find more information about security options in the [official documentation]({{ site.url }}{{ site.baseurl }}/docs/getting-started/configuration/)

#### Setting up database encryption

By default, the service operates with an unencrypted database. However, this is not ideal for a production deployment, as it poses a security risk. To enable encryption, you will need to generate a private `RSA key` and set it in the configuration file.
The key must be **base64 encoded**. If you don't have an `RSA key`, you can generate one using the `prldevops` command line tool.

```sh
prldevops gen-rsa --file=private.pem
cat private.pem | base64
```

<div class="flex flex-center">
  <img src="{{ site.url }}{{ site.baseurl }}/img/examples/base64_db_key.png" alt="Database base64 RSA key"/>
</div>

You should copy and add a very long string to the configuration file.

```yaml
environment:
  api_port: 80
  mode: catalog
  encryption_private_key: <base64_encoded_private_key>
```

#### Setting up TLS

After encrypting our database, we can further enhance security by enabling TLS to safeguard the communication between the service and clients. To do this, you will require an SSL certificate and private key. Once obtained, encode them to base64 and add to the configuration file.

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

The final step in our setup process involves configuring the *JWT* signing method, which is responsible for signing the *JWT* tokens used by the service for authentication. We offer a variety of signing methods, but for this example, we will be using the **HMACS** method. This involves using a large, secret key (similar to a password, but much longer) to sign the tokens. You won't need to remember the key, as it will be used automatically in the background.

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

We just need to transfer the configuration file to the service folder.

```sh
cp config.yml /usr/local/bin/config.yml
```

### Changing the Root password

Typically, the service is initiated with a randomly generated root password. However, in order to access the service's REST API, you will need to update this password. To do so, we provide a simple command line. For instance, we will set the password as `VeryStr0ngP@ssw0rd`, but you may choose any password you prefer.

```prldevops
prldevops update-root-password --password=VeryStr0ngP@ssw0rd
```

### Starting the service

To initiate the service, all you need to do is execute the command given below:

```prldevops
prldevops
```

We can also install it as a daemon to run in the background.

```sh
sudo prldevops install service /usr/bin/prldevops
launchctl start com.parallels.devops-service
```

This command will start the service, and the REST API can then be accessed at `http://localhost:80`.

## Pushing the Golden Master Image to the Remote Catalog

We now have our` DevOps Remote Catalog` service up and running, which means we can proceed to push the Golden Master image to the service. But before we start, we need to take care of a few requirements. 
The `DevOps Remote Catalog` works by storing only the metadata of the Golden Master image. This implies that the actual image will be stored in a remote location. In this example, we will be using an **S3 bucket** to store the image. However, you can use any other compatible storage service like **Azure Blob Storage** or **jfrog artifactory**. Most of these providers offer a free tier that you can use to test this feature. For more information about the architecture, you can refer to this link [here]({{ site.url }}{{ site.baseurl }}/docs/catalog/overview/#architecture).

### Creating the PDFile

We have designed an easy way to automate the `push` and `pull` process of virtual machines (VMs) using a reusable file called pdfile. This file contains essential information required to `push` and `pull` VMs from our catalog. It is similar to a dockerfile but specific for VMs.

To create a pdfile, you need to include the following content:

```pdfile
TO localhost
INSECURE true

AUTHENTICATE USERNAME root
AUTHENTICATE PASSWORD VeryStr0ngP@ssw0rd

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

This will push the VM from the *LOCAL_PATH* to the localhost *remote catalog* with everyone having access to it. However, we can restrict access to specific claims and roles for each catalog entry. To achieve this, you can add the `ROLE` and/or `CLAIM` into the pdfile. For example, if you want to restrict access to users with the **developer** claim, you can add the following to the pdfile

```pdfile
TO localhost
INSECURE true

AUTHENTICATE USERNAME root
AUTHENTICATE PASSWORD VeryStr0ngP@ssw0rd

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

To push the VM image to the catalog, you can run the following command:

```prldevops
prldevops push /path/to/pdfile
```

{% include notification.html message="This process duration may vary depending on your VM size and internet speed." status="is-warning" %}

### Pulling the Golden Master Image

Now that we have the Golden Master image in the catalog, we can allow our team to pull it and use it to create new virtual machines.

### Requirements

Your team now has access to the Golden Master image in the catalog, which they can use to create virtual machines. To be able to do this, they will need to have the Parallels Desktop DevOps Service and Parallels Desktop installed on their machines, using the same installation process as before. You can check it out **[here](#install-parallels-desktop-devops-service)**

### Pulling the Golden Master Image

To make it easy for the team members, we have created a pdfile that contains all the information required to pull the Golden Master image from the catalog. This file can be shared with anyone and stored in a repository for easy access. The user will need to authenticate with the service to be able to pull the image, and the service will check if the user has the required claims to access it.

Here is an example of a Pull pdfile:

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

To pull the image from the catalog, the user will need to authenticate with the service by passing either the `--username` and `--password` flags or by using the `--api-key` flag with a valid key. If the user has the required claims to access the image, the service will then pull the image from the remote storage and create a new virtual machine with the same software stack as the Golden Master image. Otherwise, it will return a **not found** error.

```prldevops
prldevops pull /path/to/pdfile --username=foo --password=bar
```

After the image is pulled, the user will have a new virtual machine with the same software stack as the Golden Master image, ready to be used. 

{% include notification.html message="It is important to note that this process may take some time, depending on the size of the virtual machine and the user's internet connection." status="is-warning" %}

## Conclusion

In this example, we have demonstrated how the Parallels Desktop DevOps Service can be used to create a Golden Master image. This image can then be shared with your team through the Parallels Desktop DevOps Service and its Remote Catalog feature. 

We have also highlighted how the Parallels Desktop DevOps Service can be utilized to retrieve the Golden Master image from the catalog and generate a new virtual machine with the same software stack as the Golden Master image.

This streamlined approach facilitates the secure sharing of Golden Master images with your team, ensuring that everyone has access to the same software stack.