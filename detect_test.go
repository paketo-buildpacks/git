package git_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"

	"github.com/paketo-buildpacks/git"
	"github.com/paketo-buildpacks/git/fakes"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir      string
		bindingResolver *fakes.BindingResolver

		detect packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		bindingResolver = &fakes.BindingResolver{}

		detect = git.Detect(bindingResolver)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when a .git directory is present", func() {
		it.Before(func() {
			err := os.Mkdir(filepath.Join(workingDir, ".git"), os.ModeDir)
			Expect(err).NotTo(HaveOccurred())
		})

		it("detects", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
				Platform:   packit.Platform{Path: "some-platform"},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{}))

			Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal("some-platform"))
			Expect(bindingResolver.ResolveCall.Receives.Typ).To(Equal("git-credentials"))
		})
	})

	context("when a .git directory is not present", func() {
		context("when in a submodule (.git is a file, not dir)", func() {
			it.Before(func() {
				err := os.WriteFile(filepath.Join(workingDir, ".git"), nil, os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			it("fails detections", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
					Platform:   packit.Platform{Path: "some-platform"},
				})
				Expect(err).To(MatchError(packit.Fail.WithMessage("failed to find .git directory and no git credential service bindings present")))
			})
		})

		context("when there are no git-credentials service bindings", func() {
			it("fails detections", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
					Platform:   packit.Platform{Path: "some-platform"},
				})
				Expect(err).To(MatchError(packit.Fail.WithMessage("failed to find .git directory and no git credential service bindings present")))
			})
		})

		context("when there are git-credentials service bindings", func() {
			it.Before(func() {
				bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
					{
						Path: "some-path",
					},
				}
			})
			it("detects", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
					Platform:   packit.Platform{Path: "some-platform"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{}))

				Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal("some-platform"))
				Expect(bindingResolver.ResolveCall.Receives.Typ).To(Equal("git-credentials"))
			})
		})
	})

	context("failure cases", func() {
		context("exists check fails", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})
			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})
			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
					Platform:   packit.Platform{Path: "some-platform"},
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when binding resolution fails", func() {
			it.Before(func() {
				bindingResolver.ResolveCall.Returns.Error = errors.New("failed to resolve bindings")
			})
			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
					Platform:   packit.Platform{Path: "some-platform"},
				})
				Expect(err).To(MatchError("failed to resolve bindings"))
			})
		})
	})
}
