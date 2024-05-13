// Package provides a Git command utility that allows cloning repositories,
// setting specific repository paths, branches, tags, or SHAs for checkout.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type GitCmd struct {
	Repository string
	Ref        string
	Path       string
	GitRef     *git.Repository
}

// WithPath sets the path where the repository should be placed. If no path is
// provided (i.e., an empty string), it defaults to the current working
// directory.
func (gcmd *GitCmd) WithPath(
	// path to place the repository
	// +optional
	// +default="current working directory"
	path string,
) (*GitCmd, error) {
	if len(path) == 0 {
		curr, err := os.Getwd()
		if err != nil {
			return gcmd, fmt.Errorf(
				"error in getting current working dir %q",
				err,
			)
		}
		path = curr
	}
	gcmd.Path = path
	return gcmd, nil
}

// WithRef sets the Git reference (branch, tag, or SHA). This method requires a
// non-empty string as the ref parameter, indicating the specific reference to
// be checked out.
func (gcmd *GitCmd) WithRef(
	// the branch, tag or sha to checkout, Required.
	ref string,
) (*GitCmd, error) {
	if len(ref) == 0 {
		return gcmd, errors.New("ref value is required")
	}
	gcmd.Ref = ref
	return gcmd, nil
}

// WithRepository sets the GitHub repository name for the GitCmd and returns a
// modified GitCmd object and an error if the repository name is empty.
func (gcmd *GitCmd) WithRepository(
	// github repository name with owner, for example tora/bora, Required
	repository string,
) (*GitCmd, error) {
	if len(repository) == 0 {
		return gcmd, errors.New("repository value is required")
	}
	gcmd.Repository = repository
	return gcmd, nil
}

// Checkout clones a repository at the specified path and reference, falling
// back to the default branch if the reference is not found.
func (gcmd *GitCmd) Checkout(
	ctx context.Context,
) (*GitCmd, error) {
	repo, err := git.PlainClone(
		gcmd.Path,
		false,
		&git.CloneOptions{
			URL:           gcmd.Repository,
			ReferenceName: plumbing.ReferenceName(gcmd.Ref),
		})
	if err == nil {
		gcmd.GitRef = repo
		return gcmd, nil
	}
	if !errors.Is(err, plumbing.ErrReferenceNotFound) {
		return nil, fmt.Errorf("error in checking out repo %q", err)
	}
	hrepo, err := git.PlainClone(
		gcmd.Path,
		false,
		&git.CloneOptions{URL: gcmd.Repository},
	)
	if err != nil {
		return nil, fmt.Errorf("error in checking out default branch %q", err)
	}
	wtree, err := hrepo.Worktree()
	if err != nil {
		return nil, fmt.Errorf(
			"error in getting worktree from default branch %q",
			err,
		)
	}
	err = wtree.Checkout(
		&git.CheckoutOptions{Hash: plumbing.NewHash(gcmd.Ref)},
	)
	if err != nil {
		return nil, fmt.Errorf("error in checking out ref %s %q", gcmd.Ref, err)
	}
	gcmd.GitRef = repo
	return gcmd, nil
}
