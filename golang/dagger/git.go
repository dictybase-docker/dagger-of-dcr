package main

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GitReference open a git repository located at the specified path,
// returning the current HEAD reference.
func GitReference(path string) (*plumbing.Reference, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			return nil, fmt.Errorf("given folder %s is not a git repo", path)
		}
		return nil, fmt.Errorf("unknown git repo open error %q", err)
	}
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("error in getting head %q", err)
	}
	return head, nil
}
