package git

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"
)

//go:generate faux --interface BindingResolver --output fakes/binding_resolver.go
type BindingResolver interface {
	Resolve(typ, provider, platformDir string) ([]servicebindings.Binding, error)
}

type GitCredentialManager struct {
	bindingResolver BindingResolver
	executable      Executable
	logs            scribe.Emitter
}

func NewGitCredentialManager(bindingResolver BindingResolver, executable Executable, logs scribe.Emitter) GitCredentialManager {
	return GitCredentialManager{
		bindingResolver: bindingResolver,
		executable:      executable,
		logs:            logs,
	}
}

func (g GitCredentialManager) Setup(workingDir, platformDir string) error {
	g.logs.Process("Configuring credentials")

	bindings, err := g.bindingResolver.Resolve("git-credentials", "", platformDir)
	if err != nil {
		return err
	}

	uniqueContext := map[string]interface{}{}

	for _, b := range bindings {
		context := "credential.helper"
		if entry, ok := b.Entries["context"]; ok {
			domain, err := entry.ReadString()
			if err != nil {
				return err
			}

			context = fmt.Sprintf("credential.%s.helper", strings.TrimSpace(domain))
		}

		// Checks to see if the context is unique if not error because having
		// multiple credentials for the same context is not supported
		_, exists := uniqueContext[context]
		if !exists {
			uniqueContext[context] = nil
		} else {
			return fmt.Errorf("failed: there are two or more bindings for the same context: please remove limit the bindings to one per context")
		}

		buffer := bytes.NewBuffer(nil)
		err := g.executable.Execute(pexec.Execution{
			Args: []string{
				"config",
				"--global",
				context,
				fmt.Sprintf("!f() { cat %s; }; f", filepath.Join(b.Path, "credentials")),
			},
			Dir:    workingDir,
			Stdout: buffer,
			Stderr: buffer,
		})

		if err != nil {
			g.logs.Detail(buffer.String())
			return err
		}

	}

	g.logs.Break()
	return nil
}
