/*
Package main provides functionality to set up and manage Kubernetes clusters
using Kops and kubectl binaries. It includes methods to configure the Kops
environment, export kubeconfig files, and set various parameters such as cluster
name, state storage, and credentials.
*/
package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
)

type Kops struct {
	Kubectl     string
	KopsV       string
	Cluster     string
	Storage     string
	Name        string
	Credentials *File
}

// KopsContainer sets up a container with specified versions of kubectl and kops binaries from given URLs
func (kmg *Kops) KopsContainer(ctx context.Context) *Container {
	kubectlBinary := fmt.Sprintf(
		"https://storage.googleapis.com/kubernetes-release/release/v%s/bin/linux/%s/kubectl",
		kmg.Kubectl,
		"amd64",
	)
	kopsBinary := fmt.Sprintf(
		"https://github.com/kubernetes/kops/releases/download/v%s/kops-linux-amd64",
		kmg.KopsV,
	)
	return dag.Container().
		From("alpine:3.20.0").
		WithFile(
			"/usr/bin/kubectl",
			dag.HTTP(kubectlBinary),
			ContainerWithFileOpts{Permissions: 0755},
		).
		WithFile(
			"/usr/bin/kops",
			dag.HTTP(kopsBinary),
			ContainerWithFileOpts{Permissions: 0755},
		)
}

// ExportKubectl exports the kubeconfig file for the specified Kops cluster to a specified output path
func (kmg *Kops) ExportKubectl(ctx context.Context) (*File, error) {
	credFile := "/opt/credentials.json"
	outPath := "/work/out"
	uuid.EnableRandPool()
	rndId, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("error in generating random uuid %s", err)
	}
	outFile := filepath.Join(outPath, fmt.Sprintf("%s.yaml", rndId.String()))
	cmd := []string{
		"kops",
		"export",
		"kubeconfig",
		"--name",
		kmg.Cluster,
		"--state",
		kmg.Storage,
		"--kubeconfig",
		outFile,
		"--admin",
	}
	return kmg.KopsContainer(ctx).
		WithFile(credFile, kmg.Credentials, ContainerWithFileOpts{Permissions: 0644}).
		WithEnvVariable(
			"GOOGLE_APPLICATION_CREDENTIALS",
			credFile,
			ContainerWithEnvVariableOpts{},
		).WithExec(cmd).File(outFile), nil
}

// WithName sets the name of the kubectl output file
func (kmg *Kops) WithName(
	ctx context.Context,
	// name of the kubectl output file
	// + default="kubefromkops.yaml"
	name string,
) *Kops {
	kmg.Name = name
	return kmg
}

// WithCredentials sets the credentials file
func (kmg *Kops) WithCredentials(
	ctx context.Context,
	// credentials file(google cloud for the time being), required
	credentials *File,
) *Kops {
	kmg.Credentials = credentials
	return kmg
}

// WithStateStorage sets the location of the state storage
func (kmg *Kops) WithStateStorage(
	ctx context.Context,
	// location of state storeage, required
	storage string,
) *Kops {
	kmg.Storage = storage
	return kmg
}

// WithCluster sets the cluster name
func (kmg *Kops) WithCluster(
	ctx context.Context,
	// cluster name, required
	name string,
) *Kops {
	kmg.Cluster = name
	return kmg
}

// WithKops sets the Kops version
func (kmg *Kops) WithKops(
	ctx context.Context,
	// kops version
	// + default="1.27.0"
	version string,
) *Kops {
	kmg.KopsV = version
	return kmg
}

// WithKubectl sets the kubectl version
func (kmg *Kops) WithKubectl(
	ctx context.Context,
	// kubectl version
	// + default="1.24.11"
	version string,
) *Kops {
	kmg.Kubectl = version
	return kmg
}
