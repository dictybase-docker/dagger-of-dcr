package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

const (
	githubURL = "https://github.com"
)

var (
	shaRe    = regexp.MustCompile("^[0-9a-f]{7,40}$")
	semverRe = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)
	bre      = regexp.MustCompile(`^refs/heads|tags/(.+)$`)
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
	Image string
	// Name of the docker image tag
	DockerImageTag string
}

// Payload represents the payload information for a deployment
type Payload struct {
	Repository      string `json:"repository"`
	DaggerVersion   string `json:"dagger_version"`
	DaggerChecksum  string `json:"dagger_checksum"`
	Cluster         string `json:"cluster"`
	Storage         string `json:"storage"`
	KubeConfig      string `json:"kube_config"`
	Artifact        string `json:"artifact"`
	DockerImage     string `json:"docker_image"`
	DockerImageTag  string `json:"docker_image_tag"`
	Dockerfile      string `json:"dockerfile"`
	DockerNamespace string `json:"docker_namespace"`
	Application     string `json:"application"`
	Stack           string `json:"stack"`
	RunId           int    `json:"run_id"`
	Environment     string `json:"environment"`
}

// WithNamespace sets the docker namespace
func (cmg *ContainerImage) WithNamespace(
	ctx context.Context,
	// The docker namespace under which the image will be pushed
	// +default="dictybase"
	namespace string,
) *ContainerImage {
	cmg.Namespace = namespace
	return cmg
}

// WithRef sets the Git reference (branch, tag, or SHA)
func (cmg *ContainerImage) WithRef(
	ctx context.Context,
	// the branch, tag or sha to checkout
	ref string,
) *ContainerImage {
	cmg.Ref = ref
	return cmg
}

// WithRepository sets the GitHub repository name
func (cmg *ContainerImage) WithRepository(
	ctx context.Context,
	// github repository name with owner, for example tora/bora, Required
	repository string,
	// whether or not to prepend githubURL to the repository
	// +optional
	// +default=true
	shouldPrepend bool,
) *ContainerImage {
	if shouldPrepend {
		cmg.Repository = fmt.Sprintf("%s/%s", githubURL, repository)
	} else {
		cmg.Repository = repository
	}
	return cmg
}

// WithDockerfile sets the Dockerfile path
func (cmg *ContainerImage) WithDockerfile(
	ctx context.Context,
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
	ctx context.Context,
	// name of the image to be built
	image string,
) *ContainerImage {
	cmg.Image = image
	return cmg
}

