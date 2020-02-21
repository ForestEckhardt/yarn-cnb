package integration

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSimpleApp(t *testing.T, context spec.G, it spec.S) {
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

	context("when the build process is online", func() {
		it("should build a working OCI image for a simple app", func() {
			var err error
			var logs fmt.Stringer
			image, logs, err = pack.WithVerbose().Build.
				WithBuildpacks(nodeURI, yarnURI).
				WithNoPull().
				Execute(name, filepath.Join("testdata", "simple_app"))
			Expect(err).NotTo(HaveOccurred(), logs.String)

			container, err = docker.Container.Run.WithCommand("yarn --version").Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(ContainerLogs(container.ID)).Should(MatchRegexp(`1\.\d+\.\d+`))
		})
	})

	context("when the build process is offline", func() {
		it("should build a working OCI image for a simple app", func() {
			var err error
			var logs fmt.Stringer
			image, logs, err = pack.WithVerbose().Build.
				WithBuildpacks(nodeCachedURI, yarnCachedURI).
				WithNoPull().
				WithNetwork("none").
				Execute(name, filepath.Join("testdata", "simple_app"))
			Expect(err).NotTo(HaveOccurred(), logs.String)

			container, err = docker.Container.Run.WithCommand("yarn --version").Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(ContainerLogs(container.ID)).Should(MatchRegexp(`1\.\d+\.\d+`))
		})
	})
}
