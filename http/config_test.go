// Copyright (c) 2022 Denis Vergnes
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package http_test

import (
	"time"

	"github.com/dvergnes/log-collector/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Config", func() {

	Describe("LoadConfig", func() {
		var (
			fs   = afero.NewMemMapFs()
			err  error
			conf *http.Config
		)

		BeforeEach(func() {
			Expect(fs.MkdirAll("/var/log/", 0755))
			_, err := fs.Create("/tmp/ut.log")
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("config is not fully specified", func() {
			JustBeforeEach(func() {
				conf, err = http.LoadConfig([]byte(`port: 8888`), fs)
			})
			It("sets some defaults", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(conf.BufferSize).Should(BeEquivalentTo(4096))
				Expect(conf.MaxEvents).Should(BeEquivalentTo(10_000))
				Expect(conf.ShutdownTimeout).Should(Equal(30 * time.Second))
				Expect(conf.LogFolder).Should(Equal("/var/log/"))
			})
		})

		When("log folder is invalid", func() {
			DescribeTable("it should return an error", func(data []byte, msg string) {
				_, err := http.LoadConfig(data, fs)
				Expect(err).Should(MatchError(msg))
			},
				Entry("declared folder does not exist", []byte("log_folder: /foo/bar"), "log folder declared in configuration does not exist"),
				Entry("log folder is not a directory", []byte("log_folder: /tmp/ut.log"), "log folder declared in configuration is not a directory"),
			)
		})

		When("shutdown timeout is invalid", func() {
			JustBeforeEach(func() {
				conf, err = http.LoadConfig([]byte(`shutdown_timeout: -1`), fs)
			})
			It("should return an error", func() {
				Expect(err).Should(MatchError("shutdown timeout must be strictly positive"))
			})
		})
	})

})
