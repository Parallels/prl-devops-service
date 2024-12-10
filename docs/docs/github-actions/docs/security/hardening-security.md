---
layout: page
title: Security
subtitle: Harden Security
menubar: docs_menu
show_sidebar: false
toc: true
---

# Hardening Security

## Setting up database encryption

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

## Setting up TLS

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

### Setting up the JWT signing method

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

## Changing the Root password

Typically, the service is initiated with a randomly generated root password. However, in order to access the service's REST API, you will need to update this password. To do so, we provide a simple command line. For instance, we will set the password as `VeryStr0ngP@ssw0rd`, but you may choose any password you prefer.

```prldevops
prldevops update-root-password --password=VeryStr0ngP@ssw0rd
```