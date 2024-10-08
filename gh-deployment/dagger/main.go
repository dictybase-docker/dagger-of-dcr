package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"
)

const (
	githubURL = "https://github.com"
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
	// Kubectl file path
	KubectlFile string // New public attribute
}

// CreateGitHubDeployment creates a GitHub deployment
func (ghd *GhDeployment) CreateGithubDeployment(
	ctx context.Context,
	// Github token for making api requests
	token string,
) (string, error) {
	var dplId string

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
			"kubectl_config":   ghd.KubectlFile,
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
	return fmt.Sprintf("%d", dpl.GetID()), nil
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

// WithKubectlFile sets the kubectl file path
func (ghd *GhDeployment) WithKubectlFile(
	// Kubectl file path, Required
	kubectlFile string,
) (*GhDeployment, error) {
	if len(kubectlFile) == 0 {
		return ghd, errors.New("kubectlFile value is required")
	}
	ghd.KubectlFile = kubectlFile
	return ghd, nil
}

func (ghd *GhDeployment) GenerateImageTag(
	ctx context.Context,
) error {
	source := dag.Gitter().
		WithRef(ghd.Ref).
		WithRepository(fmt.Sprintf("%s/%s", githubURL, ghd.Repository)).
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
	deploymentID string,
	// Status, Required
	status string,
	// Github token for making api requests, Required
	token string,
) error {
	owner, repo, err := parseOwnerRepo(ghd.Repository)
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
	_, _, err = client.Repositories.CreateDeploymentStatus(
		ctx,
		owner,
		repo,
		depId,
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

// ListDeployments lists deployments for the GitHub repository
func (ghd *GhDeployment) ListGithubDeployments(
	ctx context.Context,
	// Github token for making api requests
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

	opts := &github.DeploymentsListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		deployments, resp, err := client.Repositories.ListDeployments(
			ctx,
			owner,
			repo,
			opts,
		)
		if err != nil {
			return fmt.Errorf("error listing deployments: %v", err)
		}

		for _, dpl := range deployments {
			fmt.Printf(
				"[Deployment ID]: %d, [Description]: %s [Environment]: %s\n",
				dpl.GetID(),
				dpl.GetDescription(),
				dpl.GetEnvironment(),
			)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

// ListDeploymentsWithStatus fetches deployments and their statuses for the GitHub repository
func (ghd *GhDeployment) ListDeploymentsWithStatus(
	ctx context.Context,
	// Github token for making api requests
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

	opts := &github.DeploymentsListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		deployments, resp, err := client.Repositories.ListDeployments(
			ctx,
			owner,
			repo,
			opts,
		)
		if err != nil {
			return fmt.Errorf("error listing deployments: %v", err)
		}

		for _, dpl := range deployments {
			fmt.Printf(
				"[Deployment ID]: %d, [Description]: %s\n",
				dpl.GetID(),
				dpl.GetDescription(),
			)

			// Fetch deployment statuses
			statusOpts := &github.ListOptions{PerPage: 100}
			for {
				statuses, statusResp, err := client.Repositories.ListDeploymentStatuses(
					ctx,
					owner,
					repo,
					dpl.GetID(),
					statusOpts,
				)
				if err != nil {
					return fmt.Errorf(
						"error listing deployment statuses: %v",
						err,
					)
				}

				for _, dps := range statuses {
					fmt.Printf(
						"[Status]: %s\n",
						dps.GetState(),
					)
				}

				if statusResp.NextPage == 0 {
					break
				}
				statusOpts.Page = statusResp.NextPage
			}
			fmt.Println()
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

// RemoveUnsuccessfulDeployments removes all deployments without a success status
func (ghd *GhDeployment) RemoveUnsuccessfulDeployments(
	ctx context.Context,
	// Github token for making api requests
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

	opts := &github.DeploymentsListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		deployments, resp, err := client.Repositories.ListDeployments(
			ctx,
			owner,
			repo,
			opts,
		)
		if err != nil {
			return fmt.Errorf("error listing deployments: %v", err)
		}

		for _, dpl := range deployments {
			hasSuccessStatus, err := ghd.checkDeploymentStatus(
				ctx,
				client,
				owner,
				repo,
				dpl.GetID(),
			)
			if err != nil {
				return err
			}
			if !hasSuccessStatus {
				_, err := client.Repositories.DeleteDeployment(
					ctx,
					owner,
					repo,
					dpl.GetID(),
				)
				if err != nil {
					return fmt.Errorf(
						"error deleting deployment %d: %v",
						dpl.GetID(),
						err,
					)
				}
				fmt.Printf("Removed deployment ID: %d\n", dpl.GetID())
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

func (ghd *GhDeployment) checkDeploymentStatus(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	deploymentID int64,
) (bool, error) {
	statusOpts := &github.ListOptions{PerPage: 100}
	allStatuses := make([]string, 0)

	for {
		statuses, resp, err := client.Repositories.ListDeploymentStatuses(
			ctx, owner, repo, deploymentID, statusOpts,
		)
		if err != nil {
			return false, fmt.Errorf(
				"error listing deployment statuses: %v",
				err,
			)
		}
		for _, dps := range statuses {
			allStatuses = append(allStatuses, dps.GetState())
		}
		if resp.NextPage == 0 {
			break
		}
		statusOpts.Page = resp.NextPage
	}
	if slices.Contains(allStatuses, "success") {
		return true, nil
	}

	return false, nil
}
