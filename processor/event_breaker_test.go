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
	"bufio"
	"github.com/dvergnes/log-collector/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/dvergnes/log-collector/processor"
)

var _ = Describe("EventBreaker", func() {
	var (
		eventBreaker *processor.EventBreaker
		reader       *mocks.TailReader
	)

	BeforeEach(func() {
		reader = &mocks.TailReader{}
		eventBreaker = processor.NewEventBreaker(reader, bufio.ScanLines, 1024)
	})

	AfterEach(func() {
		reader.AssertExpectations(GinkgoT())
	})

	Describe("Next", func() {
		When("reader returns EOF", func() {
			It("should return EOF", func() {})
		})

		When("there is no event separator", func() {
			It("should return whatever is present", func() {})
		})

		When("there are some events", func() {
			content := `truncated_event
event_1
event_2
`
			BeforeEach(func() {
				reader.On("Read", mock.MatchedBy(func(buf []byte) bool {
					copy(buf, content)
					return true
				})).Return(len(content), nil).Once()
				reader.On("SeekToEnd", uint32(16))
			})
			It("should return events one by one", func() {
				e1, err := eventBreaker.Next()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(e1).Should(Equal("event_1"))

				e2, err := eventBreaker.Next()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(e2).Should(Equal("event_2"))

			})
		})

		When("buffer is empty after reading all available events", func() {
			It("should fill buffer by calling the reader where it left off", func() {})
		})

		When("splitter returns an error while searching for the event boundary", func() {
			It("should propagate the error", func() {})
		})

		When("splitter returns an error while breaking the events", func() {
			It("should propagate the error", func() {})
		})
	})
})
