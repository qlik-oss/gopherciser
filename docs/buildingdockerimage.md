# Building a Docker image

To create a Docker image from this package, run the following Docker command from within the unzipped Gopherciser package:
```bash
docker build . -t qlik/gopherciser:$(cat version)
```

The Prometheus metrics port can be overridden with the `PORT` argument:

```bash
docker build . -t qlik/gopherciser:$(cat version) --build-arg PORT=9191
```