package main

import (
	"os"

	git "github.com/paketo-buildpacks/git"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/pexec"
)

func main() {
	logEmitter := git.NewLogEmitter(os.Stdout)

	packit.Run(
		git.Detect(),
		git.Build(
			pexec.NewExecutable("git"),
			logEmitter,
		),
	)
}
