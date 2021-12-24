package app_endpoint_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/timotto/ardumower-relay/internal/app_endpoint"
	"time"
)

var _ = Describe("Parameters", func() {
	It("sets defaults", func() {
		uut := app_endpoint.Parameters{}

		Expect(uut.Validate()).ToNot(HaveOccurred())

		Expect(uut.Timeout).ToNot(BeZero())
	})

	It("does not overwrite given values", func() {
		givenTimeout := time.Hour

		uut := app_endpoint.Parameters{Timeout: givenTimeout}

		Expect(uut.Validate()).ToNot(HaveOccurred())

		Expect(uut.Timeout).To(Equal(givenTimeout))
	})
})
