# Dagger Golang Module

This module provides a set of functions to run Go language tests within a containerized environment.

## Functions

- `Test`: This function runs golang tests. It takes in a context, an optional Go version, a source directory to test, and an optional slice of strings representing additional arguments to the go test command.

- `Lint`: This function runs golangci-lint on the Go source code. It takes in a context, an optional string specifying the version of golangci-lint to use, the source directory to test, and an optional slice of strings representing additional arguments to the golangci-lint command.

- `Publish`: This function builds and pushes a Docker image to a Docker registry. It takes in a context, an optional source directory where the Docker context is located, an optional docker namespace, an optional path to the Dockerfile, the name of the image to be built, and the tag of the image to be built.

## Usage

To use this module, you need to import it in your Go project and create an instance of the `Golang` struct. Then, you can call the `Test`, `Lint`, and `Publish` methods on this instance.

Here is an example of how to use the `Test` method:

```go
package main

import (
	"context"
	"fmt"

	dagger "path/to/dagger/module"
)

func main() {
	gom := &dagger.Golang{}
	ctx := context.Background()
	src := &dagger.Directory{Path: "./src"}

	result, err := gom.Test(ctx, "go-1.21", src, []string{"-v"})
	if err != nil {
		fmt.Println("Error running tests:", err)
		return
	}

	fmt.Println("Test result:", result)
}
```

In this example, we create a context and a `Directory` struct representing the source directory to test. We then call the `Test` method with these values, along with the Go version and additional arguments for the `go test` command. The `Test` method returns the test result and any error that occurred.

The `Lint` and `Publish` methods can be used in a similar way.
