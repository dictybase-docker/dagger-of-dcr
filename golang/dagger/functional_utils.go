package main

import (
	F "github.com/IBM/fp-go/function"
)

var (
	base               = F.Curry2(unCurriedBase)
	wolfiWithGoInstall = F.Curry2(unCurriedwolfiWithGoInstall)
	prepareWorkspace   = F.Bind12of3(uncurriedPrepareWorkspace)
	goTestRunner       = F.Curry2(uncurriedGoTestRunner)
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

func uncurriedPrepareWorkspace(
	src *Directory,
	mount string,
	ctr *Container,
) *Container {
	return ctr.WithMountedDirectory(mount, src).
		WithWorkdir(mount).
		WithExec([]string{"go", "mod", "download"})
}
