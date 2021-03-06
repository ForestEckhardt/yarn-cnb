package yarn_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitYarn(t *testing.T) {
	suite := spec.New("yarn", spec.Report(report.Terminal{}))
	suite("Build", testBuild)
	suite("CacheHandler", testCacheHandler)
	suite("Clock", testClock)
	suite("BuildpackYAMLParser", testBuildpackYMLParser)
	suite("Detect", testDetect)
	// suite("InstallProcess", testInstallProcess)
	suite("LogEmitter", testLogEmitter)
	suite("PackageJSONParser", testPackageJSONParser)
	suite.Run(t)
}
