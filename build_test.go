package git_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/paketo-community/git"
	"github.com/paketo-community/git/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		workingDir string
		cnbDir     string

		executable *fakes.Executable
		buffer     *bytes.Buffer

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewEmitter(buffer)

		executable = &fakes.Executable{}
		executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
			fmt.Fprintf(execution.Stdout, "sha123456789")
			return nil
		}

		build = git.Build(executable, logger)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a result that builds correctly", func() {
		result, err := build(packit.BuildContext{
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			Stack:      "some-stack",
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{},
			},
			Layers: packit.Layers{Path: layersDir},
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(Equal(packit.BuildResult{
			Plan: packit.BuildpackPlan{Entries: nil},
			Layers: []packit.Layer{
				{
					Name:             "git",
					Path:             filepath.Join(layersDir, "git"),
					Build:            true,
					Launch:           true,
					LaunchEnv:        packit.Environment{},
					SharedEnv:        map[string]string{"REVISION.default": "sha123456789"},
					BuildEnv:         packit.Environment{},
					ProcessLaunchEnv: map[string]packit.Environment{},
				},
			},
			Launch: packit.LaunchMetadata{
				Labels: map[string]string{
					"org.opencontainers.image.revision": "sha123456789",
				},
			},
		}))

		Expect(buffer).To(ContainLines(
			"Some Buildpack some-version",
			"  Configuring build environment",
			`    REVISION -> "sha123456789"`,
			"",
			"  Configuring launch environment",
			`    REVISION -> "sha123456789"`,
			"",
		))
	})

	context("when the executable fails", func() {
		it.Before(func() {
			executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
				return errors.New("some-error")
			}
		})
		it("returns the error", func() {
			_, err := build(packit.BuildContext{})
			Expect(err).To(MatchError("failed to execute 'git rev-parse HEAD': some-error"))
		})
	})
}
