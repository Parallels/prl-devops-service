---
layout: page
title: Rest API
subtitle: Authentication and Security
menubar: docs_menu
show_sidebar: false
toc: true
---

# Authentication and Security

The Parallels Desktop REST API uses a token-based authentication system. This system is designed to be simple and secure, and it is based on the OAuth 2.0 protocol. It then also uses a Role-Based Access Control (RBAC) system to control the access to the API, allowing to define which users can access which resources. Each token is issued for a specific user and has a limited lifetime.

This together with the use of HTTPS, ensures that the communication between the client and the server is secure.

We support the latest token signing algorithm, RS256, and the tokens are signed using a private key that is stored on the host. The public key is then used to verify the token signature.

You can also use the HMAC algorithm to sign the tokens, but we recommend using RS256 as it is more secure.

## How do I get a token?

To get a token, you need to send a POST request to the `/auth/token` endpoint with the following parameters:

- `username`: The username of the user.
- `password`: The password of the user.

for example:

```json
{
  "username": "admin",
  "password": "password"
}
```

this will return a token that you can use to authenticate your requests.

## How do I use the token?

Once you have the token, you need to include it in the `Authorization` header of your requests, for example:

```http
GET /api/v1/machines HTTP/1.1
Host: localhost
Authorization: Bearer <token>
```

## How do I validate the tokens?

You can validate the tokens by sending a GET request to the `/auth/token/validate` endpoint with the token in the `Authorization` header, for example:

```http
GET /api/v1/auth/token/validate HTTP/1.1
Host: localhost
Body: {
    "token": "<token>"
}
```

this will return a response with the status of the token.
example:

```json
{
    "message": "Token is expired",
    "code": 401
}
```

## Database Encryption

The DevOps service uses a database to store the tokens and the users, this is a file based system, and we encrypt the database to ensure that the data is secure. We use the AES-256 encryption algorithm to encrypt the database.
<p>
{% include notification.html message="By default the database will not be encrypted and you will need to setup a password for it to be, you can find how to setup the password in the [configuration guide](/prl-devops-service/docs/getting-started/configuration/#password-complexity)" status="is-warning" %}
</p>
<p>
{% include notification.html message="If you change the private key, you will not be able to decrypt
the database and you will lose all the data." status="is-danger" %}
</p>

### Generate a new private key

To generate a new private key, you can use the following command:

```powershell
devops gen-rsa --file=private.pem
```


## Password Brute Force Protection

The DevOps REST API also includes a password brute force protection system. This system is designed to protect the API from brute force attacks by limiting the number of failed login attempts. If the number of failed login attempts exceeds the limit, the user will be locked out for a specific period of time. We also provide a increasing delay between the login attempts to make it harder for the attackers to guess the password.

This can also be configured in the configuration file of the DevOps service, please check the [configuration guide]({{ site.url }}{{ site.baseurl }}/docs/getting-started/configuration/#brute-force-protection) for more information.

## Password Complexity

We implemented a password complexity system to ensure that the passwords are strong and secure. This system enforces the use of strong passwords by requiring a minimum length, and a combination of uppercase, lowercase, numbers, and special characters.

We can also configure the password complexity in the configuration file of the DevOps service, please check the [configuration guide]({{ site.url }}{{ site.baseurl }}/docs/getting-started/configuration/#password-complexity) for more information.

## How do we store user passwords?

We store the user passwords using the bcrypt algorithm, this is a secure algorithm that is designed to be slow and hard to crack. This ensures that the passwords are secure and that they cannot be easily cracked.
While the bcrypt algorithm is secure,it also can be a bit slow, for this reason we also provide other ways like the common SHA-256 algorithm, but we recommend using bcrypt.