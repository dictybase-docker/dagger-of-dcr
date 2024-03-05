package main

import (
	F "github.com/IBM/fp-go/function"
)

var (
	base               = F.Curry2(unCurriedBase)
	wolfiWithGoInstall = F.Curry2(unCurriedwolfiWithGoInstall)
	prepareWorkspace   = F.Bind12of3(uncurriedPrepareWorkspace)
	goTestRunner       = F.Curry2(uncurriedGoTestRunner)
	goLintRunner       = F.Curry2(uncurriedGoLintRunner)
)

func unCurriedBase(base string, ctr *Container) *Container {
	return ctr.From(base)

}

func unCurriedwolfiWithGoInstall(version string, ctr *Container) *Container {
	return ctr.WithExec([]string{"apk", "update"}).
		WithExec([]string{"apk", "add", version})
}

func uncurriedGoTestRunner(cmd []string, ctr *Container) *Container {
	return ctr.WithExec(append([]string{"go", "test", "./..."}, cmd...))
}

func uncurriedGoLintRunner(cmd []string, ctr *Container) *Container {
	return ctr.WithExec(
		append([]string{"golangci-lint", "run", "-c", ".golangci.yml"}, cmd...),
	)
}

func modCache(ctr *Container) *Container {
	return ctr.WithExec([]string{"go", "mod", "download"})

}

func uncurriedPrepareWorkspace(
	src *Directory,
	mount string,
	ctr *Container,
) *Container {
	return ctr.WithMountedDirectory(mount, src).
		WithWorkdir(mount)
}
