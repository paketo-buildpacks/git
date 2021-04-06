package git

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
)

const (
	// LayerNameGit is the name of the layer that is used to store git environment variables.
	LayerNameGit = "git"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) (err error)
}

func Build(executable Executable, logger scribe.Logger) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		varsLayer, err := context.Layers.Get(LayerNameGit)
		if err != nil {
			return packit.BuildResult{}, err
		}

		buffer := bytes.NewBuffer(nil)
		args := []string{"rev-parse", "HEAD"}
		err = executable.Execute(pexec.Execution{
			Args:   args,
			Dir:    context.WorkingDir,
			Stdout: buffer,
			Stderr: buffer,
		})
		if err != nil {
			logger.Detail(buffer.String())
			return packit.BuildResult{}, fmt.Errorf("failed to execute 'git rev-parse HEAD': %w", err)
		}

		revision := strings.TrimSpace(buffer.String())

		varsLayer.Launch = true
		varsLayer.Build = true
		varsLayer.SharedEnv.Default("REVISION", revision)

		logger.Process("Configuring shared environment")
		logger.Subprocess("%s", scribe.NewFormattedMapFromEnvironment(varsLayer.SharedEnv))
		logger.Break()

		return packit.BuildResult{
			Layers: []packit.Layer{varsLayer},
		}, nil
	}
}
