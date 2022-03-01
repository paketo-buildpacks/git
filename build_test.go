package git_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
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

		executable        *fakes.Executable
		credentialManager *fakes.CredentialManager

		buffer *bytes.Buffer

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewEmitter(buffer)

		executable = &fakes.Executable{}
		executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
			fmt.Fprintf(execution.Stdout, "sha123456789")
			return nil
		}

		credentialManager = &fakes.CredentialManager{}

		build = git.Build(executable, credentialManager, logger)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when there is a .git directory in the workingDir", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(workingDir, ".git"), os.ModePerm)).To(Succeed())
		})
		it("returns a result that builds correctly", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				Platform:   packit.Platform{Path: "some-platform"},
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("git"))
			Expect(layer.Path).To(Equal(filepath.Join(layersDir, "git")))
			Expect(layer.Build).To(BeTrue())
			Expect(layer.Launch).To(BeTrue())
			Expect(layer.SharedEnv).To(Equal(packit.Environment{"REVISION.default": "sha123456789"}))

			Expect(result.Launch).To(Equal(packit.LaunchMetadata{
				Labels: map[string]string{
					"org.opencontainers.image.revision": "sha123456789",
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

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"rev-parse", "HEAD"}))

			Expect(credentialManager.SetupCall.Receives.PlatformPath).To(Equal("some-platform"))
			Expect(credentialManager.SetupCall.Receives.WorkingDir).To(Equal(workingDir))
		})
	})

	context("when there is not a .git directory in the workingDir", func() {
		it("returns a result that builds correctly", func() {
			result, err := build(packit.BuildContext{
				WorkingDir: workingDir,
				Platform:   packit.Platform{Path: "some-platform"},
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Layers: packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(0))

			Expect(buffer).NotTo(ContainLines(
				"Some Buildpack some-version",
				"  Configuring build environment",
				`    REVISION -> "sha123456789"`,
				"",
				"  Configuring launch environment",
				`    REVISION -> "sha123456789"`,
				"",
			))

			Expect(executable.ExecuteCall.CallCount).To(Equal(0))

			Expect(credentialManager.SetupCall.Receives.PlatformPath).To(Equal("some-platform"))
			Expect(credentialManager.SetupCall.Receives.WorkingDir).To(Equal(workingDir))
		})
	})

	context("failure cases", func() {
		context("when the exists check fails", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})
			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})
			it("returns the error", func() {
				_, err := build(packit.BuildContext{
					WorkingDir: workingDir,
					Platform:   packit.Platform{Path: "some-platform"},
					Layers:     packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
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

		context("when the credential setup fails", func() {
			it.Before(func() {
				credentialManager.SetupCall.Returns.Err = errors.New("setup failed")
			})
			it("returns the error", func() {
				_, err := build(packit.BuildContext{})
				Expect(err).To(MatchError("failed to configure given credentials: setup failed"))
			})

		})
	})
}
