[build-cli]: ./build.md
[build-docker]: ./docker.md
[architecture]: ./architecture.md
[develop]: ./develop.md
[generate-docs]: ../../generatedocs/README.md

# Building Gopherciser

[build-cli] | [build-docker] | [develop] | [generate-docs] | [architecture]

## Prerequisites

Gopherciser requires a Golang 1.23 build environment or later.

## Build commands

### Full build

You can build Gopherciser locally using the `make` command:

`make build`

This produces Linux, Darwin and Microsoft Windows binaries in the build folder.

### Quick build

`make quickbuild`

This leaves the build folders as-is before building (that is, no cleaning is done etc.) and produces a build for the local operating system only. 

To do a cleaning before building, run `make clean`. 

### Docker build

Information how to work with docker images can be found at [build-docker].

### Building the documentation

The file `documentation.go`, used to generate wiki and tooltips in the GUI, can be generated using:

```bash
go generate
```

The wiki is updated on any push to master, however it can be generated locally:

```bash
make genwiki
```

For more information, see [generate-docs].

## Test commands

### Running normal tests

You can run tests using the `make` command:

`make test`

This runs all normal tests.

### Running all tests

`make alltests`

This runs all tests with verbose output and without relying on cache. This also creates coverage files in csv and html format.

## Linting commands

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