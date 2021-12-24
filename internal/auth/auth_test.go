package auth_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/timotto/ardumower-relay/internal/auth"
)

var _ = Describe("Auth", func() {
	var param Parameters
	When("auth is disabled", func() {
		BeforeEach(func() {
			param = Parameters{}
			Expect(param.Validate()).ToNot(HaveOccurred())
		})
		It("returns the same static user", func() {
			uut := NewAuth(aLogger(), param)
			Expect(uut.Start(nil)).ToNot(HaveOccurred())

			user1, err := uut.LookupUser(aGetRequest())
			Expect(err).ToNot(HaveOccurred())

			user2, err := uut.LookupUser(aPostRequest())
			Expect(err).ToNot(HaveOccurred())

			user3, err := uut.LookupUser(aGetRequest())
			Expect(err).ToNot(HaveOccurred())

			Expect(user1.Id()).To(Equal(user2.Id()))
			Expect(user1.Id()).To(Equal(user3.Id()))
		})
	})
})
