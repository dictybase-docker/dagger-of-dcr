package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"
)

var (
	shaRe    = regexp.MustCompile("^[0-9a-f]{7,40}$")
	semverRe = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)
	bre      = regexp.MustCompile(`^refs/heads|tags/(.+)$`)
)

type GhDeployment struct {
	// Repository name with owner, for example, "tora/bora"
	Repository string
	// Git reference, for example, "refs/heads/main"
	Ref string
	// Docker Image tag
	DockerImageTag string
	// Application name
	Application string
	// Stack name
	Stack string
	// Environment name, default is "development"
	Environment string
	// Dockerfile path
	Dockerfile string
	// Docker namespace
	DockerNamespace string
	// Docker image
	DockerImage string
	// Project name
	Project string
}

// CreateGitHubDeployment creates a GitHub deployment
func (ghd *GhDeployment) CreateGithubDeployment(
	ctx context.Context,
	// Github token for making api requests
	token string,
) (int64, error) {
	var dplId int64

	owner, repo, err := parseOwnerRepo(ghd.Repository)
	if err != nil {
		return dplId, err
	}

	if len(ghd.DockerImageTag) == 0 {
		if err := ghd.GenerateImageTag(ctx); err != nil {
			return dplId, err
		}
	}

	deploymentRequest := &github.DeploymentRequest{
		Environment: github.String(ghd.Environment),
		AutoMerge:   github.Bool(false),
		Ref:         github.String(ghd.Ref),
		Description: github.String(
			fmt.Sprintf(
				"Deploying %s to %s environment",
				ghd.Application,
				ghd.Environment,
			),
		),
		RequiredContexts: &[]string{}, // Skip status checks
		Payload: map[string]interface{}{
			"project":          ghd.Project,
			"dockerfile":       ghd.Dockerfile,
			"docker_image":     ghd.DockerImage,
			"docker_image_tag": ghd.DockerImageTag,
			"docker_namespace": ghd.DockerNamespace,
			"application":      ghd.Application,
			"stack":            ghd.Stack,
			"repository":       ghd.Repository,
		},
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	dpl, _, err := client.Repositories.CreateDeployment(
		ctx,
		owner,
		repo,
		deploymentRequest,
	)
	if err != nil {
		return dplId, fmt.Errorf("error in creating deployment %s", err)
	}
	return dpl.GetID(), nil
}

// WithRepository sets the GitHub repository name
func (ghd *GhDeployment) WithRepository(
	// GitHub repository name with owner, for example, "tora/bora", Required
	repository string,
) (*GhDeployment, error) {
	if len(repository) == 0 {
		return ghd, errors.New("repository value is required")
	}
	ghd.Repository = repository
	return ghd, nil
}

// WithRef sets the Git reference (branch, tag, or SHA)
func (ghd *GhDeployment) WithRef(
	// Git reference, for example, "refs/heads/main", Required
	ref string,
) (*GhDeployment, error) {
	if len(ref) == 0 {
		return ghd, errors.New("ref value is required")
	}
	ghd.Ref = ref
	return ghd, nil
}

// WithImageTag sets the docker image tag
func (ghd *GhDeployment) WithImageTag(
	// Image tag, optional
	imageTag string,
) (*GhDeployment, error) {
	ghd.DockerImageTag = imageTag
	return ghd, nil
}

// WithApplication sets the application
func (ghd *GhDeployment) WithApplication(
	// Application, Required
	application string,
) (*GhDeployment, error) {
	if len(application) == 0 {
		return ghd, errors.New("application value is required")
	}
	ghd.Application = application
	return ghd, nil
}

// WithStack sets the stack
func (ghd *GhDeployment) WithStack(
	// Stack, Required
	stack string,
) (*GhDeployment, error) {
	if len(stack) == 0 {
		return ghd, errors.New("stack value is required")
	}
	ghd.Stack = stack
	return ghd, nil
}

// WithEnvironment sets the environment with a default value of "development"
func (ghd *GhDeployment) WithEnvironment(
	// Environment, Optional
	environment string,
	// +default="development"
) (*GhDeployment, error) {
	ghd.Environment = environment
	return ghd, nil
}

// WithDockerfile sets the Dockerfile path
func (ghd *GhDeployment) WithDockerfile(
	// Dockerfile path, Required
	dockerfile string,
) (*GhDeployment, error) {
	if len(dockerfile) == 0 {
		return ghd, errors.New("dockerfile value is required")
	}
	ghd.Dockerfile = dockerfile
	return ghd, nil
}

// WithDockerNamespace sets the Docker namespace
func (ghd *GhDeployment) WithDockerNamespace(
	// Docker namespace, Required
	dockerNamespace string,
) (*GhDeployment, error) {
	if len(dockerNamespace) == 0 {
		return ghd, errors.New("dockerNamespace value is required")
	}
	ghd.DockerNamespace = dockerNamespace
	return ghd, nil
}

// WithDockerImage sets the Docker image
func (ghd *GhDeployment) WithDockerImage(
	// Docker image, Required
	dockerImage string,
) (*GhDeployment, error) {
	if len(dockerImage) == 0 {
		return ghd, errors.New("dockerImage value is required")
	}
	ghd.DockerImage = dockerImage
	return ghd, nil
}

// WithProject sets the project name
func (ghd *GhDeployment) WithProject(
	// Project name, Required
	project string,
) (*GhDeployment, error) {
	if len(project) == 0 {
		return ghd, errors.New("project value is required")
	}
	ghd.Project = project
	return ghd, nil
}

func (ghd *GhDeployment) GenerateImageTag(
	ctx context.Context,
) error {
	source := dag.Gitter().
		WithRef(ghd.Ref).
		WithRepository(ghd.Repository).
		Checkout()
	var genTag string
	switch {
	case semverRe.MatchString(ghd.Ref):
		genTag = ghd.Ref
	case shaRe.MatchString(ghd.Ref):
		genTag = fmt.Sprintf("sha-%s", formatSha(ghd.Ref))
	case bre.MatchString(ghd.Ref):
		match := bre.FindStringSubmatch(ghd.Ref)
		genTag = match[1]
	default:
		dtag, err := ghd.generateDefaultTag(ctx, source)
		if err != nil {
			return err
		}
		genTag = dtag
	}
	ghd.DockerImageTag = genTag
	return nil
}

// SetDeploymentStatus sets the deployment status using the GitHub API
func (ghd *GhDeployment) SetDeploymentStatus(
	ctx context.Context,
	// Deployment ID, Required
	deploymentID int,
	// Status, Required
	status string,
	// Github token for making api requests, Required
	token string,
) error {
	owner, repo, err := parseOwnerRepo(ghd.Repository)
	if err != nil {
		return err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	_, _, err = client.Repositories.CreateDeploymentStatus(
		ctx,
		owner,
		repo,
		int64(deploymentID),
		&github.DeploymentStatusRequest{State: github.String(status)},
	)
	if err != nil {
		return fmt.Errorf("error in setting deployment status: %s", err)
	}
	return nil
}

func (ghd *GhDeployment) generateDefaultTag(
	ctx context.Context,
	source *Directory,
) (string, error) {
	commitHash, err := dag.Git().
		Load(source).
		Command([]string{"rev-parse", "HEAD"}).Stdout(ctx)
	if err != nil {
		return "", err
	}
	parsedRef, err := dag.Gitter().WithRef(ghd.Ref).ParseRef(ctx)
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
			"invalid repository format, expected owner/repo %s",
			ownerRepo,
		)
	}
	return parts[0], parts[1], nil
}
