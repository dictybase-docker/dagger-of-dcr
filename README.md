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
- [Running Dagger Functions](#running-dagger-functions)
  - [Example: Using Dagger CLI with Gitter Module](#example-using-dagger-cli-with-gitter-module)
  - [Example: Using Dagger CLI with Kops Module](#example-using-dagger-cli-with-kops-module)
  - [Example: Using Dagger CLI with Container Image Module](#example-using-dagger-cli-with-container-image-module)
  - [Example: Using Dagger CLI with Golang Module](#example-using-dagger-cli-with-golang-module)
  - [Example: Using Dagger CLI with PulumiOps Module](#example-using-dagger-cli-with-pulumiops-module)

## Overview

The project is divided into several modules, each with its own specific functionality:

1. **Gitter Module**: Provides a `Gitter` struct to manipulate Git repositories, including setting repository details and performing actions like checkout and inspect.
2. **Container Image Module**: Manages and builds Docker container images based on Git references. It includes methods to set various properties of the container image and generate appropriate Docker image tags.
3. **Golang Module**: Provides functions to run Go language tests, lint Go source code, and publish Docker images.
4. **Kops Module**: Provides functionality to set up and manage Kubernetes clusters using Kops and kubectl binaries.
5. **PulumiOps Module**: Provides functionality to manage Pulumi operations, including setting Pulumi version, backend, credentials, and Kubernetes configuration.

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

### PulumiOps Module

- **WithPulumi**: Sets the Pulumi version for the PulumiOps instance.
- **WithBackend**: Sets the backend for storing state for the PulumiOps instance.
- **WithCredentials**: Sets the credentials file for the PulumiOps instance.
- **WithKubeConfig**: Sets the Kubernetes configuration file for the PulumiOps instance.
- **PulumiContainer**: Returns a container with the specified Pulumi version.
- **Login**: Logs into the Pulumi backend using the provided credentials.
- **KubeAccess**: Sets up Kubernetes access using the provided kubeconfig file.
- **DeployBackend**: Deploys a backend application using Pulumi configurations and specified parameters.

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

### Example: Using Dagger CLI with Gitter Module

To set the Git reference and repository, and then check out the specified reference, you can use the following Dagger CLI command:

```shell
 dagger -m gitter call with-ref --ref=develop with-repository \
    --repository=https://github.com/dictybase-playground/gdrive-image-uploadr.git \
    checkout entries
```

### Example: Using Dagger CLI with Kops Module

To export the kubeconfig file for a specified Kops cluster, you can use the following Dagger CLI command:

```shell
dagger -m kops call with-cluster --cluster=my-cluster with-state-storage \
    --storage=s3://my-state-store with-credentials --credentials=/path/to/credentials.json \
    with-name --name=my-kubeconfig.yaml export-kubectl
```

### Example: Using Dagger CLI with Container Image Module

To publish a container image to Docker Hub, you can use the following Dagger CLI command:

```shell
dagger -m container-image call with-namespace --namespace=my-namespace with-ref --ref=main \
    with-repository --repository=my-repo with-dockerfile --dockerfile=./Dockerfile \
    with-image --image=my-image publish-from-repo --user=my-dockerhub-user --password=my-dockerhub-password
```

### Example: Using Dagger CLI with Golang Module

To run Go language tests within a containerized environment, you can use the following Dagger CLI command:

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

### Example: Using Dagger CLI with PulumiOps Module

To deploy a backend application using Pulumi configurations and specified parameters, you can use the following Dagger CLI command:

```shell
dagger -m pulumi-ops call deploy-backend --src=/path/to/source --project=backend_application \
    --app=my-app --tag=latest --stack=dev
```
