package git

import (
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
)

func Detect(bindingResolver BindingResolver) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		exist, err := fs.Exists(filepath.Join(context.WorkingDir, ".git"))
		if err != nil {
			return packit.DetectResult{}, err
		}

		bindings, err := bindingResolver.Resolve("git-credentials", "", context.Platform.Path)
		if err != nil {
			return packit.DetectResult{}, err
		}

		if !exist && len(bindings) == 0 {
			return packit.DetectResult{}, packit.Fail.WithMessage("failed to find .git directory and no git credential service bindings present")
		}

		return packit.DetectResult{}, nil
	}
}
