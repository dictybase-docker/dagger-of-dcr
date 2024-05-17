# Dagger Golang Project

This project provides a set of tools and functions to manage Git repositories, run Go language tests, and build and publish Docker container images. It is designed to work within a containerized environment, leveraging the Dagger framework for seamless integration and execution.

## Overview

The project is divided into several modules, each with its own specific functionality:

1. **Gitter Module**: Provides a `Gitter` struct to manipulate Git repositories, including setting repository details and performing actions like checkout and inspect.
2. **Container Image Module**: Manages and builds Docker container images based on Git references. It includes methods to set various properties of the container image and generate appropriate Docker image tags.
3. **Golang Module**: Provides functions to run Go language tests, lint Go source code, and publish Docker images.

## Modules and Functions

### Gitter Module

- **WithRef**: Sets the Git reference (branch, tag, or SHA) for the `Gitter` instance.
- **WithRepository**: Sets the GitHub repository name for the `Gitter` instance.
- **Checkout**: Clones the repository and checks out the specified reference.
- **CommitHash**: Retrieves the short commit hash of the HEAD from the specified Git repository.
- **Inspect**: Clones the given repository and returns a Terminal instance for inspection.
- **ParseRef**: Extracts the branch name from a Git reference string or returns the original reference if no match is found.

### Container Image Module

- **WithNamespace**: Sets the Docker namespace under which the image will be pushed.
- **WithRef**: Sets the Git reference (branch, tag, or SHA) for the `ContainerImage` instance.
- **WithRepository**: Sets the GitHub repository name for the `ContainerImage` instance.
- **WithDockerfile**: Sets the Dockerfile path for the `ContainerImage` instance.
- **WithImage**: Sets the image name for the `ContainerImage` instance.
- **PublishFromRepo**: Publishes a container image to Docker Hub.
- **FakePublishFromRepo**: Publishes a container image to a temporary repository with a time-to-live of 10 minutes.
- **ImageTag**: Generates a Docker image tag based on the provided Git reference.

### Golang Module

- **Test**: Runs Go language tests within a containerized environment.
- **Lint**: Runs golangci-lint on the Go source code.
- **Publish**: Builds and pushes a Docker image to a Docker registry.

## Usage

To use the functions provided by this project, create instances of the respective structs (`Gitter`, `ContainerImage`, `Golang`) and call the desired methods. For example, to checkout a Git repository:

```
