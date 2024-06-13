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
