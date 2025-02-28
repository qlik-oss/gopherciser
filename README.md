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

## Building gopherciser

Documentation how to build the gopherciser can be found [here](./docs/dev/building.md).

## Gopherciser in a docker container

Documentation how to build docker images and run gopherciser from a docker container can be found [here](./docs/dev/docker.md).

## Gopherciser architecture

A description of the gopherciser architecture can be found [here](./docs/dev/architecture.md)

## Adding actions and features

Documentation how and where to develop additions to gopherciser can be found [here](./docs/dev/develop.md)