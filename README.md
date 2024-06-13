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

- **WithRef**: Sets the Git reference (branch, tag, or SHA) for the `Gitter` instance.
- **WithRepository**: Sets the git repository name for the `Gitter` instance.
- **Checkout**: Clones the repository and checks out the specified reference.

### Container Image Module

- **WithNamespace**: Sets the Docker namespace under which the image will be pushed.
- **WithRef**: Sets the Git reference (branch, tag, or SHA) for the `ContainerImage` instance.
- **WithRepository**: Sets the GitHub repository name for the `ContainerImage` instance.
- **WithDockerfile**: Sets the Dockerfile path for the `ContainerImage` instance.
- **WithImage**: Sets the image name for the `ContainerImage` instance.
- **PublishFromRepo**: Publishes a container image to Docker Hub.

### Golang Module

- **Test**: Runs Go language tests within a containerized environment.
- **Lint**: Runs golangci-lint on the Go source code.
- **Publish**: Builds and pushes a Docker image to a Docker registry.

### Kops Module

- **KopsContainer**: Sets up a container with specified versions of kubectl and kops binaries from given URLs.
- **ExportKubectl**: Exports the kubeconfig file for the specified Kops cluster to a specified output path.
- **WithName**: Sets the name of the kubectl output file.
- **WithCredentials**: Sets the credentials file.
- **WithStateStorage**: Sets the location of the state storage.
- **WithCluster**: Sets the cluster name.
- **WithKops**: Sets the Kops version.
- **WithKubectl**: Sets the kubectl version.

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
#### Using Dagger CLI with gitter Module
To checkout a branch of a public git repository and list all files

```shell
dagger -m gitter call with-ref --ref=develop with-repository \
    --repository=https://github.com/dictybase-playground/gdrive-image-uploadr.git \
    checkout entries
```


#### Using Dagger CLI with kops Module
To export the kubeconfig file for a specified kops cluster, you can use the
following Dagger CLI command:

```shell
dagger -m kops call with-cluster --cluster=my-cluster with-state-storage \
    --storage=gs://my-state-store with-credentials --credentials=/path/to/credentials.json \
    with-name export-kubectl --output=./my-kubeconfig.yaml
```
