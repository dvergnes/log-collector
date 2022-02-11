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
		eventBreaker = processor.NewEventBreaker(reader, bufio.ScanLines, 35)
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
				})).Return(len(content), nil).Once()
			})
			It("should return whatever was there", func() {
				e1, err := eventBreaker.Next()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(e1).Should(Equal(content))
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
			var (
				chunk2 = "event_spanning_over_chunks\n"
				chunk1 = `anning_over_chunks
event_1
`
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
				reader.On("SeekToEnd", uint32(len("anning_over_chunks\n")))
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

			When("splitter returns an error while searching for the event boundary", func() {

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

			When("splitter returns an error while breaking the events", func() {
				call := 0
				BeforeEach(func() {
					reader.On("Read", mock.MatchedBy(func(buf []byte) bool {
						copy(buf, content)
						return true
					})).Return(len(content), nil).Once()
					reader.On("SeekToEnd", uint32(6))

					eventBreaker = processor.NewEventBreaker(reader, func([]byte, bool) (int, []byte, error) {
						if call == 0 {
							call++
							return len("start\n"), []byte("start"), nil
						}
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
