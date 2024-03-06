# Dagger Golang Module

This module provides a set of functions to run Go language tests within a containerized environment.

## Functions

- `Test`: This function runs golang tests. It takes in a context, an optional Go version, a source directory to test, and an optional slice of strings representing additional arguments to the go test command.

- `Lint`: This function runs golangci-lint on the Go source code. It takes in a context, an optional string specifying the version of golangci-lint to use, the source directory to test, and an optional slice of strings representing additional arguments to the golangci-lint command.

- `Publish`: This function builds and pushes a Docker image to a Docker registry. It takes in a context, an optional source directory where the Docker context is located, an optional docker namespace, an optional path to the Dockerfile, the name of the image to be built, and the tag of the image to be built.

## Utility Functions

The `functional_utils.go` file contains utility functions that are used by the main functions. These include functions for setting up the base container, installing Go, preparing the workspace, running Go tests, running Go linter, setting up Docker build, and authenticating with Docker registry.
