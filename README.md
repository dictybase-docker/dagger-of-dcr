# dagger of dcr project

This project provides a set of tools and functions to manage Git repositories,
run Go language tests, and build and publish Docker container images. It is
designed to work within a containerized environment, leveraging the Dagger
framework for seamless integration and execution.

## Overview

The project is divided into several modules, each with its own specific functionality:

1. **Gitter Module**: Provides a `Gitter` struct to manipulate Git repositories, including setting repository details and performing actions like checkout and inspect.
2. **Container Image Module**: Manages and builds Docker container images based on Git references. It includes methods to set various properties of the container image and generate appropriate Docker image tags.
3. **Golang Module**: Provides functions to run Go language tests, lint Go source code, and publish Docker images.
4. **Kops Module**: Provides functionality to set up and manage Kubernetes clusters using Kops and kubectl binaries.

## Modules and Functions

### Gitter Module
- **Checkout**: Clones the repository and checks out the specified reference.

### Container Image Module
- **PublishFromRepo**: Publishes a container image to Docker Hub.

### Golang Module
- **Test**: Runs Go language tests within a containerized environment.
- **Lint**: Runs golangci-lint on the Go source code.
- **Publish**: Builds and pushes a Docker image to a Docker registry.

### Kops Module
- **ExportKubectl**: Exports the kubeconfig file for the specified Kops cluster to a specified output path.

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

```shell
dagger -m gitter call with-ref --ref=develop with-repository \
    --repository=https://github.com/dictybase-playground/gdrive-image-uploadr.git \
    checkout entries
```

