## Introduction
This is a project related to a tutorial on how to enable Kubernetes probes.

## Prerequisites
- Kubernetes cluster
- kubectl
- Docker
- Container Registry (Docker Hub, Azure Container Registry, etc.)

## Link to the tutorial


## Compile the go application
```bash
go build -o main main.go
```

## init the go module
This will create a go.mod file which will be used to manage the dependencies of the application:

```bash
go mod init k8s-probes
```

## tidy the go module
This will download the dependencies in the go.mod file
```bash
go mod tidy
```

## How to search for Go images in terminal
```bash
docker search golang
```


## Start The Application
```bash
go run main.go
```


## build the docker image
```bash
docker build -t supera/k8s-probes .
```


## If you build on Mac Apple Silicon

The following command will create a new builder instance and set it as the current builder instance:

```bash
docker buildx create --use
docker buildx inspect --bootstrap
```

### Build
The following command will build the image for both the amd64 and arm64 platforms and push it to the Docker Hub registry:

```bash
docker buildx build --platform linux/amd64,linux/arm64 -t supera/k8s-probes --push .
```

if you just want to buiild for linux/amd64, you can use the following command:

```bash
docker buildx build --platform linux/amd64 -t supera/k8s-probes --push .
```

--push flag is used to push the image to the registry.
