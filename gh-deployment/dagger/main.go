// A generated module for GhDeployment functions
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
	"errors"
)

type GhDeployment struct {
	// Repository name
	Repository string
	// Git reference
	Ref            string
	DaggerVersion  string
	DaggerChecksum string
	Cluster        string
	Storage        string
	KubeConfig     string
	Artifact       string
	ImageTag       string
	Application    string
	Stack          string
	RunId          int
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

// WithApplication sets the application name
func (ghd *GhDeployment) WithApplication(
	// Application name, Required
	application string,
) (*GhDeployment, error) {
	if len(application) == 0 {
		return ghd, errors.New("application value is required")
	}
	ghd.Application = application
	return ghd, nil
}

// WithStack sets the stack name
func (ghd *GhDeployment) WithStack(
	// Stack name, Required
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
