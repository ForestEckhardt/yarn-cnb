package yarn

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/packit"
)

const (
	PlanDependencyYarn = "yarn"
	PlanDependencyNode = "node"
)

type BuildPlanMetadata struct {
	VersionSource string `toml:"version-source"`
	Build         bool   `toml:"build"`
	Launch        bool   `toml:"launch"`
}

//go:generate faux --interface PackageJSONVersionParser --output fakes/package_json_version_parser.go
type PackageJSONVersionParser interface {
	ParseVersion(path string) (version string, err error)
}

//go:generate faux --interface BuildpackYMLVersionParser --output fakes/buildpack_yml_version_parser.go
type BuildpackYMLVersionParser interface {
	ParseVersion(path string) (version string, err error)
}

func Detect(packageJSONParser PackageJSONVersionParser, buildpackYMLParser BuildpackYMLVersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var requires []packit.BuildPlanRequirement

		yarnVersion, err := buildpackYMLParser.ParseVersion(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if yarnVersion != "" {
			requires = append(requires, packit.BuildPlanRequirement{
				Name:    PlanDependencyYarn,
				Version: yarnVersion,
			})
		}

		nodeVersion, err := packageJSONParser.ParseVersion(filepath.Join(context.WorkingDir, "package.json"))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return packit.DetectResult{}, packit.Fail
			}

			return packit.DetectResult{}, err
		}

		nodeRequirement := packit.BuildPlanRequirement{
			Name: PlanDependencyNode,
			Metadata: BuildPlanMetadata{
				Build:  true,
				Launch: true,
			},
		}

		if nodeVersion != "" {
			nodeRequirement.Version = nodeVersion
			nodeRequirement.Metadata = BuildPlanMetadata{
				VersionSource: "package.json",
				Build:         true,
				Launch:        true,
			}
		}

		requires = append(requires, nodeRequirement)

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: PlanDependencyYarn},
				},
				Requires: requires,
			},
		}, nil
	}
}
