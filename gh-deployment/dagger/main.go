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
	// Docker Image tag
	DockerImageTag string
	// Application name
	Application string
	// Stack name
	Stack string
	// Run ID
	RunId int
	// Environment name, default is "development"
	Environment string
	// Dockerfile path
	Dockerfile string
	// Docker namespace
	DockerNamespace string
	// Docker image
	DockerImage string
}

// CreateGitHubDeployment creates a GitHub deployment
func (ghd *GhDeployment) CreateGithubDeployment(
	ctx context.Context,
	// Github token for making api requests
	token string,
) (int, error) {
	var dplId int

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
			"dockerfile":       ghd.Dockerfile,
			"dagger_version":   ghd.DaggerVersion,
			"dagger_checksum":  ghd.DaggerChecksum,
			"cluster":          ghd.Cluster,
			"storage":          ghd.Storage,
			"kube_config":      ghd.KubeConfig,
			"artifact":         ghd.Artifact,
			"docker_image":     ghd.DockerImageTag,
			"docker_image_tag": ghd.DockerImageTag,
			"docker_namespace": ghd.DockerNamespace,
			"application":      ghd.Application,
			"stack":            ghd.Stack,
			"run_id":           ghd.RunId,
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
	return int(dpl.GetID()), nil
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
