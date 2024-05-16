// A generated module for ContainerImage functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

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

func (cmg *ContainerImage) ImageTag(
	ctx context.Context,
) (*ContainerImage, error) {
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
		commitHash, err := dag.Git().
			Load(source).
			Command([]string{"rev-parse", "HEAD"}).Stdout(ctx)
		if err != nil {
			return nil, err
		}
		parsedRef, err := dag.Gitter().WithRef(cmg.Ref).ParseRef(ctx)
		if err != nil {
			return nil, err
		}
		genTag = fmt.Sprintf(
			"%s-%s",
			parsedRef,
			formatSha(commitHash),
		)
	}
	cmg.DockerImageTag = genTag
	return cmg, nil
}

func (cmg *ContainerImage) FakePublishFromRepo(
	ctx context.Context,
) *Container {
	return new(Container)
	/* var genTag string
	switch {
	case semverRe.MatchString(ref):
		genTag = ref
	case shaRe.MatchString(ref):
		genTag = fmt.Sprintf("sha-%s", formatSha(ref))
	default:
		commitHash, err := dag.Git().
			Load(source).
			Command([]string{"rev-parse", "HEAD"}).Stdout(ctx)
		if err != nil {
			return nil, err
		}
		parsedRef, err := dag.Gitter().WithRef(ref).ParseRef(ctx)
		if err != nil {
			return nil, err
		}
		genTag = fmt.Sprintf(
			"%s-%s",
			parsedRef,
			formatSha(commitHash),
		)
	} */
}

func formatSha(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
