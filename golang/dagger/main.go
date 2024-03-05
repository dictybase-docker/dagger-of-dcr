// Package main provides a set of functions to run Go language tests within a containerized environment.
package main

import (
	"context"

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
