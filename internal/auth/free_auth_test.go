package auth_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/timotto/ardumower-relay/internal/auth"
	"math/rand"
)

var _ = Describe("FreeAuth", func() {
	var uut Auth
	BeforeEach(func() {
		uut = NewAuth(aLogger(), valid(Parameters{}))
	})
	When("the request has no credentials", func() {
		It("returns an error", func() {
			user, err := uut.LookupUser(aGetRequest())
			Expect(user).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(ErrAuthRequired))
		})
	})
	It("never rejects any given credentials", func() {
		for i := 0; i < 1000; i++ {
			user, err := uut.LookupUser(aGetRequestWithCredentials(aRandomString(), aRandomString()))
			Expect(err).ToNot(HaveOccurred())
			Expect(user).ToNot(BeNil())
		}
	})

	It("differentiates between users", func() {
		user1, err := uut.LookupUser(aGetRequestWithCredentials("user1", "password1"))
		Expect(err).ToNot(HaveOccurred())
		Expect(user1).ToNot(BeNil())

		user2, err := uut.LookupUser(aGetRequestWithCredentials("user2", "password1"))
		Expect(err).ToNot(HaveOccurred())
		Expect(user2).ToNot(BeNil())

		user2b, err := uut.LookupUser(aGetRequestWithCredentials("user2", "password2"))
		Expect(err).ToNot(HaveOccurred())
		Expect(user2b).ToNot(BeNil())

		Expect(user1.Id()).ToNot(Equal(user2.Id()))

		Expect(user2b.Id()).ToNot(Equal(user2.Id()))
	})
})

func aRandomString() string {
	const letters = "01234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ,./<>?;':\"[]\\{}|!@#$%^&*()-=_+`~"
	const length = 8

	result := make([]byte, 0)
	for i := 0; i < length; i++ {
		result = append(result, letters[rand.Intn(len(letters))])
	}

	return string(result)
}
