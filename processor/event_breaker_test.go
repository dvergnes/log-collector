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
	"bytes"
	"errors"
	"io"

	"github.com/dvergnes/log-collector/mocks"
	"github.com/dvergnes/log-collector/processor"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("EventBreaker", func() {
	var (
		eventBreaker *processor.EventBreaker
		reader       *mocks.TailReader
	)

	BeforeEach(func() {
		reader = &mocks.TailReader{}
		eventBreaker = processor.NewEventBreaker(reader, processor.ReverseScanLines, 35)
	})

	AfterEach(func() {
		reader.AssertExpectations(GinkgoT())
	})

	Describe("Next", func() {
		When("reader returns EOF", func() {
			BeforeEach(func() {
				reader.On("Read", mock.Anything).Return(0, io.EOF)
			})
			It("should return EOF", func() {
				e1, err := eventBreaker.Next()
				Expect(err).Should(MatchError(io.EOF))
				Expect(e1).Should(BeEmpty())
			})
		})

		When("there is no event separator", func() {
			content := `truncated_event  event_1  event_2`
			BeforeEach(func() {
				reader.On("Read", mock.MatchedBy(func(buf []byte) bool {
					copy(buf, content)
					return true
				})).Return(len(content), nil).Twice()
				reader.On("SeekToEnd", uint32(len(content)))
			})
			It("should return whatever was read", func() {
				e1, err := eventBreaker.Next()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(e1).Should(Equal(content))
			})
		})

		When("there are empty events", func() {
			BeforeEach(func() {
				content := bytes.Buffer{}
				content.WriteString("event_1")
				for i := 0; i < 40; i++ {
					content.WriteByte('\n')
				}
				content.WriteString("event_2")
				content.Bytes()
				n := content.Len()
				call := 0
				reader.On("Read", mock.MatchedBy(func(buf []byte) bool {
					b := content.Bytes()
					if call == 0 {
						copy(buf, b[n-len(buf):n])
					} else {
						copy(buf, b[:n-len(buf)])
					}
					call++
					return true
				})).Return(func(b []byte) int {
					if call <= 1 {
						return len(b)
					} else {
						return n - len(b)
					}
				}, nil).Times(3)
				reader.On("SeekToEnd", uint32(len("event_1")))
			})
			It("should skip the empty events", func() {
				e1, err := eventBreaker.Next()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(e1).Should(Equal("event_2"))

				e2, err := eventBreaker.Next()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(e2).Should(Equal("event_1"))
			})
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
				//reader.On("SeekToEnd", uint32(16))
			})
			It("should return events one by one in reverse order", func() {
				e1, err := eventBreaker.Next()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(e1).Should(Equal("event_2"))

				e2, err := eventBreaker.Next()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(e2).Should(Equal("event_1"))

			})
		})

		When("buffer is empty after reading all available events", func() {
			var (
				chunk2 = "event_spanning_over_chunks"
				chunk1 = `anning_over_chunks
event_1`
				chunkIdx = 0
			)

			BeforeEach(func() {
				reader.On("Read", mock.MatchedBy(func(buf []byte) bool {
					if chunkIdx == 0 {
						copy(buf, chunk1)
						chunkIdx++
					} else {
						copy(buf, chunk2)
					}
					return true
				})).Return(func(data []byte) int {
					if bytes.Contains(data, []byte(chunk1)) {
						return len(chunk1)
					} else {
						return len(chunk2)
					}
				}, nil).Twice()
				reader.On("SeekToEnd", uint32(len("anning_over_chunks")))
			})
			It("should fill buffer by calling the reader where it left off", func() {
				e1, err := eventBreaker.Next()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(e1).Should(Equal("event_1"))

				e2, err := eventBreaker.Next()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(e2).Should(Equal("event_spanning_over_chunks"))
			})
		})

		When("splitter fails to split", func() {
			var (
				content       = "start\nevent"
				criticalError = errors.New("oops")
			)

			When("splitter returns an error while breaking the events", func() {

				BeforeEach(func() {
					reader.On("Read", mock.MatchedBy(func(buf []byte) bool {
						copy(buf, content)
						return true
					})).Return(len(content), nil).Once()
					eventBreaker = processor.NewEventBreaker(reader, func([]byte, bool) (int, []byte, error) {
						return 0, nil, criticalError
					}, 35)
				})

				It("should propagate the error", func() {
					e1, err := eventBreaker.Next()
					Expect(err).Should(MatchError(criticalError))
					Expect(e1).Should(BeEmpty())
				})
			})

		})
	})
})
