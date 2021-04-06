package main

import (
	"github.com/paketo-buildpacks/packit"
	git "github.com/paketo-buildpacks/git"
)

func main() {
	packit.Run(git.Detect(), git.Build())
}
