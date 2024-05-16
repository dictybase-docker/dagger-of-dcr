// Package main provides a Gitter struct to manipulate git repositories,
// including setting repository details and performing actions like checkout and inspect.
package main

import (
	"context"
	"errors"
	"regexp"
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
		Checkout(gcmd.ParseRef(ctx)).
		Directory()
}

// CommitHash retrieves the short commit hash of the HEAD from the specified Git repository.
func (gcmd *Gitter) CommitHash(ctx context.Context) (string, error) {
	return dag.Git().
		Clone(gcmd.Repository).
		Checkout(gcmd.ParseRef(ctx)).
		Command([]string{"rev-parse", "--short", "HEAD"}).Stdout(ctx)
}

// Inspect clones the given repository and returns a Terminal instance for inspection
func (gcmd *Gitter) Inspect(ctx context.Context) *Terminal {
	return dag.Git().
		Clone(gcmd.Repository).
		Checkout(gcmd.ParseRef(ctx)).
		Inspect()
}

// ParseRef extracts the branch name from a Git reference string or returns the original reference if no match is found.
func (gcmd *Gitter) ParseRef(ctx context.Context) string {
	match := bre.FindStringSubmatch(gcmd.Ref)
	if len(match) > 1 {
		return match[1]
	}
	return gcmd.Ref
}
