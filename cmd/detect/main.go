package main

import (
	"github.com/ForestEckhardt/yarn-cnb/yarn"
	"github.com/cloudfoundry/packit"
)

func main() {
	packageJSONParser := yarn.NewPackageJSONParser()
	buildpackYMLParser := yarn.NewBuildpackYMLParser()

	packit.Detect(yarn.Detect(packageJSONParser, buildpackYMLParser))
}
