package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testCredentialConfiguration(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when building a simple app that configures credentials", func() {
		var (
			image     occam.Image
			container occam.Container
			root      string
			name      string
			source    string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			err = os.Rename(filepath.Join(source, ".git.bak"), filepath.Join(source, ".git"))
			Expect(err).NotTo(HaveOccurred())

			root, err = filepath.Abs("./..")
			Expect(err).ToNot(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("builds an OCI image with git environment variables and sets credentials in credential manager during build", func() {
			var err error
			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.Git,
					settings.Buildpacks.CredentialFill,
				).
				WithEnv(map[string]string{
					"SERVICE_BINDING_ROOT": "/bindings",
				}).
				WithVolumes(fmt.Sprintf("%s:/bindings/credentials", filepath.Join(root, "integration", "testdata", "credential_binding"))).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Configuring build environment",
				`    REVISION -> "2df6ac40991b695cc6c31faa79926980ff7dc0ff"`,
				"",
				"  Configuring launch environment",
				`    REVISION -> "2df6ac40991b695cc6c31faa79926980ff7dc0ff"`,
				"",
				"  Configuring credentials",
				"",
				"Paketo Credential Fill Buildpack",
				"protocol=https",
				"host=example.com",
				"username=username",
				"password=password/token",
			))

			Expect(image.Labels).To(HaveKeyWithValue("org.opencontainers.image.revision", "2df6ac40991b695cc6c31faa79926980ff7dc0ff"))

			container, err = docker.Container.Run.
				WithCommand("echo $REVISION").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() string {
				logs, _ := docker.Container.Logs.Execute(container.ID)
				return logs.String()
			}).Should(ContainSubstring("2df6ac40991b695cc6c31faa79926980ff7dc0ff"))
		})
	})
}
