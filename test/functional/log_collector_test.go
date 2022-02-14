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

package functional_test

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/dvergnes/log-collector/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("LogCollector", func() {
	When("request is valid", func() {
		BeforeEach(func() {
			Expect(afero.WriteFile(fs, "/var/log/file.log",[]byte(`
dvergnes submits MR 432
alice submits MR 123
jdoe approves MR 123
dvergnes closes MR 432
dvergnes approves MR 123
bob approves MR 123
dvergnes submits MR 890
jdoe approves MR 890
`), 0755)).Should(Succeed())
		})
		It("should return a valid response", func() {
			resp, err:=http.Get("http://localhost:9999/log?file=file.log&limit=3&filter=dvergnes")
			Expect(err).ShouldNot(HaveOccurred())
			body, err := io.ReadAll(resp.Body)
			Expect(err).ShouldNot(HaveOccurred())
			logResp := api.LogResponse{}
			Expect(json.Unmarshal(body, &logResp)).Should(Succeed())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			Expect(logResp.Events).Should(HaveLen(3))
			Expect(logResp.Events[0]).Should(Equal("dvergnes submits MR 890"))
			Expect(logResp.Events[1]).Should(Equal("dvergnes approves MR 123"))
			Expect(logResp.Events[2]).Should(Equal("dvergnes closes MR 432"))
		})
	})

	When("request is invalid", func() {
		It("should return a response with an error", func() {
			resp, err:=http.Get("http://localhost:9999/log?file=not_found")
			Expect(err).ShouldNot(HaveOccurred())
			body, err := io.ReadAll(resp.Body)
			Expect(err).ShouldNot(HaveOccurred())
			errResp := api.ErrorResponse{}
			Expect(json.Unmarshal(body, &errResp)).Should(Succeed())
			Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))
			Expect(errResp.Code).Should(Equal("invalid.parameter"))
			Expect(errResp.Details).Should(Equal("file /var/log/not_found was not found"))
		})
	})
})
