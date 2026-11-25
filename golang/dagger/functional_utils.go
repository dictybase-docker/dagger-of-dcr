package main

import (
	"dagger/golang/internal/dagger"

	F "github.com/IBM/fp-go/function"
)

var (
	base             = F.Curry2(unCurriedBase)
	prepareWorkspace = F.Bind12of3(uncurriedPrepareWorkspace)
	goLintRunner     = F.Curry2(uncurriedGoLintRunner)
)

func unCurriedBase(base string, ctr *dagger.Container) *dagger.Container {
	return ctr.From(base)
}

func uncurriedGoLintRunner(cmd []string, ctr *dagger.Container) *dagger.Container {
	return ctr.WithExec(
		append([]string{"golangci-lint", "run", "-c", ".golangci.yml"}, cmd...),
	)
}

func uncurriedPrepareWorkspace(
	src *dagger.Directory,
	mount string,
	ctr *dagger.Container,
) *dagger.Container {
	return ctr.WithMountedDirectory(mount, src).
		WithWorkdir(mount)
}
