package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"
)

const (
	pulumiOpsRepo   = "https://github.com/dictybase-docker/cluster-ops.git"
	pulumiOpsBranch = "master"
	githubURL       = "https://github.com"
)

type Payload struct {
	Project         string `json:"project"`
	Repository      string `json:"repository"`
	DaggerVersion   string `json:"dagger_version"`
	DaggerChecksum  string `json:"dagger_checksum"`
	Cluster         string `json:"cluster"`
	Storage         string `json:"storage"`
	KubeConfig      string `json:"kubectl_config"`
	Artifact        string `json:"artifact"`
	DockerImage     string `json:"docker_image"`
	DockerImageTag  string `json:"docker_image_tag"`
	Dockerfile      string `json:"dockerfile"`
	DockerNamespace string `json:"docker_namespace"`
	Application     string `json:"application"`
	Stack           string `json:"stack"`
	RunId           int    `json:"run_id"`
}

// PulumiOps represents the Pulumi operations configuration.
type PulumiOps struct {
	Backend     string
	Version     string
	KubeConfig  *File
	Credentials *File
	// Repository name
	Repository string
}

// WithPulumi sets the Pulumi version for the PulumiOps instance.
func (pmo *PulumiOps) WithPulumi(
	ctx context.Context,
	// pulumi version
	// + default="3.108.0"
	version string,
) *PulumiOps {
	pmo.Version = version
	return pmo
}

// WithBackend sets the backend for storing state for the PulumiOps instance.
func (pmo *PulumiOps) WithBackend(
	ctx context.Context,
	// pulumi backend for storing state, required
	backend string,
) *PulumiOps {
	pmo.Backend = backend
	return pmo
}

// WithCredentials sets the credentials file for the PulumiOps instance.
func (pmo *PulumiOps) WithCredentials(
	ctx context.Context,
	// credentials file(google cloud for the time being), required
	credentials *File,
) *PulumiOps {
	pmo.Credentials = credentials
	return pmo
}

// WithKubeConfig sets the Kubernetes configuration file for the PulumiOps instance.
func (pmo *PulumiOps) WithKubeConfig(
	ctx context.Context,
	// kubernetes configuration file, required
	config *File,
) *PulumiOps {
	pmo.KubeConfig = config
	return pmo
}

// WithRepository sets the repository name
func (pmo *PulumiOps) WithRepository(
	ctx context.Context,
	// Repository name, Required
	repository string,
) *PulumiOps {
	pmo.Repository = repository
	return pmo
}

// PulumiContainer returns a container with the specified Pulumi version.
func (pmo *PulumiOps) PulumiContainer(ctx context.Context) *Container {
	return dag.Container().
		From(fmt.Sprintf("pulumi/pulumi:%s", pmo.Version))
}

// Login logs into the Pulumi backend using the provided credentials.
func (pmo *PulumiOps) Login(ctx context.Context) *Container {
	credFile := "/opt/credentials.json"
	return pmo.PulumiContainer(ctx).
		WithFile(credFile, pmo.Credentials, ContainerWithFileOpts{Permissions: 0644}).
		WithEnvVariable(
			"GOOGLE_APPLICATION_CREDENTIALS",
			credFile,
			ContainerWithEnvVariableOpts{},
		).
		WithExec([]string{"login", pmo.Backend})
}

// KubeAccess sets up Kubernetes access using the provided kubeconfig file.
func (pmo *PulumiOps) KubeAccess(ctx context.Context) *Container {
	kubeConfigFile := "/opt/kubernetes.yaml"
	return pmo.Login(ctx).
		WithFile(kubeConfigFile, pmo.KubeConfig, ContainerWithFileOpts{Permissions: 0644}).
		WithEnvVariable(
			"KUBECONFIG",
			kubeConfigFile,
			ContainerWithEnvVariableOpts{},
		)
}

// DeployApp deploys a dictycr application using Pulumi configurations and specified parameters.
func (pmo *PulumiOps) DeployApp(
	ctx context.Context,
	// project folder under src that has to be deployed
	// + default="backend_application"
	project string,
	// application that has to be deployed,required
	app string,
	// image tag that has to deployed, required
	tag string,
	// pulumi stack name
	// + default="dev"
	stack string,
) (string, error) {
	opsDir := dag.Gitter().
		WithRef(pulumiOpsBranch).
		WithRepository(pulumiOpsRepo).
		Checkout()
	return pmo.KubeAccess(ctx).
		WithMountedDirectory("/mnt", opsDir).
		WithWorkdir("/mnt").
		WithExec(
			[]string{
				"-C", project,
				"-s", stack,
				"config", "set",
				fmt.Sprintf("%s.tag", app), tag,
				"--path",
			},
		).
		WithExec(
			[]string{
				"-C",
				project,
				"-s",
				stack,
				"up",
				"-y",
				"-r",
				"-f",
				"--non-interactive",
			},
		).
		Stdout(ctx)
}

