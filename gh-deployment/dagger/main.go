package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"
)

type GhDeployment struct {
	// Repository name with owner, for example, "tora/bora"
	Repository string
	// Git reference, for example, "refs/heads/main"
	Ref string
	// Dagger version, for example, "v0.11.6"
	DaggerVersion string
	// Dagger checksum
	DaggerChecksum string
	// Cluster name
	Cluster string
	// Storage name
	Storage string
	// Kube config file path
	KubeConfig string
	// Artifact name
	Artifact string
	// Image tag
	ImageTag string
	// Application name
	Application string
	// Stack name
	Stack string
	// Run ID
	RunId int
	// Environment name, default is "development"
	Environment string
}

// CreateGitHubDeployment creates a GitHub deployment
func (ghd *GhDeployment) CreateGithubDeployment(
	ctx context.Context,
	// Github token for makding api requests
	token string,
) (int, error) {
	var dplId int
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	ownerRepo := ghd.Repository
	owner, repo, err := parseOwnerRepo(ownerRepo)
	if err != nil {
		return dplId, err
	}

	deploymentRequest := &github.DeploymentRequest{
		Environment:      github.String(ghd.Environment),
		AutoMerge:        github.Bool(false),
		Ref:              github.String(ghd.Ref),
		Description:      github.String("Deployment created by GhDeployment"),
		RequiredContexts: &[]string{}, // Skip status checks
		Payload: map[string]interface{}{
			"dagger_version":  ghd.DaggerVersion,
			"dagger_checksum": ghd.DaggerChecksum,
			"cluster":         ghd.Cluster,
			"storage":         ghd.Storage,
			"kube_config":     ghd.KubeConfig,
			"artifact":        ghd.Artifact,
			"image_tag":       ghd.ImageTag,
			"application":     ghd.Application,
			"stack":           ghd.Stack,
			"run_id":          ghd.RunId,
			"repository":      ghd.Repository,
		},
	}

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
	return int(dpl.GetID()), nil
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

// WithDaggerVersion sets the Dagger version
func (ghd *GhDeployment) WithDaggerVersion(
	// Dagger version, for example, "v0.11.6", Required
	daggerVersion string,
) (*GhDeployment, error) {
	if len(daggerVersion) == 0 {
		return ghd, errors.New("daggerVersion value is required")
	}
	ghd.DaggerVersion = daggerVersion
	return ghd, nil
}

// WithDaggerChecksum sets the Dagger checksum
func (ghd *GhDeployment) WithDaggerChecksum(
	// Dagger checksum, Required
	daggerChecksum string,
) (*GhDeployment, error) {
	if len(daggerChecksum) == 0 {
		return ghd, errors.New("daggerChecksum value is required")
	}
	ghd.DaggerChecksum = daggerChecksum
	return ghd, nil
}

// WithCluster sets the cluster
func (ghd *GhDeployment) WithCluster(
	// Cluster, Required
	cluster string,
) (*GhDeployment, error) {
	if len(cluster) == 0 {
		return ghd, errors.New("cluster value is required")
	}
	ghd.Cluster = cluster
	return ghd, nil
}

// WithStorage sets the storage
func (ghd *GhDeployment) WithStorage(
	// Storage, Required
	storage string,
) (*GhDeployment, error) {
	if len(storage) == 0 {
		return ghd, errors.New("storage value is required")
	}
	ghd.Storage = storage
	return ghd, nil
}

// WithKubeConfig sets the kube config
func (ghd *GhDeployment) WithKubeConfig(
	// Kube config, Required
	kubeConfig string,
) (*GhDeployment, error) {
	if len(kubeConfig) == 0 {
		return ghd, errors.New("kubeConfig value is required")
	}
	ghd.KubeConfig = kubeConfig
	return ghd, nil
}

// WithArtifact sets the artifact
func (ghd *GhDeployment) WithArtifact(
	// Artifact, Required
	artifact string,
) (*GhDeployment, error) {
	if len(artifact) == 0 {
		return ghd, errors.New("artifact value is required")
	}
	ghd.Artifact = artifact
	return ghd, nil
}

// WithImageTag sets the image tag
func (ghd *GhDeployment) WithImageTag(
	// Image tag, Required
	imageTag string,
) (*GhDeployment, error) {
	if len(imageTag) == 0 {
		return ghd, errors.New("imageTag value is required")
	}
	ghd.ImageTag = imageTag
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

// WithRunId sets the run ID
func (ghd *GhDeployment) WithRunId(
	// Run ID, Required
	runId int,
) (*GhDeployment, error) {
	if runId == 0 {
		return ghd, errors.New("runId value is required")
	}
	ghd.RunId = runId
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
