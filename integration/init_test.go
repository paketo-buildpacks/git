package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var settings struct {
	Buildpacks struct {
		Git    string
		GoDist string
	}
	Buildpack struct {
		Name string
		ID   string
	}
	Config struct {
		GoDist string `json:"go-dist"`
	}
}

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	file, err := os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&settings.Config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.DecodeReader(file, &settings)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	buildpackStore := occam.NewBuildpackStore()

	settings.Buildpacks.Git, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.GoDist, err = buildpackStore.Get.
		Execute(settings.Config.GoDist)
	Expect(err).ToNot(HaveOccurred())

	SetDefaultEventuallyTimeout(5 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Default", testDefault)
	suite.Run(t)
}
