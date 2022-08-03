package main

import (
	"os"

	"github.com/paketo-buildpacks/git"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"
)

func main() {
	executable := pexec.NewExecutable("git")
	emitter := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	bindingResolver := servicebindings.NewResolver()

	packit.Run(
		git.Detect(bindingResolver),
		git.Build(
			executable,
			git.NewGitCredentialManager(bindingResolver, executable, emitter),
			emitter,
		),
	)
}
