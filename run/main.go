package main

import (
	"os"

	git "github.com/paketo-buildpacks/git"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {

	packit.Run(
		git.Detect(),
		git.Build(
			pexec.NewExecutable("git"),
			scribe.NewLogger(os.Stdout),
		),
	)
}
