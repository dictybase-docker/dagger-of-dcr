package main

import (
	"context"
	"fmt"
)

// WithRedisVersion sets the version of Redis to use.
func (gom *Golang) WithRedisVersion(
	// The version of Redis to use
	// +optional
	// +default="7.0.12"
	version string,
) *Golang {
	gom.RedisVersion = version
	return gom
}

// WithRedisPort sets the port to expose Redis on.
func (gom *Golang) WithRedisPort(
	// The port to expose Redis on
	// +optional
	// +default=6379
	port int,
) *Golang {
	gom.RedisPort = port
	return gom
}

// TestsWithRedis runs Go tests in a container with Redis.
func (gom *Golang) TestsWithRedis(
	ctx context.Context,
	// The source directory to test, Required.
	src *Directory,
	// An optional slice of strings representing additional arguments to the go test command
	// +optional
	args []string,
) (string, error) {
	redisService := dag.Container().
		From(fmt.Sprintf("redis:%s-alpine", gom.RedisVersion)).
		WithExposedPort(gom.RedisPort).
		AsService()

	redisHost, err := redisService.Hostname(ctx)
	if err != nil {
		return redisHost, fmt.Errorf(
			"error in retrieving redis host %w",
			err,
		)
	}

	return gom.PrepareTestContainer(ctx).
		WithServiceBinding("redis", redisService).
		WithEnvVariable("REDIS_SERVICE_HOST", redisHost).
		WithEnvVariable("REDIS_SERVICE_PORT", fmt.Sprintf("%d", gom.RedisPort)).
		WithMountedDirectory(PROJ_MOUNT, src).
		WithWorkdir(PROJ_MOUNT).
		WithExec([]string{"go", "mod", "download"}).
		WithExec(append([]string{
			"gotestsum", "--format", gom.GotestSumFormatter, "--",
		}, args...)).
		Stdout(ctx)
}

// TestsWithRedisFromGithub fetches a GitHub repository and runs Go tests with ArangoDB.
func (gom *Golang) TestsWithRedisFromGithub(
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
	return gom.TestsWithRedis(ctx, source, args)
}
