package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"
	"github.com/paketo-community/git"
)

func main() {
	executable := pexec.NewExecutable("git")
	emitter := scribe.NewEmitter(os.Stdout)
	bindingResolver := servicebindings.NewResolver()

	packit.Run(
		git.Detect(),
		git.Build(
			executable,
			git.NewGitCredentialManager(bindingResolver, executable, emitter),
			emitter,
		),
	)
}
