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
	"errors"
	"io"
	"strconv"

	"github.com/dvergnes/log-collector/mocks"
	"github.com/dvergnes/log-collector/processor"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filter", func() {

	var (
		delegate *mocks.EventProcessor

		ep processor.EventProcessor

		s   string
		err error
	)

	BeforeEach(func() {
		delegate = &mocks.EventProcessor{}
	})

	AfterEach(func() {
		delegate.AssertExpectations(GinkgoT())
	})

	Describe("WithLimit", func() {
		const limit = 10
		BeforeEach(func() {
			ep = processor.WithLimit(delegate, limit)
		})

		JustBeforeEach(func() {
			s, err = ep.Next()
		})

		When("decorated processor returns an error", func() {
			criticalError := errors.New("oops")
			BeforeEach(func() {
				delegate.On("Next").Return("", criticalError).Once()
			})
			It("should propagate the error", func() {
				Expect(err).Should(Equal(criticalError))
			})
		})

		When("limit is reached", func() {
			BeforeEach(func() {
				delegate.On("Next").Return("raw", nil).Times(limit)
				for i := 0; i < limit; i++ {
					s, err := ep.Next()
					Expect(err).ShouldNot(HaveOccurred())
					Expect(s).Should(Equal("raw"))
				}
			})
			It("should return EOF", func() {
				Expect(err).Should(Equal(io.EOF))
				Expect(s).Should(BeEmpty())
			})
		})
	})

	Describe("WithFilter", func() {

		BeforeEach(func() {
			ep = processor.WithFilter(delegate, func(s string) bool {
				i, err := strconv.Atoi(s)
				return err == nil && i%2 == 0
			})
		})

		JustBeforeEach(func() {
			s, err = ep.Next()
		})

		When("decorated processor returns an error", func() {
			criticalError := errors.New("oops")
			BeforeEach(func() {
				delegate.On("Next").Return("", criticalError).Once()
			})
			It("should propagate the error", func() {
				Expect(err).Should(Equal(criticalError))
			})
		})

		When("predicate filters events", func() {
			BeforeEach(func() {
				i := 0
				delegate.On("Next").Return(func() string {
					i++
					return strconv.Itoa(i)
				}, nil).Twice()
			})

			It("should not return the filtered events", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(s).Should(Equal("2"))
			})
		})
	})

})
