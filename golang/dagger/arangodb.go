package main

import (
	"context"
	"fmt"
)

// TestsWithArangoDB runs Go tests in a container with an ArangoDB.
func (gom *Golang) TestsWithArangoDB(
	ctx context.Context,
	// The source directory to test, Required.
	src *Directory,
	// An optional slice of strings representing additional arguments to the go test command
	// +optional
	args []string,
) (string, error) {
	arangoService := dag.Container().
		From(fmt.Sprintf("%s:%s", "arangodb", gom.ArangoVersion)).
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

	return gom.PrepareTestContainer(ctx).
		WithServiceBinding("arango", arangoService).
		WithEnvVariable("ARANGO_HOST", arangoHost).
		WithEnvVariable("ARANGO_PASS", gom.ArangoPassword).
		WithEnvVariable("ARANGO_USER", "root").
		WithMountedDirectory(PROJ_MOUNT, src).
		WithWorkdir(PROJ_MOUNT).
		WithExec([]string{"go", "mod", "download"}).
		WithExec(append([]string{
			"gotestsum", "--format", gom.GotestSumFormatter, "--",
		}, args...)).
		Stdout(ctx)
}

// TestsWithArangoDBFromGithub fetches a GitHub repository and runs Go tests with ArangoDB.
func (gom *Golang) TestsWithArangoDBFromGithub(
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
	// Call TestsWithArangoDB with the fetched directory
	return gom.TestsWithArangoDB(ctx, source, args)
}

// WithArangoPort sets the port to expose ArangoDB on.
func (gom *Golang) WithArangoPort(
	// The port to expose ArangoDB on
	// +optional
	// +default=8529
	port int,
) *Golang {
	gom.ArangoPort = port
	return gom
}

// WithArangoVersion sets the version of ArangoDB to use.
func (gom *Golang) WithArangoVersion(
	// The version of ArangoDB to use
	// +optional
	// +default="3.10.9"
	version string,
) *Golang {
	gom.ArangoVersion = version
	return gom
}

// WithArangoPassword sets the root password for the ArangoDB instance.
func (gom *Golang) WithArangoPassword(
	// The root password for the ArangoDB instance
	// +optional
	// +default="golam"
	password string,
) *Golang {
	gom.ArangoPassword = password
	return gom
}
