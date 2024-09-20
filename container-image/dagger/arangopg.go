package main

import (
	"context"
	"fmt"
)

// CreateArangoPostgresContainer creates a container based on ArangoDB 3.10.14,
// updates it, installs PostgreSQL 14, and restic on the linux/amd64 platform
func (cmg *ContainerImage) CreateArangoPostgresContainer(
	ctx context.Context,
) (*Container, error) {
	return dag.Container(ContainerOpts{Platform: "linux/amd64"}).
			From("arangodb:3.11.6").
			WithExec([]string{"apk", "update"}).
			WithExec([]string{"apk", "add", "postgresql14", "curl", "bzip2", "redis-cli"}).
			WithExec([]string{"curl", "-L", "https://github.com/restic/restic/releases/download/v0.17.0/restic_0.17.0_linux_amd64.bz2", "-o", "/tmp/restic_0.17.0_linux_amd64.bz2"}).
			WithExec([]string{"bunzip2", "/tmp/restic_0.17.0_linux_amd64.bz2"}).
			WithExec([]string{"mv", "/tmp/restic_0.17.0_linux_amd64", "/usr/local/bin/restic"}).
			WithExec([]string{"chmod", "+x", "/usr/local/bin/restic"}).
			WithExec([]string{"rm", "-f", "/tmp/restic_0.17.0_linux_amd64.bz2"}),
		nil
}

// BuildAndPublishArangoPostgresContainer builds and publishes the ArangoPostgres container
func (cmg *ContainerImage) BuildAndPublishArangoPostgresContainer(
	ctx context.Context,
	// dockerhub user name
	user string,
	// dockerhub password, use an api token
	password string,
) (string, error) {
	container, err := cmg.CreateArangoPostgresContainer(ctx)
	if err != nil {
		return "", fmt.Errorf(
			"error creating ArangoPostgres container: %w",
			err,
		)
	}

	tag := fmt.Sprintf("%s/%s:%s", cmg.Namespace, cmg.Image, cmg.Ref)
	_, err = container.
		WithRegistryAuth(
			"docker.io",
			user,
			dag.SetSecret("docker-pass", password),
		).
		Publish(ctx, tag)
	if err != nil {
		return "", fmt.Errorf(
			"error publishing ArangoPostgres container: %w",
			err,
		)
	}

	return tag, nil
}
