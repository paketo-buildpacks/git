package git_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"
	"github.com/paketo-community/git"
	"github.com/paketo-community/git/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testGitCredentialManager(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		bindingResolver *fakes.BindingResolver
		executable      *fakes.Executable

		executions []pexec.Execution

		buffer      *bytes.Buffer
		platformDir string

		gitCredentialManager git.GitCredentialManager
	)

	it.Before(func() {
		var err error

		bindingResolver = &fakes.BindingResolver{}

		executable = &fakes.Executable{}
		executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
			executions = append(executions, execution)
			return nil
		}

		buffer = bytes.NewBuffer(nil)
		platformDir, err = os.MkdirTemp("", "platform")
		Expect(err).NotTo(HaveOccurred())

		gitCredentialManager = git.NewGitCredentialManager(bindingResolver, executable, scribe.NewEmitter(buffer))
	})

	it.After(func() {
		Expect(os.RemoveAll(platformDir)).To(Succeed())
	})

	context("Setup", func() {
		context("when there are no bound credentials", func() {
			it("run no configuration commands", func() {
				err := gitCredentialManager.Setup("working-dir", platformDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal(platformDir))
				Expect(executable.ExecuteCall.CallCount).To(Equal(0))

				Expect(buffer.String()).ToNot(ContainSubstring("Configuring credentials"))
			})
		})

		context("when there is one bound credential", func() {
			context("when there is no context field", func() {
				it.Before(func() {
					bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
						{
							Path: "some-path",
						},
					}
				})

				it("runs a configuration command", func() {
					err := gitCredentialManager.Setup("working-dir", platformDir)
					Expect(err).NotTo(HaveOccurred())

					Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal(platformDir))
					Expect(executable.ExecuteCall.CallCount).To(Equal(1))

					Expect(executions[0].Dir).To(Equal("working-dir"))
					Expect(executions[0].Args).To(Equal([]string{
						"config",
						"--global",
						"credential.helper",
						fmt.Sprintf("!f() { cat %q; }; f", filepath.Join("some-path", "credentials")),
					}))

					Expect(buffer.String()).To(ContainSubstring("Configuring credentials"))
					Expect(buffer.String()).To(ContainSubstring("Added 1 custom git credential manager(s) to the git config"))
				})
			})

			context("when there is a context field", func() {
				it.Before(func() {
					Expect(os.WriteFile(filepath.Join(platformDir, "example"), []byte(`https://example.com`), 0644)).To(Succeed())

					bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
						{
							Path: "some-path",
							Entries: map[string]*servicebindings.Entry{
								"context": servicebindings.NewEntry(filepath.Join(platformDir, "example")),
							},
						},
					}
				})

				it("runs a configuration command", func() {
					err := gitCredentialManager.Setup("working-dir", platformDir)
					Expect(err).NotTo(HaveOccurred())

					Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal(platformDir))
					Expect(executable.ExecuteCall.CallCount).To(Equal(1))

					Expect(executions[0].Dir).To(Equal("working-dir"))
					Expect(executions[0].Args).To(Equal([]string{
						"config",
						"--global",
						"credential.https://example.com.helper",
						fmt.Sprintf("!f() { cat %q; }; f", filepath.Join("some-path", "credentials")),
					}))
				})
			})
		})

		context("when there is at least two bound credential", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(platformDir, "example"), []byte(`https://example.com`), 0644)).To(Succeed())
				bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
					{
						Path: "some-path",
					},
					{
						Path: "other-path",
						Entries: map[string]*servicebindings.Entry{
							"context": servicebindings.NewEntry(filepath.Join(platformDir, "example")),
						},
					},
				}
			})

			it("runs a configuration commands", func() {
				err := gitCredentialManager.Setup("working-dir", platformDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal(platformDir))
				Expect(executable.ExecuteCall.CallCount).To(Equal(2))

				Expect(executions[0].Dir).To(Equal("working-dir"))
				Expect(executions[0].Args).To(Equal([]string{
					"config",
					"--global",
					"credential.helper",
					fmt.Sprintf("!f() { cat %q; }; f", filepath.Join("some-path", "credentials")),
				}))

				Expect(executions[1].Dir).To(Equal("working-dir"))
				Expect(executions[1].Args).To(Equal([]string{
					"config",
					"--global",
					"credential.https://example.com.helper",
					fmt.Sprintf("!f() { cat %q; }; f", filepath.Join("other-path", "credentials")),
				}))

				Expect(buffer.String()).To(ContainSubstring("Configuring credentials"))
				Expect(buffer.String()).To(ContainSubstring("Added 2 custom git credential manager(s) to the git config"))
			})
		})

		context("failure cases", func() {
			context("when the binding resolver fails", func() {
				it.Before(func() {
					bindingResolver.ResolveCall.Returns.Error = errors.New("failed to resolve bindings")
				})

				it("returns an error", func() {
					err := gitCredentialManager.Setup("working-dir", platformDir)
					Expect(err).To(MatchError("failed to resolve bindings"))
				})
			})

			context("when it fails to read the context entry", func() {
				it.Before(func() {
					Expect(os.WriteFile(filepath.Join(platformDir, "example"), []byte(`https://example.com`), 0000)).To(Succeed())

					bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
						{
							Path: "some-path",
							Entries: map[string]*servicebindings.Entry{
								"context": servicebindings.NewEntry(filepath.Join(platformDir, "example")),
							},
						},
					}
				})

				it("returns an error", func() {
					err := gitCredentialManager.Setup("working-dir", platformDir)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when there are two entries with the same context", func() {
				it.Before(func() {

					bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
						{
							Path: "some-path",
						},
						{
							Path: "other-path",
						},
					}
				})

				it("returns an error", func() {
					err := gitCredentialManager.Setup("working-dir", platformDir)
					Expect(err).To(MatchError("failed: there are two or more bindings for the same context: please limit the bindings to one per context"))
				})
			})

			context("when a command fails", func() {
				it.Before(func() {
					bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
						{
							Path: "some-path",
						},
					}

					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						fmt.Fprintln(execution.Stdout, "build error stdout")
						fmt.Fprintln(execution.Stderr, "build error stderr")
						return errors.New("command failed")
					}
				})

				it("returns an error", func() {
					err := gitCredentialManager.Setup("working-dir", platformDir)
					Expect(err).To(MatchError("command failed"))

					Expect(buffer).To(ContainLines(
						"  Configuring credentials",
						"        build error stdout",
						"        build error stderr",
					))
				})
			})
		})
	})
}
