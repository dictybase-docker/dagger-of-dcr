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

type Golang struct{}

// Runs golang tests
func (gom *Golang) Test(
	ctx context.Context,
	// An optional Go version to use for testing
	// +optional
	// +default="go-1.21"
	version string,
	// The source directory to test, Required.
	src *Directory,
	// An optional slice of strings representing additional arguments to the go test command
	// +optional
	args []string,
) (string, error) {
	return F.Pipe5(
		dag.Container(),
		base(WOLFI_BASE),
		wolfiWithGoInstall(version),
		prepareWorkspace(src, PROJ_MOUNT),
		modCache,
		goTestRunner(args),
	).Stdout(ctx)
}

// Lint runs golangci-lint on the Go source code
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
