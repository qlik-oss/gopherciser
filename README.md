# Gopherciser

[![CircleCI](https://circleci.com/gh/qlik-oss/gopherciser.svg?style=svg)](https://circleci.com/gh/qlik-oss/gopherciser)

![Gopherciser logo](docs/images/logo.png)

Gopherciser is used for load testing (that is, stress testing and performance measurement) in Qlik Sense® Enterprise deployments. It is based on [enigma-go](https://github.com/qlik-oss/enigma-go), which is a library for communication with the Qlik® Associative Engine. 

Gopherciser can run standalone, but is also included in the Qlik Sense Enterprise Scalability Tools (QSEST), which are available for download [here](https://community.qlik.com/t5/Qlik-Scalability/Qlik-Sense-Enterprise-Scalability-Tools/gpm-p/1579916).

For more information on how to perform load testing with Gopherciser see the [wiki](https://github.com/qlik-oss/gopherciser/wiki/introduction), this readme documents building and development of gopherciser.

## Cloning repo

This repo contains the wiki as a submodule, to clone sub modules when cloning the project 

```bash
git clone --recurse-submodules git@github.com:qlik-oss/gopherciser.git
```

If repo was cloned manually, the wiki submodule can be checked out using

```bash
git submodule update --init --recursive
```

Updating submodule to version defined by current branch commit:

```bash
git submodule update
```

**Note**  the submodule will by default be in it's `master` branch. Any changes done and pushed in the submodule master branch will instantly update the wiki (i.e. don't make changes intended for a PR directly here).

## Building Gopherciser

### Prerequisites

#### Golang build environment

Gopherciser requires a Golang 1.23 build environment or later.

#### Building the documentation

The file `documentation.go`, used to generate wiki and tooltips in the GUI, can be generated using:

```bash
go generate
```

The wiki is updated on any push to master, however it can be generated locally:

```bash
make genwiki
```

For more information, see [Generating Gopherciser documentation](./generatedocs/README.md).

### Build commands

#### Full build

You can build Gopherciser locally using the `make` command:

`make build`

This produces Linux, Darwin and Microsoft Windows binaries in the build folder.

#### Quick build

`make quickbuild`

This leaves the build folders as-is before building (that is, no cleaning is done etc.) and produces a build for the local operating system only. 

To do a cleaning before building, run `make clean`. 

### Test commands

#### Running normal tests

You can run tests using the `make` command:

`make test`

This runs all normal tests.

#### Running all tests

`make alltests`

This runs all tests with verbose output and without relying on cache.

### Linting commands

You can run linter using 

`make lint` or `make lint-min` lint-min will run minimal lint for PR to be accepted.

### Verify command

`make verify` will run `quickbuild`, `test` and `lint-min`, this is a good command to run before pushing a commit to make sure CI will pass green.

## Updating Gopherciser dependencies

Do the following:

1. Update the modules:
   * `go get -u`: Update the modules to the  most recent minor or patch version.
   * `go get -u=patch`: Update the modules to the latest patch for minor version.
   * `go get github.com/some/lib@v1.2.3`: Get a particular version.
2. Run `go mod tidy` to remove any unused modules.
3. Run `go mod verify` to add packages needed for test packages etc.

## Pulling the Docker image

A Docker login is needed before the images can be pulled. 

### Create a token
1. Create a new token with the scope `read:packages` [here](https://github.com/settings/tokens).
2. Save your token to, for example, a file (or use an environment variable or similar).

### Log in

Log in with Docker to `ghcr.io`.

Using a token stored in the file github.token: 

```bash
docker login -u yourgithubusername --password=$(cat github.token) ghcr.io
```

Using the token in the environmental variable GITHUB_TOKEN:

```bash
docker login -u yourgithubusername --password=$GITHUB_TOKEN ghcr.io
```

### Pull docker image

The latest master version:

```bash
docker pull ghcr.io/qlik-oss/gopherciser/gopherciser:latest
```

Specific released version:

```bash
docker pull ghcr.io/qlik-oss/gopherciser/gopherciser:0.21.1
```

## Building a local Docker image

To create a Docker image locally, run the following make command:
```bash
make build-docker
```

## VSCode snippets for gopherciser development in VSCode

Documentation how to use snippets to help with development of gopherciser actions when using VSCodium or VSCode can be found [here](./docs/vscode/Readme.md).
