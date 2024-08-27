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
	githubURL  = "https://github.com"
)

type Golang struct {
	ArangoPassword     string
	ArangoVersion      string
	ArangoPort         int
	RedisPassword      string
	RedisVersion       string
	RedisPort          int
	GolangVersion      string
	GotestSumFormatter string
}

// Test runs Go tests
func (gom *Golang) Test(
	ctx context.Context,
	// The source directory to test, Required.
	src *Directory,
	// An optional slice of strings representing additional arguments to the go test command
	// +optional
	args []string,
) (string, error) {
	return gom.PrepareTestContainer(ctx).
		WithMountedDirectory(PROJ_MOUNT, src).
		WithWorkdir(PROJ_MOUNT).
		WithExec([]string{"go", "mod", "download"}).
		WithExec(append([]string{
			"gotestsum",
			"--format-hide-empty-pkg",
			"--format", gom.GotestSumFormatter, "--",
		}, args...)).
		Stdout(ctx)
}

// Lint runs golangci-lint on the Go source code in a containerized environment.
// It uses a specified version of golangci-lint to perform static code analysis.
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

func fetchAndValidateEnvVars(envVar string) (string, error) {
	value := os.Getenv(envVar)
	if len(value) == 0 {
		return "", fmt.Errorf("value of %s env variable is not set", envVar)
	}
	return value, nil
}

// WithGolangVersion sets the version of Golang to use.
func (gom *Golang) WithGolangVersion(
	// The version of Golang to use
	// +optional
	// +default="1.22.6"
	version string,
) *Golang {
	gom.GolangVersion = version
	return gom
}

// WithGotestSumFormatter sets the output formatter for gotestsum
func (gom *Golang) WithGotestSumFormatter(
	// The output formatter to use for gotestsum
	// +optional
	// +default="pkgname"
	formatter string,
) *Golang {
	gom.GotestSumFormatter = formatter
	return gom
}

// PrepareTestContainer creates a container with Golang and installs gotestsum and gotestdox
func (gom *Golang) PrepareTestContainer(
	ctx context.Context,
) *Container {
	return dag.Container().
		From(fmt.Sprintf("golang:%s-alpine", gom.GolangVersion)).
		WithExec([]string{"apk", "add", "--no-cache", "git"}).
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@latest"}).
		WithExec([]string{"go", "install", "github.com/bitfield/gotestdox/cmd/gotestdox@latest"})
}
