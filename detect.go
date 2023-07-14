package git

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
)

func Detect(bindingResolver BindingResolver) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		exist, err := gitDirExists(context.WorkingDir)
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

func gitDirExists(workingDir string) (bool, error) {
	info, err := os.Stat(filepath.Join(workingDir, ".git"))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}
