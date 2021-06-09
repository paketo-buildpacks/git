package main

import (
	"os"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/paketo-community/git"
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