// PublishFromRepo publishes a container image to Docker Hub
func (cmg *ContainerImage) PublishFromRepo(
	ctx context.Context,
	// dockerhub user name
	user string,
	// dockerhub password, use an api token
	password string,
) (string, error) {
	cont, err := cmg.GenerateImageTag(ctx)
	if err != nil {
		return "", err
	}
	_, err = cont.WithRegistryAuth(
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
	if err != nil {
		return "", fmt.Errorf("error in publishing docker container %s", err)
	}
	return cmg.DockerImageTag, nil
}

// PublishFromRepoWithDeploymentID publishes a container image to Docker Hub
// using deployment information from a specified GitHub deployment ID.
func (cmg *ContainerImage) PublishFromRepoWithDeploymentID(
	ctx context.Context,
	// dockerhub user name
	user string,
	// dockerhub password, use an api token
	password string,
	// deployment ID
	deploymentID string,
	// GitHub token for making API requests
	token string,
) error {
	return cmg.publishFromRepoWithDeploymentIDCommon(
		ctx,
		user,
		password,
		deploymentID,
		token,
		func(source *Directory, dpl *github.Deployment, pload Payload) *Container {
			return dag.Container().
				Build(source, ContainerBuildOpts{Dockerfile: pload.Dockerfile})
		},
	)
}

// FakePublishFromRepo publishes a container image to a temporary repository with a time-to-live of 10 minutes.
func (cmg *ContainerImage) FakePublishFromRepo(
	ctx context.Context,
) (string, error) {
	cont, err := cmg.GenerateImageTag(ctx)
	if err != nil {
		return "", err
	}
	return cont.Publish(
		ctx,
		fmt.Sprintf(
			"ttl.sh/%s-%s-%s:10m",
			cmg.Namespace,
			cmg.Image,
			cmg.GenerateImageTag,
		),
	)
}

// ImageTag generates a Docker image tag based on the provided Git reference
func (cmg *ContainerImage) GenerateImageTag(
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
	case bre.MatchString(cmg.Ref):
		match := bre.FindStringSubmatch(cmg.Ref)
		genTag = match[1]
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

// parseOwnerRepo splits the repository string into owner and repo
func parseOwnerRepo(ownerRepo string) (string, string, error) {
	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf(
			"invalid repository format, expected 'owner/repo'",
		)
	}
	return parts[0], parts[1], nil
}

// PublishFrontendFromRepoWithDeploymentID publishes a frontend container image to Docker Hub
// using deployment information from a specified GitHub deployment ID.
func (cmg *ContainerImage) PublishFrontendFromRepoWithDeploymentID(
	ctx context.Context,
	// dockerhub user name
	user string,
	// dockerhub password, use an api token
	password string,
	// deployment ID
	deploymentID string,
	// GitHub token for making API requests
	token string,
) error {
	owner, repo, err := parseOwnerRepo(cmg.Repository)
	if err != nil {
		return err
	}
	depId, err := strconv.ParseInt(deploymentID, 10, 64)
	if err != nil {
		return fmt.Errorf("error in converting string to int64 %s", err)
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	deployment, _, err := client.Repositories.GetDeployment(
		ctx,
		owner,
		repo,
		depId,
	)
	if err != nil {
		return fmt.Errorf(
			"error in getting deployment information: %s",
			err,
		)
	}
	var pload Payload
	if err := json.Unmarshal(deployment.Payload, &pload); err != nil {
		return fmt.Errorf("error in decoding payload %s", err)
	}
	if pload.Repository != cmg.Repository {
		return fmt.Errorf(
			"payload repo %s and given repo %s does not match",
			pload.Repository,
			cmg.Repository,
		)
	}
	source := dag.Gitter().
		WithRef(deployment.GetRef()).
		WithRepository(fmt.Sprintf("%s/%s", githubURL, pload.Repository)).
		Checkout()
	allImages := strings.Split(pload.DockerImage, ":")
	allDockerfiles := strings.Split(pload.Dockerfile, ":")
	grp, ctx := errgroup.WithContext(ctx)
	for idx, file := range allDockerfiles {
		grp.Go(func() error {
			_, err := dag.Container().
				Build(source, ContainerBuildOpts{
					Dockerfile: file,
					BuildArgs: []BuildArg{
						{
							Name:  "BUILD_STATE",
							Value: deployment.GetEnvironment(),
						},
					},
				}).
				WithRegistryAuth("docker.io", user, dag.SetSecret("docker-pass", password)).
				Publish(ctx, fmt.Sprintf(
					"%s/%s:%s",
					pload.DockerNamespace,
					allImages[idx],
					pload.DockerImageTag,
				))
			if err != nil {
				return fmt.Errorf(
					"error in publishing docker container %s",
					err,
				)
			}
			return nil
		})
	}
	return grp.Wait()
}

// publishFromRepoWithDeploymentIDCommon is a common method for publishing container images
func (cmg *ContainerImage) publishFromRepoWithDeploymentIDCommon(
	ctx context.Context,
	user string,
	password string,
	deploymentID string,
	token string,
	buildFunc func(source *Directory, deployment *github.Deployment, pload Payload) *Container,
) error {
	owner, repo, err := parseOwnerRepo(cmg.Repository)
	if err != nil {
		return err
	}
	depId, err := strconv.ParseInt(deploymentID, 10, 64)
	if err != nil {
		return fmt.Errorf("error in converting string to int64 %s", err)
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	deployment, _, err := client.Repositories.GetDeployment(
		ctx,
		owner,
		repo,
		depId,
	)
	if err != nil {
		return fmt.Errorf(
			"error in getting deployment information: %s",
			err,
		)
	}
	var pload Payload
	if err := json.Unmarshal(deployment.Payload, &pload); err != nil {
		return fmt.Errorf("error in decoding payload %s", err)
	}
	if pload.Repository != cmg.Repository {
		return fmt.Errorf(
			"payload repo %s and given repo %s does not match",
			pload.Repository,
			cmg.Repository,
		)
	}
	source := dag.Gitter().
		WithRef(deployment.GetRef()).
		WithRepository(fmt.Sprintf("%s/%s", githubURL, pload.Repository)).
		Checkout()

	container := buildFunc(source, deployment, pload)

	_, err = container.
		WithRegistryAuth(
			"docker.io",
			user,
			dag.SetSecret("docker-pass", password),
		).Publish(
		ctx,
		fmt.Sprintf(
			"%s/%s:%s",
			pload.DockerNamespace,
			pload.DockerImage,
			pload.DockerImageTag,
		),
	)
	if err != nil {
		return fmt.Errorf("error in publishing docker container %s", err)
	}
	return nil
}
