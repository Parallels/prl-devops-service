## Contributing

[fork]: https://github.com/Parallels/pd-api-service/fork
[pr]: https://github.com/Parallels/pd-api-service/compare
[code-of-conduct]: CODE_OF_CONDUCT.md

Hi there! We're thrilled that you'd like to contribute to this project. Your help is essential for keeping it great.

We accept pull requests for bug fixes and features where we've discussed the approach in an issue and given the go-ahead for a community member to work on it. We'd also love to hear about ideas for new features as issues.

We track issues on our project board [here](https://github.com/orgs/Parallels/projects/6/views/1).

Please do:

* Check existing issues to verify that the [bug][bug issues] or [feature request][feature request issues] has not already been submitted.
* Open an issue if things aren't working as expected.
* Open an issue to propose a significant change.
* Open a pull request to fix a bug.
* Open a pull request to fix documentation about a command.

Please avoid:

* Opening pull requests for issues marked `needs-design`, `needs-investigation`, or `blocked`.

Contributions to this project are released to the public under the [project's open source license](LICENSE).

Please note that this project is released with a [Contributor Code of Conduct][code-of-conduct]. By participating in this project you agree to abide by its terms.

## Prerequisites for running and testing code

These are one time installations required to be able to test your changes locally as part of the pull request (PR) submission process.

1. Install [Go](https://golang.org/dl/) for Mac
2. Install [VS Code](https://code.visualstudio.com/download) for Mac
3. Install [Parallels Desktop](https://www.parallels.com/uk/products/desktop/pro/) for Mac
4. Install [Hashicorp Packer](https://www.packer.io/downloads) for Mac
5. Install [Vagrant](https://www.vagrantup.com/downloads.html) for Mac
6. Install the dependencies. From the repository root run:

```bash
cd src
go mod download
go mod tidy
```

## Dev loop & Testing changes

The API is written in go to start it in debug mode you can go to Visual Studio Code `Run and Debug` and select the Parallels Desktop Api. This will start the service in debug mode and you can then use the swagger documentation to test the endpoints.

**Attention** There are some environment variables that might need to be set so you will need to create a .env file in the root of the repository and add the following variables:

```bash
ROOT_PASSWORD=<root_password>
```

### Build

Build changes (one time):

```bash
go build ./src
```

### Running tests

```bash
go test ./src/...
```

## Submitting a pull request

1. [Fork][fork] and clone the repository
1. Configure and install the dependencies (in the repository src folder): `go mod download`
1. Create a new branch: `git checkout -b my-branch-name`
1. Make your change, add tests, and make sure the tests and linter still pass
1. Push to your fork and [submit a pull request][pr]

Here are a few things you can do that will increase the likelihood of your pull request being accepted:

- Format your code with [prettier](https://prettier.io/).
- Write tests.
- Keep your change as focused as possible. If there are multiple changes you would like to make that are not dependent upon each other, consider submitting them as separate pull requests.
- Write a [good commit message](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).

## Resources

- [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
- [Using Pull Requests](https://help.github.com/articles/about-pull-requests/)
- [GitHub Help](https://help.github.com)

[bug issues]: https://github.com/Parallels/pd-api-service/labels/bug
[feature request issues]: https://github.com/Parallels/pd-api-service/labels/enhancement