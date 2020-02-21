package yarn_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ForestEckhardt/yarn-cnb/yarn"
	"github.com/ForestEckhardt/yarn-cnb/yarn/fakes"
	"github.com/cloudfoundry/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		packageJSONParser  *fakes.PackageJSONVersionParser
		buildpackYMLParser *fakes.BuildpackYMLVersionParser
		workingDir         string
		detect             packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(workingDir, "yarn.lock"), []byte{}, 0644)
		Expect(err).NotTo(HaveOccurred())

		packageJSONParser = &fakes.PackageJSONVersionParser{}
		packageJSONParser.ParseVersionCall.Returns.Version = "some-version"

		buildpackYMLParser = &fakes.BuildpackYMLVersionParser{}

		detect = yarn.Detect(packageJSONParser, buildpackYMLParser)
	})

	it("returns a plan that provides yarn", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: workingDir,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: "yarn"},
			},
			Requires: []packit.BuildPlanRequirement{
				{
					Name:    "node",
					Version: "some-version",
					Metadata: yarn.BuildPlanMetadata{
						VersionSource: "package.json",
						Build:         true,
						Launch:        true,
					},
				},
			},
		}))

		Expect(packageJSONParser.ParseVersionCall.Receives.Path).To(Equal(filepath.Join(workingDir, "package.json")))
		Expect(buildpackYMLParser.ParseVersionCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
	})

	context("when the node version is not in the package.json file", func() {
		it.Before(func() {
			packageJSONParser.ParseVersionCall.Returns.Version = ""
		})

		it("returns a plan that provides yarn", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "yarn"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "node",
						Metadata: yarn.BuildPlanMetadata{
							Build:  true,
							Launch: true,
						},
					},
				},
			}))

			Expect(packageJSONParser.ParseVersionCall.Receives.Path).To(Equal(filepath.Join(workingDir, "package.json")))
			Expect(buildpackYMLParser.ParseVersionCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
		})
	})

	context("there is a yarn version in the buildpack.yml", func() {
		it.Before(func() {
			buildpackYMLParser.ParseVersionCall.Returns.Version = "some-version"
		})

		it("returns a plan that provides and requires yarn", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "yarn"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name:    "yarn",
						Version: "some-version",
					}, {
						Name:    "node",
						Version: "some-version",
						Metadata: yarn.BuildPlanMetadata{
							VersionSource: "package.json",
							Build:         true,
							Launch:        true,
						},
					},
				},
			}))

			Expect(packageJSONParser.ParseVersionCall.Receives.Path).To(Equal(filepath.Join(workingDir, "package.json")))
			Expect(buildpackYMLParser.ParseVersionCall.Receives.Path).To(Equal(filepath.Join(workingDir, "buildpack.yml")))
		})
	})

	context("when there is no package.json file", func() {
		it.Before(func() {
			_, err := os.Stat("/no/such/package.json")
			packageJSONParser.ParseVersionCall.Returns.Err = err
		})

		it("fails detection", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(MatchError(packit.Fail))
		})
	})

	context("failure cases", func() {
		context("when buildpack.yml cannot be read", func() {
			it.Before(func() {
				buildpackYMLParser.ParseVersionCall.Returns.Err = errors.New("failed to read buildpack.yml")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError("failed to read buildpack.yml"))
			})
		})

		context("when the package.json cannot be read", func() {
			it.Before(func() {
				packageJSONParser.ParseVersionCall.Returns.Err = errors.New("failed to read package.json")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError("failed to read package.json"))
			})
		})
	})
}
