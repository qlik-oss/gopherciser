# Gopherciser

[![CircleCI](https://circleci.com/gh/qlik-oss/gopherciser.svg?style=svg)](https://circleci.com/gh/qlik-oss/gopherciser)

![Gopherciser logo](docs/images/logo.png)

Gopherciser is used for load testing (that is, stress testing and performance measurement) in Qlik Sense® Enterprise deployments. It is based on [enigma-go](https://github.com/qlik-oss/enigma-go), which is a library for communication with the Qlik® Associative Engine. 

Gopherciser can run standalone, but is also included in the Qlik Sense Enterprise Scalability Tools (QSEST), which are available for download [here](https://community.qlik.com/t5/Qlik-Scalability/Qlik-Sense-Enterprise-Scalability-Tools/gpm-p/1579916).

For more information on how to perform load testing with Gopherciser see the [wiki](https://github.com/qlik-oss/gopherciser/wiki/introduction), this readme documents building and development of gopherciser.

## Cloning repo

This repo contains the wiki as a submodule, to clone sub modules when cloning the project 

```bash
git clone --recurse-submodules git@github.com:qlik-oss/gopherciser.wiki.git
```

If repo was cloned manually, the wiki submodule can be checkd out using

```bash
git submodule update --init --recursive
```

**Note**  the submodule will by default be in it's `master` branch. Any changes done and pushed in the submodule master branch will instantly update the wiki (i.e. don't make changes intended for a PR directly here).

## Building Gopherciser

### Prerequisites

#### Golang build environment

Gopherciser requires a Golang 1.13 build environment or later.

#### Installing tools

**Note:** Since Gopherciser uses Go modules, do not install tools using the `go get` command while inside the Gopherciser repository. 

To install tools, use the `cd` command to leave the Gopherciser repository directory and then use `go get`.

#### Windows-specific prerequisites

If you use Git Bash, but do not have `make.exe` installed, do the following to install it: 

1. Go to [ezwinports](https://sourceforge.net/projects/ezwinports/).
2. Download `make-4.x-y-without-guile-w32-bin.zip` (make sure to get the version without guile).
3. Extract the ZIP file.
4. Copy the contents to the `Git\mingw64\` directory (the default location of mingw64 is `C:\Program Files\Git\mingw64`), but **do not** overwrite or replace any existing files.

#### Building the documentation

The documentation can be generated from json with:

```bash
go generate
```

To generate wiki run

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

## Updating Gopherciser dependencies

Do the following:

1. Update the modules:
   * `go get -u`: Update the modules to the  most recent minor or patch version.
   * `go get -u=patch`: Update the modules to the latest patch for minor version.
   * `go get github.com/some/lib@v1.2.3`: Get a particular version.
2. Run `go mod tidy` to remove any unused modules.
3. Run `go mod verify` to add packages needed for test packages etc.

## Pulling the Docker image

Unfortunately, the GitHub packages Docker repo is not very "public" (see this [community thread](https://github.community/t5/GitHub-Actions/docker-pull-from-public-GitHub-Package-Registry-fail-with-quot/td-p/32782)). This means a Docker login is needed before the images can be pulled. 

Do the following:

1. Create a new token with the scope `read:packages` [here](https://github.com/settings/tokens).
2. Save your token to, for example, a file (or use an environment variable or similar).
3. Log in with Docker to `docker.pkg.github.com`.

Using a token stored in the file github.token: 

```bash
docker login -u yourgithubusername --password=$(cat github.token) docker.pkg.github.com
```

Using the token in the environmental variable GITHUB_TOKEN:

```bash
docker login -u yourgithubusername --password=$GITHUB_TOKEN docker.pkg.github.com
```

4. Pull the Docker image.

The latest master version:

```bash
docker pull docker.pkg.github.com/qlik-oss/gopherciser/gopherciser:latest
```

Specific released version:

```bash
docker pull docker.pkg.github.com/qlik-oss/gopherciser/gopherciser:0.4.10
```

### Using the image in Kubernetes

To use the image in Kubernetes (for example, to perform executions as part of a Kubernetes job), credentials for the GitHub package registry need to be added the same way a private registry is used, see documentation [here](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).

## VSCode snippets for gopherciser development in VSCode

In the `docs/vscode` folder there is a file called `gopherciser.code-snippets` containing snippets which can be used with VSCode.

### Installing the the snippets

Snippets can be "installed" 2 diffrent ways

1. Copy the file `gopherciser.code-snippets` from the `docs/vscode` folder into the folder `.vscode` in the repo. This works on any OS.
2. On *nix systems it's recommended to create a symbolic link instead of copying the file to have it automatically be kept up to date. In the main repo folder create the symbolic link with the command `ln -s ../docs/vscode/gopherciser.code-snippets .vscode/gopherciser.code-snippets`.

### Using the template

Start writing the name of the snippet to and press enter.

`action`: Adds skeleton of a scenario action. This should be used the following way:

1. Create an empty file in `scenario` folder with the name of the new action, e.g. `dummy.go` for the action `dummy`.
2. Start writing `action` and press enter.
3. Change the action struct name if necessary, then press *tab* to write a description.