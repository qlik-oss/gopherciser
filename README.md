# Gopherciser

Gopherciser is used for load testing (that is, stress testing and performance measurement) in Qlik Sense® Enterprise deployments. It is based on [enigma-go](https://github.com/qlik-oss/enigma-go), which is a library for communication with the Qlik® Associative Engine. 

For an introduction to Gopherciser, see [Load testing - an introduction](./docs/README.md).

For information on how to set up load scenarios, see [Setting up load scenarios](./docs/settingup.md).

For information on the code structure, see [Architecture](./architecture.md).

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
see [Generating Gopherciser documentation](./generatedocs/README.md)

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
