package main

import (
	"context"
	"fmt"
)

// PulumiOps represents the Pulumi operations configuration.
type PulumiOps struct {
	Backend     string
	Version     string
	KubeConfig  *File
	Credentials *File
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

// DeployBackend deploys a backend application using Pulumi configurations and specified parameters.
func (pmo *PulumiOps) DeployBackend(
	ctx context.Context,
	// project/folder containing pulumi configurations for deploying,
	// required
	src *Directory,
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
	return pmo.KubeAccess(ctx).
		WithMountedDirectory("/mnt", src).
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
				"-C", project, "-s",
				stack, "up", "-y", "-r", "-f", "--non-interactive",
			},
		).Stdout(ctx)
}
