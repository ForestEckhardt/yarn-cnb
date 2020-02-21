package integration

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testCaching(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker

		imageIDs     map[string]struct{}
		containerIDs map[string]struct{}

		imageName string
	)

	it.Before(func() {
		var err error
		imageIDs = make(map[string]struct{})
		containerIDs = make(map[string]struct{})

		pack = occam.NewPack()
		docker = occam.NewDocker()

		imageName, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		for id, _ := range containerIDs {
			Expect(docker.Container.Remove.Execute(id)).To(Succeed())
		}

		for id, _ := range imageIDs {
			Expect(docker.Image.Remove.Execute(id)).To(Succeed())
		}

		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(imageName))).To(Succeed())
	})

	context("when the node_modules are NOT vendored", func() {
		it("should build a working OCI image for a simple app", func() {
			var err error
			var logs fmt.Stringer

			build := pack.WithNoColor().Build.WithBuildpacks(nodeURI, yarnURI).WithNoPull()

			firstImage, logs, err := build.Execute(imageName, filepath.Join("testdata", "simple_app"))
			Expect(err).NotTo(HaveOccurred(), logs.String)

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[1].Key).To(Equal("org.cloudfoundry.yarn"))
			Expect(firstImage.Buildpacks[1].Layers).To(HaveKey("yarn"))

			container, err := docker.Container.Run.WithCommand("yarn --version").Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[container.ID] = struct{}{}

			Eventually(ContainerLogs(container.ID)).Should(MatchRegexp(`1\.\d+\.\d+`))

			secondImage, logs, err := build.Execute(imageName, filepath.Join("testdata", "simple_app"))
			Expect(err).NotTo(HaveOccurred(), logs.String)

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[1].Key).To(Equal("org.cloudfoundry.yarn"))
			Expect(secondImage.Buildpacks[1].Layers).To(HaveKey("yarn"))

			container, err = docker.Container.Run.WithCommand("yarn --version").Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[container.ID] = struct{}{}

			Eventually(ContainerLogs(container.ID)).Should(MatchRegexp(`1\.\d+\.\d+`), ContainerLogs(container.ID))

			// buildpackVersion, err := GetGitVersion()
			// Expect(err).ToNot(HaveOccurred())

			splitLogs := GetBuildLogs(logs.String())
			Expect(splitLogs).To(ContainSequence([]interface{}{
				fmt.Sprintf("Yarn Buildpack %s", "0.0.0"),
				"  Reusing cached layer /layers/org.cloudfoundry.yarn/yarn",
				"",
			},
			), logs.String)
		})
	})
}
