[build-cli]: ./build.md
[build-docker]: ./docker.md
[architecture]: ./architecture.md
[develop]: ./develop.md
[generate-docs]: ../../generatedocs/README.md

# Working with docker

[build-cli] | [build-docker] | [develop] | [generate-docs] | [architecture]

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

## Running gopherciser from a docker container locally

```bash
docker run ghcr.io/qlik-oss/gopherciser/gopherciser:latest ./gopherciser -h
```
