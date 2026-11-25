package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"dagger/pulumi-ops/internal/dagger"

	"github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"
)

const (
	pulumiOpsRepo = "https://github.com/dictybase-docker/cluster-ops.git"
	githubURL     = "https://github.com"
	pulumiConfig  = "config"
	pulumiPath    = "--path"
	pulumiSet     = "set"
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
	KubeConfig  *dagger.File
	Credentials *dagger.File
	// Repository name
	Repository string
}

// WithPulumi sets the Pulumi version for the PulumiOps instance.
func (pmo *PulumiOps) WithPulumi(
	// pulumi version
	// + default="3.108.0"
	version string,
) *PulumiOps {
	pmo.Version = version
	return pmo
}

// WithBackend sets the backend for storing state for the PulumiOps instance.
func (pmo *PulumiOps) WithBackend(
	// pulumi backend for storing state, required
	backend string,
) *PulumiOps {
	pmo.Backend = backend
	return pmo
}

// WithCredentials sets the credentials file for the PulumiOps instance.
func (pmo *PulumiOps) WithCredentials(
	// credentials file(google cloud for the time being), required
	credentials *dagger.File,
) *PulumiOps {
	pmo.Credentials = credentials
	return pmo
}

// WithKubeConfig sets the Kubernetes configuration file for the PulumiOps instance.
func (pmo *PulumiOps) WithKubeConfig(
	// kubernetes configuration file, required
	config *dagger.File,
) *PulumiOps {
	pmo.KubeConfig = config
	return pmo
}

// WithRepository sets the repository name
func (pmo *PulumiOps) WithRepository(
	// Repository name, Required
	repository string,
) *PulumiOps {
	pmo.Repository = repository
	return pmo
}

// PulumiContainer returns a container with the specified Pulumi version.
func (pmo *PulumiOps) PulumiContainer() *dagger.Container {
	return dag.Container().
		From(fmt.Sprintf("pulumi/pulumi:%s", pmo.Version))
}

// Login logs into the Pulumi backend using the provided credentials.
func (pmo *PulumiOps) Login() *dagger.Container {
	credFile := "/opt/credentials.json"
	return pmo.PulumiContainer().
		WithFile(credFile, pmo.Credentials, dagger.ContainerWithFileOpts{Permissions: 0o644}).
		WithEnvVariable(
			"GOOGLE_APPLICATION_CREDENTIALS",
			credFile,
			dagger.ContainerWithEnvVariableOpts{},
		).
		WithExec([]string{"login", pmo.Backend})
}

// KubeAccess sets up Kubernetes access using the provided kubeconfig file.
func (pmo *PulumiOps) KubeAccess() *dagger.Container {
	kubeConfigFile := "/opt/kubernetes.yaml"
	return pmo.Login().
		WithFile(kubeConfigFile, pmo.KubeConfig, dagger.ContainerWithFileOpts{Permissions: 0o644}).
		WithEnvVariable(
			"KUBECONFIG",
			kubeConfigFile,
			dagger.ContainerWithEnvVariableOpts{},
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
	// git ref of the cluster-ops repository to clone
	// + default="develop"
	clusterOpsRef string,
) (string, error) {
	opsDir := dag.Gitter().
		WithRef(clusterOpsRef).
		WithRepository(pulumiOpsRepo).
		Checkout()
	return pmo.KubeAccess().
		WithMountedDirectory("/mnt", opsDir).
		WithWorkdir("/mnt").
		WithExec(
			[]string{
				"-C", project,
				"-s", stack,
				pulumiConfig, pulumiSet,
				fmt.Sprintf("%s.tag", app), tag,
				pulumiPath,
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
	// git ref of the cluster-ops repository to clone
	// + default="develop"
	clusterOpsRef string,
) (string, error) {
	return pmo.deployThroughGithub(
		ctx,
		deploymentID,
		token,
		clusterOpsRef,
		func(container *dagger.Container, pload Payload) *dagger.Container {
			return container.WithExec(
				[]string{
					"-C", pload.Project,
					"-s", pload.Stack,
					pulumiConfig, pulumiSet,
					fmt.Sprintf(
						"%s.tag",
						pload.Application,
					), pload.DockerImageTag,
					pulumiPath,
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
	// git ref of the cluster-ops repository to clone
	// + default="develop"
	clusterOpsRef string,
) (string, error) {
	return pmo.deployThroughGithub(
		ctx,
		deploymentID,
		token,
		clusterOpsRef,
		func(container *dagger.Container, pload Payload) *dagger.Container {
			return container.WithExec(
				[]string{
					"-C", pload.Project,
					"-s", pload.Stack,
					pulumiConfig, pulumiSet, pulumiPath,
					"properties.image.tag",
					pload.DockerImageTag,
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
	clusterOpsRef string,
	setConfigFunc func(*dagger.Container, Payload) *dagger.Container,
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
		WithRef(clusterOpsRef).
		WithRepository(pulumiOpsRepo).
		Checkout()
	container := pmo.WithKubeConfig(pmo.KubeConfig).
		KubeAccess().
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
			"-f",
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
	// git ref of the cluster-ops repository to clone
	// + default="develop"
	clusterOpsRef string,
) (string, error) {
	return pmo.deployThroughGithub(
		ctx,
		deploymentID,
		token,
		clusterOpsRef,
		func(container *dagger.Container, pload Payload) *dagger.Container {
			execCmd := []string{
				"-C", pload.Project,
				"-s", pload.Stack,
				pulumiConfig, "set-all", pulumiPath,
			}
			for idx := range strings.Split(pload.Application, ":") {
				execCmd = append(
					execCmd,
					"--plaintext",
					fmt.Sprintf(
						"'properties.apps[%d].tag'=%s",
						idx,
						pload.DockerImageTag,
					),
				)
			}
			return container.WithExec(execCmd)
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
