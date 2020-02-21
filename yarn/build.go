package yarn

import (
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/postal"
)

//go:generate faux --interface CacheMatcher --output fakes/cache_matcher.go
type CacheMatcher interface {
	Match(metadata map[string]interface{}, key, sha string) bool
}

//go:generate faux --interface DependencyService --output fakes/dependency_service.go
type DependencyService interface {
	Resolve(path, name, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
}

func Build(dependencyService DependencyService, cacheMatcher CacheMatcher, clock Clock, logEmitter LogEmitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logEmitter.BuildpackTitle(context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		yarnLayer, err := context.Layers.Get("yarn", packit.LaunchLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}

		//TODO:Write a dep resolver
		dependency, err := dependencyService.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), "yarn", "*", context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		if !cacheMatcher.Match(yarnLayer.Metadata, "cache_sha", dependency.SHA256) {
			logEmitter.Logger.Process("Executing build process")

			err = yarnLayer.Reset()
			if err != nil {
				return packit.BuildResult{}, err
			}

			logEmitter.Logger.Subprocess("Installing Yarn %s", dependency.Version)

			then := clock.Now()

			err = dependencyService.Install(dependency, context.CNBPath, yarnLayer.Path)
			if err != nil {
				return packit.BuildResult{}, err
			}

			logEmitter.CompletionTime(then)

			yarnLayer.Metadata = map[string]interface{}{
				"built_at":  clock.Now().Format(time.RFC3339Nano),
				"cache_sha": dependency.SHA256,
			}

			//TODO:Add logging
			yarnLayer.SharedEnv.Append("PATH", yarnLayer.Path, string(os.PathListSeparator))
		} else {
			logEmitter.ReusingLayer(yarnLayer.Path)
		}

		return packit.BuildResult{
			Plan: context.Plan,
			Layers: []packit.Layer{
				yarnLayer,
			},
			Processes: []packit.Process{
				{
					Type:    "web",
					Command: "yarn start",
				},
			},
		}, nil
	}
}
