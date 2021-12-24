package fake_app_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/timotto/ardumower-relay/test/fake_app"
	. "github.com/timotto/ardumower-relay/test/testbed"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("FakeApp", func() {
	Describe("Send", func() {
		It("sends a request to the relay server", func() {
			expectedResponseStatusCode := http.StatusFailedDependency
			expectedResponseBody := "expected response"

			rec := &recordingHandler{
				ResponseStatus: expectedResponseStatusCode,
				ResponseBody:   []byte(expectedResponseBody),
			}
			server := httptest.NewServer(rec)
			defer server.Close()

			givenUsername := "some user"
			givenPassword := "some password"
			uut := NewFakeApp(&Testbed{
				RelayServerUrl: server.URL,
				Username:       givenUsername,
				Password:       givenPassword,
			})

			givenRequestBody := "given body"
			actualResponse, err := uut.Send(givenRequestBody)
			Expect(err).ToNot(HaveOccurred())

			By("sending the request to the relay server", func() {
				Eventually(rec.RequestCount).Should(Equal(1))
			})

			By("providing basic auth creds", func() {
				req := rec.Request(0)
				actualUsername, actualPassword, ok := req.BasicAuth()
				Expect(ok).To(BeTrue())
				Expect(actualUsername).To(Equal(givenUsername))
				Expect(actualPassword).To(Equal(givenPassword))
			})

			By("using the same request content-type as the app", func() {
				req := rec.Request(0)
				Expect(req.Header.Get("content-type")).To(Equal("application/x-www-form-urlencoded; charset=UTF-8"))
			})

			By("returning the response", func() {
				Expect(actualResponse).To(HaveHTTPStatus(expectedResponseStatusCode))
				Expect(actualResponse).To(HaveHTTPBody(expectedResponseBody))
			})
		})
	})
})
