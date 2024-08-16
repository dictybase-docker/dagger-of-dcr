// Package main provides a set of functions to run Go language tests within a containerized environment.
package main

import (
	"context"
	"fmt"
	"os"

	F "github.com/IBM/fp-go/function"
)

const (
	PROJ_MOUNT = "/app"
	WOLFI_BASE = "cgr.dev/chainguard/wolfi-base"
	LINT_BASE  = "golangci/golangci-lint"
)

type Golang struct {
	ArangoPassword string
	ArangoVersion  string
	ArangoPort     int
	GolangVersion  string
}

// Test runs Go tests in a containerized environment.
// It sets up a Wolfi-based container with the specified Go version,
// prepares the workspace, and executes the tests.
//
// Parameters:
//   - ctx: The context for the operation.
//   - src: The source directory to test (required).
//   - args: Optional slice of strings representing additional arguments to the go test command.
//
// Returns:
//   - A string containing the stdout of the test execution.
//   - An error if any step in the process fails.
func (gom *Golang) Test(
	ctx context.Context,
	// The source directory to test, Required.
	src *Directory,
	// An optional slice of strings representing additional arguments to the go test command
	// +optional
	args []string,
) (string, error) {
	return F.Pipe5(
		dag.Container(),
		base(WOLFI_BASE),
		wolfiWithGoInstall(gom.GolangVersion),
		prepareWorkspace(src, PROJ_MOUNT),
		modCache,
		goTestRunner(args),
	).Stdout(ctx)
}

// Lint runs golangci-lint on the Go source code in a containerized environment.
// It uses a specified version of golangci-lint to perform static code analysis.
//
// Parameters:
//   - ctx: The context for the operation.
//   - version: Optional string specifying the version of golangci-lint to use (default: "v1.55.2-alpine").
//   - src: The source directory to lint (required).
//   - args: Optional slice of strings representing additional arguments to the golangci-lint command.
//
// Returns:
//   - A string containing the stdout of the linting process.
//   - An error if any step in the process fails.
func (gom *Golang) Lint(
	ctx context.Context,
	// An optional string specifying the version of golangci-lint to use
	// +optional
	// +default="v1.55.2-alpine"
	version string,
	// The source directory to test, Required.
	src *Directory,
	// An optional slice of strings representing additional arguments to the
	// golangci-lint command
	// +optional
	args []string,
) (string, error) {
	return F.Pipe3(
		dag.Container(),
		base(LINT_BASE),
		prepareWorkspace(src, PROJ_MOUNT),
		goLintRunner(args),
	).Stdout(ctx)
}

// Publish builds and pushes a Docker image to a Docker registry.
// It uses the provided Docker context and Dockerfile to build the image,
// then pushes it to the specified Docker registry.
//
// Parameters:
//   - ctx: The context for the operation.
//   - src: The source directory where the Docker context is located (optional, default: ".").
//   - namespace: The Docker namespace under which the image will be pushed (optional, default: "dictybase").
//   - dockerFile: The path to the Dockerfile (optional, default: "build/package/Dockerfile").
//   - image: The name of the image to be built (required).
//   - imageTag: The tag of the image to be built (required).
//
// Returns:
//   - A string containing the result of the publish operation.
//   - An error if any step in the process fails.
func (gom *Golang) Publish(
	ctx context.Context,
	// specifies the source directory where the Docker context is located
	// +optional
	// +default="."
	src string,
	// the docker namespace under which the image will be pushed
	// +optional
	// +default="dictybase"
	namespace string,
	// specifies the path to the Dockerfile
	// +optional
	// +default="build/package/Dockerfile"
	dockerFile string,
	// name of the image to be built, Required
	image string,
	// tag of the image to be built, Required
	imageTag string,
) (string, error) {
	var empty string
	userValue, err := fetchAndValidateEnvVars("DOCKERHUB_USER")
	if err != nil {
		return empty, err
	}
	passValue, err := fetchAndValidateEnvVars("DOCKER_PASS")
	if err != nil {
		return empty, nil
	}
	return F.Pipe2(
		dag.Container(),
		setupBuild(src, dockerFile),
		dockerHubAuth(userValue, dag.SetSecret("docker-pass", passValue)),
	).Publish(ctx, fmt.Sprintf("%s/%s:%s", namespace, image, imageTag))
}

