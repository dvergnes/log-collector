package http_test

import (
	"github.com/dvergnes/log-collector/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Controller", func() {

	Describe("validateFileParameter", func() {
		When("file name is valid", func() {
			It("should not return an error", func() {
				Expect(http.ValidateFileParameter("foo.log")).Should(Succeed())
			})
		})

		When("file name is invalid", func() {
			DescribeTable("should return an error", func(file string, msg string) {
				err := http.ValidateFileParameter(file)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError(msg))
			},
				Entry("empty file name", "", "file name must not be empty"),
				Entry("file is a relative path", "../../etc/master.passwd", "file name must not contain any relative or absolute path reference"),
			)
		})

	})

	Describe("parseLimit", func() {
		const max = 100
		When("limit is valid", func() {
			It("should return the limit as a uint", func() {
				l, err := http.ParseLimit(max, "30")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(l).Should(Equal(uint(30)))
			})
		})

		When("limit is empty", func() {
			It("should return 0", func() {
				l, err := http.ParseLimit(max, "")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(l).Should(BeZero())
			})
		})

		When("limit is invalid", func() {
			DescribeTable("should return an error", func(limit string, msg string) {
				_, err := http.ParseLimit(max, limit)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(MatchError(msg))
			},
				Entry("limit is not an integer", "I am not an integer", "limit is not a valid integer"),
				Entry("limit is negative", "-5", "limit must be strictly positive"),
				Entry("limit is too big", "200", "limit must be equal or less than 100"),
			)
		})

	})

	Describe("checkFile", func() {
		var fs afero.Fs
		BeforeEach(func() {
			fs = afero.NewMemMapFs()
			Expect(fs.MkdirAll("/var/log", 0755)).Should(Succeed())
			_, err := fs.Create("/var/log/foo.log")
			Expect(err).ShouldNot(HaveOccurred())
		})
		When("file is valid", func() {
			It("should not return an error", func() {
				Expect(http.CheckFile(fs, "/var/log/foo.log")).Should(Succeed())
			})
		})

		When("file is invalid", func() {
			DescribeTable("should return an error", func(path string, msg string) {
				err := http.CheckFile(fs, path)
				Expect(err).Should(MatchError(msg))
			},
				Entry("file does not exist", "/var/log/not_found", "file /var/log/not_found was not found"),
				Entry("file is a directory", "/var/log", "file /var/log is a directory"),
			)
		})

	})

	Describe("logHandler", func() {
		When("parameters are invalid", func() {

		})

		When("limit is unspecified", func() {

		})

		When("file is invalid", func() {})

		When("filter is applied", func() {

		})

	})

})
