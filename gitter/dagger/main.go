// Package main provides a Gitter struct to manipulate git repositories,
// including setting repository details and performing actions like checkout and inspect.
package main

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

var bre = regexp.MustCompile(`refs/heads/(.+)`)

type Gitter struct {
	// Repository name
	Repository string
	// Git reference
	Ref string
	// Repository path
	Path string
}

// WithRef sets the Git reference (branch, tag, or SHA)
func (gcmd *Gitter) WithRef(
	// the branch, tag or sha to checkout, Required.
	ref string,
) (*Gitter, error) {
	if len(ref) == 0 {
		return gcmd, errors.New("ref value is required")
	}
	gcmd.Ref = ref
	return gcmd, nil
}

// WithRepository sets the GitHub repository name
func (gcmd *Gitter) WithRepository(
	// github repository name with owner, for example tora/bora, Required
	repository string,
) (*Gitter, error) {
	if len(repository) == 0 {
		return gcmd, errors.New("repository value is required")
	}
	gcmd.Repository = repository
	return gcmd, nil
}

// Checkout clones the repository and checks out the specific ref
func (gcmd *Gitter) Checkout(ctx context.Context) *Directory {
	return dag.Git().
		Clone(gcmd.Repository).
		Checkout(parseBranchName(gcmd.Ref)).
		Directory()
}

// Inspect clones the given repository and returns a Terminal instance for inspection
func (gcmd *Gitter) Inspect(ctx context.Context) *Terminal {
	return dag.Git().
		Clone(gcmd.Repository).
		Checkout(parseBranchName(gcmd.Ref)).
		Inspect()
}

func parseBranchName(ref string) string {
	prefix := "refs/heads/"
	if !strings.HasPrefix(ref, prefix) {
		return ref
	}
	parts := strings.SplitN(ref, "refs/heads/", 2)
	return parts[1]
}