func fetchAndValidateEnvVars(envVar string) (string, error) {
	value := os.Getenv(envVar)
	if len(value) == 0 {
		return "", fmt.Errorf("value of %s env variable is not set", envVar)
	}
	return value, nil
}

// TestsWithArangoDB runs Go tests in a containerized environment with an ArangoDB service.
// It sets up an ArangoDB service, prepares a Wolfi-based container with the specified Go version,
// and executes the tests with the ArangoDB service available.
//
// Parameters:
//   - ctx: The context for the operation.
//   - src: The source directory to test (required).
//   - args: Optional slice of strings representing additional arguments to the go test command.
//
// Returns:
//   - A string containing the stdout of the test execution.
//   - An error if any step in the process fails.
func (gom *Golang) TestsWithArangoDB(
	ctx context.Context,
	// The source directory to test, Required.
	src *Directory,
	// An optional slice of strings representing additional arguments to the go test command
	// +optional
	args []string,
) (string, error) {
	arangoService := dag.Container().
		From(gom.ArangoVersion).
		WithEnvVariable("ARANGO_ROOT_PASSWORD", gom.ArangoPassword).
		WithExposedPort(gom.ArangoPort).
		AsService()

	arangoHost, err := arangoService.Hostname(ctx)
	if err != nil {
		return arangoHost, fmt.Errorf(
			"error in retrieving arangodb host %w",
			err,
		)
	}

	return dag.Container().From(WOLFI_BASE).
		WithServiceBinding("arango", arangoService).
		WithEnvVariable("ARANGO_HOST", arangoHost).
		WithEnvVariable("ARANGO_PASS", gom.ArangoPassword).
		WithEnvVariable("ARANGO_USER", "root").
		WithExec([]string{"apk", "update"}).
		WithExec([]string{"apk", "add", gom.GolangVersion}).
		WithMountedDirectory(PROJ_MOUNT, src).
		WithWorkdir(PROJ_MOUNT).
		WithExec([]string{"go", "mod", "download"}).
		WithExec(append([]string{"go", "test", "-v", "./..."}, args...)).
		Stdout(ctx)
}

// WithArangoPassword sets the root password for the ArangoDB instance.
//
// Parameter:
//   - password: The root password for the ArangoDB instance (optional, default: "golam").
//
// Returns:
//   - A pointer to the modified Golang struct.
func (gom *Golang) WithArangoPassword(
	// The root password for the ArangoDB instance
	// +optional
	// +default="golam"
	password string,
) *Golang {
	gom.ArangoPassword = password
	return gom
}

// WithArangoVersion sets the version of ArangoDB to use.
//
// Parameter:
//   - version: The version of ArangoDB to use (optional, default: "3.10.9").
//
// Returns:
//   - A pointer to the modified Golang struct.
func (gom *Golang) WithArangoVersion(
	// The version of ArangoDB to use
	// +optional
	// +default="3.10.9"
	version string,
) *Golang {
	gom.ArangoVersion = version
	return gom
}

// WithArangoPort sets the port to expose ArangoDB on.
//
// Parameter:
//   - port: The port to expose ArangoDB on (optional, default: 8529).
//
// Returns:
//   - A pointer to the modified Golang struct.
func (gom *Golang) WithArangoPort(
	// The port to expose ArangoDB on
	// +optional
	// +default=8529
	port int,
) *Golang {
	gom.ArangoPort = port
	return gom
}

// WithGolangVersion sets the version of Golang to use.
//
// Parameter:
//   - version: The version of Golang to use (optional, default: "go-1.21").
//
// Returns:
//   - A pointer to the modified Golang struct.
func (gom *Golang) WithGolangVersion(
	// The version of Golang to use
	// +optional
	// +default="go-1.21"
	version string,
) *Golang {
	gom.GolangVersion = version
	return gom
}
