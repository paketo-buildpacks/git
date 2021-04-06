package git

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		_, err := os.Stat(filepath.Join(context.WorkingDir, ".git"))
		if err != nil {
			return packit.DetectResult{}, packit.Fail.WithMessage("failed to find .git directory")
		}
		return packit.DetectResult{}, nil
	}
}
