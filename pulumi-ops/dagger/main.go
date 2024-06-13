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
