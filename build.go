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
	LayerNameGitVars = "git"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) (err error)
}

func Build(executable Executable, logs LogEmitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logs.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		varsLayer, err := context.Layers.Get(LayerNameGitVars)
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
			logs.Detail(buffer.String())
			return packit.BuildResult{}, fmt.Errorf("failed to execute 'git rev-parse HEAD': %w", err)
		}

		revision := strings.TrimSpace(buffer.String())

		varsLayer.Launch = true
		varsLayer.LaunchEnv.Default("REVISION", revision)
		logs.Process("Configuring launch environment")
		logs.Subprocess("%s", scribe.NewFormattedMapFromEnvironment(varsLayer.LaunchEnv))
		logs.Break()

		return packit.BuildResult{
			Layers: []packit.Layer{varsLayer},
		}, nil
	}
}