// DeployAppThroughGithub deploys application using a GitHub deployment ID and token.
func (pmo *PulumiOps) DeployAppThroughGithub(
	ctx context.Context,
	// Deployment ID, Required
	deploymentID string,
	// GitHub token for making API requests, Required
	token string,
) (string, error) {
	return pmo.deployThroughGithub(
		ctx,
		deploymentID,
		token,
		func(container *Container, pload Payload) *Container {
			return container.WithExec(
				[]string{
					"-C", pload.Project,
					"-s", pload.Stack,
					"config", "set",
					fmt.Sprintf(
						"%s.tag",
						pload.Application,
					), pload.DockerImageTag,
					"--path",
				},
			)
		},
	)
}

// DeployBackendThroughGithub deploys the backend using a GitHub deployment ID and token.
func (pmo *PulumiOps) DeployBackendThroughGithub(
	ctx context.Context,
	// Deployment ID, Required
	deploymentID string,
	// GitHub token for making API requests, Required
	token string,
) (string, error) {
	return pmo.deployThroughGithub(
		ctx,
		deploymentID,
		token,
		func(container *Container, pload Payload) *Container {
			return container.WithExec(
				[]string{
					"-C", pload.Project,
					"-s", pload.Stack,
					"config", "set",
					fmt.Sprintf(
						"%s.tag",
						pload.Application,
					), pload.DockerImageTag,
					"--path",
				},
			)
		},
	)
}

// deployThroughGithub is a common method for deploying through GitHub
func (pmo *PulumiOps) deployThroughGithub(
	ctx context.Context,
	deploymentID string,
	token string,
	setConfigFunc func(*Container, Payload) *Container,
) (string, error) {
	var emptyStr string
	owner, repo, err := parseOwnerRepo(pmo.Repository)
	if err != nil {
		return emptyStr, err
	}
	depId, err := strconv.ParseInt(deploymentID, 10, 64)
	if err != nil {
		return emptyStr, fmt.Errorf(
			"error in converting string to int64 %s",
			err,
		)
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
		return emptyStr, fmt.Errorf(
			"error in getting deployment information: %s",
			err,
		)
	}
	var pload Payload
	if err := json.Unmarshal(deployment.Payload, &pload); err != nil {
		return emptyStr, fmt.Errorf("error in decoding payload %s", err)
	}
	if pload.Repository != pmo.Repository {
		return emptyStr, fmt.Errorf(
			"payload repo %s and given repo %s does not match",
			pload.Repository,
			pmo.Repository,
		)
	}
	opsDir := dag.Gitter().
		WithRef(pulumiOpsBranch).
		WithRepository(pulumiOpsRepo).
		Checkout()
	container := pmo.WithKubeConfig(ctx, pmo.KubeConfig).
		KubeAccess(ctx).
		WithMountedDirectory("/mnt", opsDir).
		WithWorkdir("/mnt")

	container = setConfigFunc(container, pload)

	return container.WithExec(
		[]string{
			"-C",
			pload.Project,
			"-s",
			pload.Stack,
			"up",
			"-y",
			"-r",
			"-f",
			"--non-interactive",
		},
	).Stdout(ctx)
}

// DeployFrontendThroughGithub deploys the frontend using a GitHub deployment ID and token.
func (pmo *PulumiOps) DeployFrontendThroughGithub(
	ctx context.Context,
	// Deployment ID, Required
	deploymentID string,
	// GitHub token for making API requests, Required
	token string,
) (string, error) {
	return pmo.deployThroughGithub(
		ctx,
		deploymentID,
		token,
		func(container *Container, pload Payload) *Container {
			return container.WithExec(
				[]string{
					"-C", pload.Project,
					"-s", pload.Stack,
					"config", "set",
					"frontpage.tag", pload.DockerImageTag,
					"--path",
				},
			).WithExec(
				[]string{
					"-C", pload.Project,
					"-s", pload.Stack,
					"config", "set",
					"publication.tag", pload.DockerImageTag,
					"--path",
				},
			).WithExec(
				[]string{
					"-C", pload.Project,
					"-s", pload.Stack,
					"config", "set",
					"stockcenter.tag", pload.DockerImageTag,
					"--path",
				},
			)
		},
	)
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
