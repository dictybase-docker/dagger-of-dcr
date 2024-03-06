# Dagger Golang Module
This module provides a set of functions to run Go language tests within a containerized environment.

## Install
```
dagger install github.com/dictybase-docker/dagger-of-dcr/golang@main
```
## Functions

- `test`: This function runs golang tests. It takes in a context, an optional Go version, a source directory to test, and an optional slice of strings representing additional arguments to the go test command.

- `lint`: This function runs golangci-lint on the Go source code. It takes in a context, an optional string specifying the version of golangci-lint to use, the source directory to test, and an optional slice of strings representing additional arguments to the golangci-lint command.

- `publish`: This function builds and pushes a Docker image to a Docker registry. It takes in a context, an optional source directory where the Docker context is located, an optional docker namespace, an optional path to the Dockerfile, the name of the image to be built, and the tag of the image to be built.

## Usage
To get help about for any function, for example `test`.

```
dagger -m "golang" call test --help
```

To run it 

```
dagger -m "golang" call test --src="."
```

Use other function arguments as necessary.

The `lint` and `publish` methods can be used in a similar way.



