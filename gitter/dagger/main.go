// Package main provides a Gitter struct to manipulate git repositories,
// including setting repository details and performing actions like checkout.
package main

import (
	"context"
	"errors"
)

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

// Checkout clones the repository and checks out the specific ref.
// Uses Dagger's native git support — no container build required.
func (gcmd *Gitter) Checkout(ctx context.Context) *Directory {
	return dag.Git(gcmd.Repository).Ref(gcmd.Ref).Tree()
}

// CommitHash retrieves the short commit hash at the specified ref.
func (gcmd *Gitter) CommitHash(ctx context.Context) (string, error) {
	sha, err := dag.Git(gcmd.Repository).Ref(gcmd.Ref).Commit(ctx)
	if err != nil {
		return "", err
	}
	if len(sha) >= 7 {
		return sha[:7], nil
	}
	return sha, nil
}
