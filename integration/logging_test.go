package integration

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLogging(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker

		image     occam.Image
		container occam.Container

		name string
	)

	it.Before(func() {
		var err error
		pack = occam.NewPack()
		docker = occam.NewDocker()

		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
		Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
	})

	context("when the build process runs", func() {
		it("should build a working OCI image for a simple app with useful logs", func() {
			var err error
			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().WithVerbose().Build.
				WithBuildpacks(nodeURI, yarnURI).
				WithNoPull().
				Execute(name, filepath.Join("testdata", "simple_app"))
			Expect(err).NotTo(HaveOccurred(), logs.String)

			container, err = docker.Container.Run.WithCommand("yarn --version").Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(ContainerLogs(container.ID)).Should(MatchRegexp(`1\.\d+\.\d+`))

			splitLogs := GetBuildLogs(logs.String())
			Expect(splitLogs).To(ContainSequence([]interface{}{
				fmt.Sprintf("Yarn Buildpack %s", "0.0.0"),
				"  Executing build process",
				MatchRegexp(`    Installing Yarn 1\.\d+\.\d+`),
				MatchRegexp(`      Completed in (\d+\.\d+|\d{3})`),
				"",
			}))
		})
	})
}
