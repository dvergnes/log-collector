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

package processor_test

import (
	"io"
	"time"

	"github.com/dvergnes/log-collector/processor"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Reader", func() {

	var (
		fs         afero.Fs
		tailReader *processor.TailReader
	)

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
	})

	Describe("New", func() {
		When("file does not exist", func() {
			It("should return an error", func() {
				_, err := processor.NewTailReader(fs, "I_dont_exist")
				Expect(err).Should(MatchError(ContainSubstring("failed to open file")))
			})
		})
	})

	Describe("Read", func() {
		var (
			buf   = make([]byte, 10)
			n     = 0
			err   error
			setUp = func(content string) afero.File {
				f, err := afero.TempFile(fs, "ut", "file.log")
				Expect(err).ShouldNot(HaveOccurred())
				_, err = f.WriteString(content)
				Expect(err).ShouldNot(HaveOccurred())
				tailReader, err = processor.NewTailReader(fs, f.Name())
				Expect(err).ShouldNot(HaveOccurred())
				return f
			}
		)

		AfterEach(func() {
			Expect(tailReader.Close()).Should(Succeed())
		})

		JustBeforeEach(func() {
			n, err = tailReader.Read(buf)
		})

		When("reading the last line", func() {
			BeforeEach(func() {
				setUp(`
123456789
abcdefghi
`)
			})
			It("should fill the buffer by reading from the end of file", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf[:n]).Should(Equal([]byte("abcdefghi\n")))
			})
		})

		When("reading the entire file", func() {
			BeforeEach(func() {
				setUp(`12345`)
			})
			It("should fill the buffer and reach EOF", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf[:n]).Should(Equal([]byte("12345")))
			})
		})

		When("reading the file with multiple calls", func() {
			BeforeEach(func() {
				setUp(`
123456789
abcdefghi
`)
			})
			It("should stream the file from the end to the start", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf[:n]).Should(Equal([]byte("abcdefghi\n")))

				n, err = tailReader.Read(buf)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf[:n]).Should(Equal([]byte("123456789\n")))

				n, err = tailReader.Read(buf)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(buf[:n]).Should(Equal([]byte("\n")))

				n, err = tailReader.Read(buf)
				Expect(n).Should(BeZero())
				Expect(err).Should(Equal(io.EOF))
			})
		})

		When("file is modified after TailReader is opened", func() {
			BeforeEach(func() {
				f := setUp(`abc`)
				inTheFuture := time.Now().Add(time.Second)
				Expect(fs.Chtimes(f.Name(), inTheFuture, inTheFuture)).Should(Succeed())
			})

			It("should return an error", func() {
				Expect(err).Should(Satisfy(processor.IsConcurrentAccessError))
			})
		})

	})

	Describe("SeekToEnd", func() {

	})
})
