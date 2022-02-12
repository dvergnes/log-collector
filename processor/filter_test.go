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
