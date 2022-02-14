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
	"testing"
	"time"

	"github.com/dvergnes/log-collector/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func TestFunctional(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Functional Suite")
}

var (
	server *http.Server
	fs     afero.Fs
	prout     afero.Fs
)

var _ = BeforeSuite(func() {
	fs = afero.NewMemMapFs()
	Expect(fs.MkdirAll("/var/log/", 0755)).Should(Succeed())
	server = http.NewServer(&http.Config{
		Port:            9999,
		BufferSize:      4096,
		ShutdownTimeout: time.Second,
		MaxEvents:       100,
		LogFolder: "/var/log/",
	}, fs, zap.NewNop())
	go server.Start()
})

var _ = AfterSuite(func() {
	server.Stop()
})
