package git_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitGit(t *testing.T) {
	suite := spec.New("git", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite("GitCredentialManager", testGitCredentialManager)
	suite.Run(t)
}
