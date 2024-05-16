/*
Package main provides functionality to manage and build Docker container images based on Git references.
It includes methods to set various properties of the container image and generate appropriate Docker image tags.
*/
package main

import (
	"context"
	"fmt"
	"regexp"
)

var (
	shaRe    = regexp.MustCompile("^[0-9a-f]{7,40}$")
	semverRe = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)
)

type ContainerImage struct {
	// Repository name
	Repository string
	// Git reference
	Ref string
	// The docker namespace under which the image will be pushed
	Namespace string
	// Specifies the path to the Dockerfile
	Dockerfile string
	// Name of the image to be built
	Image          string
	DockerImageTag string
}

// WithNamespace sets the docker namespace
func (cmg *ContainerImage) WithNamespace(
	// The docker namespace under which the image will be pushed
	// +optional
	// +default="dictybase"
	namespace string,
) *ContainerImage {
	cmg.Namespace = namespace
	return cmg
}

// WithRef sets the Git reference (branch, tag, or SHA)
func (cmg *ContainerImage) WithRef(
	// the branch, tag or sha to checkout
	ref string,
) *ContainerImage {
	cmg.Ref = ref
	return cmg
}

// WithRepository sets the GitHub repository name
func (cmg *ContainerImage) WithRepository(
	// github repository name with owner, for example tora/bora, Required
	repository string,
) *ContainerImage {
	cmg.Repository = repository
	return cmg
}

// WithDockerfile sets the Dockerfile path
func (cmg *ContainerImage) WithDockerfile(
	// specifies the path to the Dockerfile
	// +optional
	// +default="build/package/Dockerfile"
	dockerFile string,
) *ContainerImage {
	cmg.Dockerfile = dockerFile
	return cmg
}

// WithImage sets the image name
func (cmg *ContainerImage) WithImage(
	// name of the image to be built
	image string,
) *ContainerImage {
	cmg.Image = image
	return cmg
}

func (cmg *ContainerImage) PublishFromRepo(
	ctx context.Context,
	// dockerhub user name
	user string,
	// dockerhub password, use an api token
	password string,
) (string, error) {
	cont, err := cmg.ImageTag(ctx)
	if err != nil {
		return "", err
	}
	return cont.WithRegistryAuth(
		"docker.io",
		user,
		dag.SetSecret("docker-pass", password),
	).Publish(
		ctx,
		fmt.Sprintf(
			"%s/%s:%s",
			cmg.Namespace,
			cmg.Image,
			cmg.DockerImageTag,
		),
	)
}

// FakePublishFromRepo publishes a container image to a temporary repository with a time-to-live of 10 minutes.
func (cmg *ContainerImage) FakePublishFromRepo(
	ctx context.Context,
) (string, error) {
	cont, err := cmg.ImageTag(ctx)
	if err != nil {
		return "", err
	}
	return cont.Publish(
		ctx,
		fmt.Sprintf(
			"ttl.sh/%s-%s-%s:10m",
			cmg.Namespace,
			cmg.Image,
			cmg.DockerImageTag,
		),
	)
}

// ImageTag generates a Docker image tag based on the provided Git reference
func (cmg *ContainerImage) ImageTag(
	ctx context.Context,
) (*Container, error) {
	source := dag.Gitter().
		WithRef(cmg.Ref).
		WithRepository(cmg.Repository).
		Checkout()
	var genTag string
	switch {
	case semverRe.MatchString(cmg.Ref):
		genTag = cmg.Ref
	case shaRe.MatchString(cmg.Ref):
		genTag = fmt.Sprintf("sha-%s", formatSha(cmg.Ref))
	default:
		dtag, err := cmg.generateDefaultTag(ctx, source)
		if err != nil {
			return nil, err
		}
		genTag = dtag
	}
	cmg.DockerImageTag = genTag
	return dag.Container().
		Build(source, ContainerBuildOpts{Dockerfile: cmg.Dockerfile}), nil
}

func (cmg *ContainerImage) generateDefaultTag(
	ctx context.Context,
	source *Directory,
) (string, error) {
	commitHash, err := dag.Git().
		Load(source).
		Command([]string{"rev-parse", "HEAD"}).Stdout(ctx)
	if err != nil {
		return "", err
	}
	parsedRef, err := dag.Gitter().WithRef(cmg.Ref).ParseRef(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"%s-%s",
		parsedRef,
		formatSha(commitHash),
	), nil
}

func formatSha(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
