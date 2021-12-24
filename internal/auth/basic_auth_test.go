package auth_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/timotto/ardumower-relay/internal/auth"
	"net/http"
	"os"
	"time"
)

const (
	user1         = "user-1"
	passwordUser1 = "password-1"
	user2         = "user-2"
	passwordUser2 = "password-2"
)

var _ = Describe("BasicAuth", func() {
	var basicAuthFilename string
	var uut Auth
	BeforeEach(func() {
		file, err := os.CreateTemp(os.TempDir(), "ardumower-relay-basic-auth-conf-*")
		Expect(err).ToNot(HaveOccurred())
		_ = file.Close()
		basicAuthFilename = file.Name()

		givenThereAreCredentials(basicAuthFilename, user1, passwordUser1)

		uut = NewAuth(aLogger(), valid(Parameters{Enabled: true, Filename: basicAuthFilename}))
		Expect(uut.Start(nil)).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		uut.Stop()
		_ = os.Remove(basicAuthFilename)
	})

	It("uses the basic auth credentials for the lookup", func() {
		user, err := uut.LookupUser(aGetRequestWithCredentials(user1, passwordUser1))
		Expect(err).ToNot(HaveOccurred())
		Expect(user).ToNot(BeNil())
		Expect(user.Id()).To(Equal(user1))
	})

	When("the credential repository is changed", func() {
		It("does not need a restart", func() {
			_, err := uut.LookupUser(aGetRequestWithCredentials(user2, passwordUser2))
			Expect(err).To(HaveOccurred())

			time.AfterFunc(time.Millisecond, func() {
				givenThereAreCredentials(basicAuthFilename, user2, passwordUser2)
			})

			Eventually(func() error {
				_, err := uut.LookupUser(aGetRequestWithCredentials(user2, passwordUser2))
				return err
			}).ShouldNot(HaveOccurred())

			user, err := uut.LookupUser(aGetRequestWithCredentials(user2, passwordUser2))
			Expect(err).ToNot(HaveOccurred())
			Expect(user).ToNot(BeNil())
			Expect(user.Id()).To(Equal(user2))
		})
	})

	When("the request has no credentials", func() {
		It("returns an error", func() {
			user, err := uut.LookupUser(aGetRequest())
			Expect(user).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(ErrAuthRequired))
		})
	})

	When("the request has bad credentials", func() {
		It("returns an error", func() {
			user, err := uut.LookupUser(aGetRequestWithCredentials(user1, "bad-"+passwordUser1))
			Expect(user).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(ErrUnauthorized))
		})
	})

	When("there's an error reading the file", func() {
		It("returns an error", func() {
			uut := NewAuth(aLogger(), valid(Parameters{Enabled: true, Filename: "."}))
			Expect(uut.Start(nil)).To(HaveOccurred())
			uut.Stop()
		})
	})

	When("there's bad content in the file", func() {
		var badConfigFile string
		BeforeEach(func() {
			file, err := os.CreateTemp(os.TempDir(), "ardumower-relay-basic-auth-bad-conf-*")
			Expect(err).ToNot(HaveOccurred())

			_, err = file.WriteString("good:password\nuser-without-password\n")
			Expect(err).ToNot(HaveOccurred())

			Expect(file.Close()).ToNot(HaveOccurred())
			badConfigFile = file.Name()
		})
		AfterEach(func() {
			_ = os.Remove(badConfigFile)
		})
		It("returns an error", func() {
			uut := NewAuth(aLogger(), valid(Parameters{Enabled: true, Filename: badConfigFile}))
			Expect(uut.Start(nil)).To(HaveOccurred())
			uut.Stop()
		})
	})
})

func aGetRequestWithCredentials(username, password string) *http.Request {
	req := aGetRequest()
	req.SetBasicAuth(username, password)

	return req
}
func givenThereAreCredentials(filename, username, password string) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	Expect(err).ToNot(HaveOccurred())
	defer func() { _ = file.Close() }()
	_, err = file.WriteString(fmt.Sprintf("%s:%s\n", username, password))
	Expect(err).ToNot(HaveOccurred())
}

func valid(p Parameters) Parameters {
	Expect(p.Validate()).ToNot(HaveOccurred())

	return p
}
