# dagger of dcr project

This project provides a set of tools and functions to manage Git repositories,
run Go language tests, and build and publish Docker container images. It is
designed to work within a containerized environment, leveraging the Dagger
framework for seamless integration and execution.

## Table of Contents

- [Overview](#overview)
- [Modules and Functions](#modules-and-functions)
  - [Gitter Module](#gitter-module)
  - [Container Image Module](#container-image-module)
  - [Golang Module](#golang-module)
  - [Kops Module](#kops-module)
  - [PulumiOps Module](#pulumiops-module)
- [Usage](#usage)
- [Examples of running Dagger functions using CLI](#running-dagger-functions)
  - [Gitter Module](#gitter)
  - [Kops Module](#kops)
  - [Container Image Module](#container-image)
  - [Golang Module](#golang)
  - [PulumiOps Module](#pulumiops)

## Overview

The project is divided into several modules, each with its own specific functionality:

1. **Gitter Module**: Provides a `Gitter` struct to manipulate Git repositories, including setting repository details and performing actions like checkout and inspect.
2. **Container Image Module**: Manages and builds Docker container images based on Git references. It includes methods to set various properties of the container image and generate appropriate Docker image tags.
3. **Golang Module**: Provides functions to run Go language tests, lint Go source code, and publish Docker images.
4. **Kops Module**: Provides functionality to set up and manage Kubernetes clusters using Kops and kubectl binaries.
5. **PulumiOps Module**: Provides functionality to manage Pulumi operations, including setting Pulumi version, backend, credentials, and Kubernetes configuration.

## Modules and Functions

### Gitter Module
- `Checkout`: Clones the repository and checks out the specified reference.

### Container Image Module
- `PublishFromRepo`: Publishes a container image to Docker Hub.

### Golang Module
- `Test`: Runs Go language tests within a containerized environment.
- `Lint`: Runs golangci-lint on the Go source code.
- `Publish`: Builds and pushes a Docker image to a Docker registry.

### Kops Module
- `ExportKubectl`: Exports the kubeconfig file for the specified Kops cluster to a specified output path.

### PulumiOps Module
- `DeployApp`: Deploys a backend application using Pulumi configurations and specified parameters.

## Usage

## Running Dagger Functions

To run the Dagger functions using the Dagger command line, follow these steps:

1. **Install Dagger CLI**: Ensure you have the Dagger command line interface
   installed. You can download it from the official Dagger
   [website](https://dagger.io).
2. **Initialize Dagger**: Run `dagger init` in your project directory to
   initialize Dagger.
3. **Run Functions**: Use the `dagger call` command followed by the function
   name to execute the desired function. For example:

### Examples
#### Gitter

To set the Git reference and repository, and then check out the specified
reference, you can use the following Dagger CLI command:

```shell
 dagger -m gitter call with-ref --ref=develop with-repository \
    --repository=https://github.com/dictybase-playground/gdrive-image-uploadr.git \
    checkout entries
```

#### Kops

To export the kubeconfig file for a specified Kops cluster, you can use the
following Dagger CLI command:

```shell
dagger -m kops call with-cluster --cluster=my-cluster with-state-storage \
    --storage=s3://my-state-store with-credentials --credentials=/path/to/credentials.json \
    export-kubectl --output=./mykube.yaml
```

#### Container Image

To publish a container image to Docker Hub, you can use the following Dagger CLI
command:

```shell
dagger -m container-image call with-namespace --namespace=my-namespace with-ref --ref=main \
    with-repository --repository=my-repo with-dockerfile --dockerfile=./Dockerfile \
    with-image --image=my-image publish-from-repo --user=my-dockerhub-user --password=my-dockerhub-password
```

#### Golang

To run Go language tests within a containerized environment, you can use the
following Dagger CLI command:

```shell
dagger -m golang call test --version=go-1.21 --src=/path/to/source --args="-v ./..."
```

To run golangci-lint on the Go source code, you can use the following Dagger CLI command:

```shell
dagger -m golang call lint --version=v1.55.2-alpine --src=/path/to/source --args="run ./..."
```

To build and push a Docker image to a Docker registry, you can use the following Dagger CLI command:

```shell
dagger -m golang call publish --src=. --namespace=my-namespace --dockerfile=./Dockerfile \
    --image=my-image --imageTag=latest
```

#### PulumiOps

To deploy a backend application using Pulumi configurations and specified
parameters, you can use the following Dagger CLI command:

```shell
dagger -m pulumi-ops call deploy-app --src=/path/to/source --project=backend_application \
    --app=my-app --tag=latest --stack=dev
```

