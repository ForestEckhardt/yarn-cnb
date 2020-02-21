package yarn

import (
	"time"

	"github.com/cloudfoundry/packit/scribe"
)

type LogEmitter struct {
	scribe.Logger
}

func NewLogEmitter(logger scribe.Logger) LogEmitter {
	return LogEmitter{logger}
}

func (e LogEmitter) BuildpackTitle(name, version string) {
	e.Logger.Title("%s %s", name, version)
}

func (e LogEmitter) CompletionTime(then time.Time) {
	e.Logger.Action("Completed in %s", time.Since(then).Round(time.Millisecond))
	e.Logger.Break()
}

func (e LogEmitter) ReusingLayer(layerPath string) {
	e.Logger.Process("Reusing cached layer %s", layerPath)
	e.Logger.Break()
}

func (e LogEmitter) RunningInstall(offline bool) {
	installMessage := "Running"
	if offline {
		installMessage = "Running offline"
	}
	e.Logger.Subprocess("%s 'yarn install'", installMessage)
}

func (e LogEmitter) FoundYarnLock(found bool) {
	foundMessage := "Not found"
	if found {
		foundMessage = "Found"
	}

	e.Logger.Action("yarn.lock -> %s", foundMessage)
	e.Logger.Break()
}
