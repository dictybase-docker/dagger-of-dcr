// Package main provides a set of functions to run Go language tests within a containerized environment.
package main

import (
	"context"
	"fmt"

	"dagger/golang/internal/dagger"

	F "github.com/IBM/fp-go/function"
)

const (
	projMount             = "/app"
	lintBase              = "golangci/golangci-lint"
	githubURL             = "https://github.com"
	goModDownloadCommand  = "download"
	goModSubcommand       = "mod"
	goTestSumCommand      = "gotestsum"
	goTestSumFormatOption = "--format"
	apkCommand            = "apk"
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
	src *dagger.Directory,
	// An optional slice of strings representing additional arguments to the go test command
	// +optional
	args []string,
) (string, error) {
	return gom.PrepareTestContainer().
		WithMountedDirectory(projMount, src).
		WithWorkdir(projMount).
		WithExec([]string{"go", goModSubcommand, goModDownloadCommand}).
		WithExec(append([]string{
			goTestSumCommand,
			"--format-hide-empty-pkg",
			goTestSumFormatOption, gom.GotestSumFormatter, "--",
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
	src *dagger.Directory,
	// An optional slice of strings representing additional arguments to the
	// golangci-lint command
	// +optional
	args []string,
) (string, error) {
	return F.Pipe3(
		dag.Container(),
		base(fmt.Sprintf("%s:%s", lintBase, version)),
		prepareWorkspace(src, projMount),
		goLintRunner(args),
	).Stdout(ctx)
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
func (gom *Golang) PrepareTestContainer() *dagger.Container {
	return dag.Container().
		From(fmt.Sprintf("golang:%s-alpine", gom.GolangVersion)).
		WithExec([]string{apkCommand, "add", "--no-cache", "git"}).
		WithExec([]string{"go", "install", "gotest.tools/gotestsum@latest"}).
		WithExec([]string{"go", "install", "github.com/bitfield/gotestdox/cmd/gotestdox@latest"})
}

// TestGitHub runs Go tests on a GitHub repository
func (gom *Golang) TestGitHub(
	ctx context.Context,
	// The GitHub repository name (e.g., "username/repo")
	repository string,
	// The git reference (branch, tag, or commit) to clone and test
	gitRef string,
	// An optional slice of strings representing additional arguments to the go test command
	// +optional
	args []string,
) (string, error) {
	source := dag.Gitter().
		WithRef(gitRef).
		WithRepository(fmt.Sprintf("%s/%s", githubURL, repository)).
		Checkout()
	return gom.Test(ctx, source, args)
}

// LintGitHub runs golangci-lint on a GitHub repository
func (gom *Golang) LintGitHub(
	ctx context.Context,
	// The GitHub repository name (e.g., "username/repo")
	repository string,
	// The git reference (branch, tag, or commit) to clone and lint
	gitRef string,
	// An optional string specifying the version of golangci-lint to use
	// +optional
	// +default="v1.55.2-alpine"
	version string,
	// An optional slice of strings representing additional arguments to the golangci-lint command
	// +optional
	args []string,
) (string, error) {
	source := dag.Gitter().
		WithRef(gitRef).
		WithRepository(fmt.Sprintf("%s/%s", githubURL, repository)).
		Checkout()
	return gom.Lint(ctx, version, source, args)
}
