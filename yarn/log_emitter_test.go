package yarn_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/ForestEckhardt/yarn-cnb/yarn"
	"github.com/cloudfoundry/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLogEmitter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		emitter yarn.LogEmitter
		buffer  *bytes.Buffer
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		emitter = yarn.NewLogEmitter(scribe.NewLogger(buffer))
	})

	context("BuildpackTitle", func() {
		it("given the name and version is prints the buildpack title", func() {
			emitter.BuildpackTitle("some-buildpack", "some-version")
			Expect(buffer.String()).To(ContainSubstring("some-buildpack some-version"))
		})
	})

	context("CompletionTime", func() {
		it("returns a string that prints out the time elapsed since the passed in value round to the millisecond", func() {
			then := time.Now()
			time.Sleep(100 * time.Millisecond)
			emitter.CompletionTime(then)
			Expect(buffer.String()).To(MatchRegexp(`      Completed in (\d+\.\d+|\d{3})\w+\n\n`))
		})
	})

	context("ReusingLayer", func() {
		it("prints a layer reuse message", func() {
			emitter.ReusingLayer("some-filepath")
			Expect(buffer.String()).To(Equal("  Reusing cached layer some-filepath\n\n"))
		})
	})
}
