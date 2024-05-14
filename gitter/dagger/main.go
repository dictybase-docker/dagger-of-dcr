// A generated module for Gitter functions

package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Gitter struct {
	// Repository name
	Repository string
	// Git reference
	Ref string
	// Repository path
	Path string
}

// WithPath sets the path where the repository should be placed.
func (gcmd *Gitter) WithPath(
	// path to place the repository
	// +optional
	// +default="current working directory"
	path string,
) (*Gitter, error) {
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

// Checkout clones a repository at the specified path and reference
func (gcmd *Gitter) Checkout(ctx context.Context) (*Directory, error) {
	_, err := cloneRepo(gcmd.Path, gcmd.Ref, gcmd.Repository)
	if err != nil {
		return nil, err
	}
	return dag.Directory().Directory(gcmd.Path), nil
}

func cloneRepo(path, ref, repository string) (*git.Repository, error) {
	repo, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:           repository,
		ReferenceName: plumbing.ReferenceName(ref),
	})
	if err == nil {
		return repo, nil
	}
	if !errors.Is(err, plumbing.ErrReferenceNotFound) {
		return nil, fmt.Errorf("error in checking out repo %q", err)
	}
	return cloneDefaultBranch(path, repository, ref)
}

func cloneDefaultBranch(path, repository, ref string) (*git.Repository, error) {
	repo, err := git.PlainClone(
		path,
		false,
		&git.CloneOptions{URL: repository},
	)
	if err != nil {
		return nil, fmt.Errorf("error in checking out default branch %q", err)
	}
	return checkoutRef(repo, repository, ref)
}

func checkoutRef(
	repo *git.Repository,
	repository, ref string,
) (*git.Repository, error) {
	wtree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf(
			"error in getting worktree from default branch %q",
			err,
		)
	}
	err = wtree.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(ref)})
	if err != nil {
		return nil, fmt.Errorf("error in checking out ref %s %q", ref, err)
	}
	return repo, nil
}
