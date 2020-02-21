package main

import (
	"os"
	"time"

	"github.com/ForestEckhardt/yarn-cnb/yarn"
	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/postal"
	"github.com/cloudfoundry/packit/scribe"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	logEmitter := yarn.NewLogEmitter(logger)

	transport := cargo.NewTransport()
	dependencyService := postal.NewService(transport)

	clock := yarn.NewClock(time.Now)
	cacheHandler := yarn.NewCacheHandler()

	packit.Build(yarn.Build(dependencyService, cacheHandler, clock, logEmitter))
}
