package git

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

const (
	// LayerNameGit is the name of the layer that is used to store git environment variables.
	LayerNameGit = "git"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) (err error)
}

//go:generate faux --interface CredentialManager --output fakes/credential_manager.go
type CredentialManager interface {
	Setup(workingDir, platformPath string) (err error)
}

func Build(executable Executable, credentialManager CredentialManager, logger scribe.Emitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		layer, err := context.Layers.Get(LayerNameGit)
		if err != nil {
			return packit.BuildResult{}, err
		}

		layer.Launch = true
		layer.Build = true

		exist, err := fs.Exists(filepath.Join(context.WorkingDir, ".git"))
		if err != nil {
			return packit.BuildResult{}, err
		}

		var buildResult packit.BuildResult
		if exist {
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

			layer.SharedEnv.Default("REVISION", revision)

			logger.EnvironmentVariables(layer)

			buildResult = packit.BuildResult{
				Layers: []packit.Layer{layer},
				Launch: packit.LaunchMetadata{
					Labels: map[string]string{
						"org.opencontainers.image.revision": revision,
					},
				},
			}
		}

		err = credentialManager.Setup(context.WorkingDir, context.Platform.Path)
		if err != nil {
			return packit.BuildResult{}, fmt.Errorf("failed to configure given credentials: %w", err)
		}

		return buildResult, nil
	}
}
