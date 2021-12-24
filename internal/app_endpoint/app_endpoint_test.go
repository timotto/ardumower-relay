package app_endpoint_test

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	. "github.com/timotto/ardumower-relay/internal/app_endpoint"
	. "github.com/timotto/ardumower-relay/internal/app_endpoint/fake"
	. "github.com/timotto/ardumower-relay/internal/model/fake"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var _ = Describe("AppEndpoint", func() {
	var (
		auth   *FakeAuth
		broker *FakeBroker
		tunnel *FakeTunnel

		server *httptest.Server
	)
	BeforeEach(func() {
		auth = &FakeAuth{}
		broker = &FakeBroker{}
		tunnel = &FakeTunnel{}

		auth.LookupUserReturns(&FakeUser{}, nil)
		broker.FindReturns(tunnel, true)

		uut := NewAppEndpoint(aLogger(), Parameters{Timeout: 100 * time.Millisecond}, auth, broker)
		server = httptest.NewServer(uut)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("a POST request to /", func() {
		const (
			expectedResponse = "response from the tunnel"
			givenInput       = "input from the http request body"
		)
		var httpResponse *http.Response
		BeforeEach(func() {
			tunnel.TransferReturns(expectedResponse, nil)

			var err error
			httpResponse, err = http.Post(server.URL, "text/plain", strings.NewReader(givenInput))
			Expect(err).ToNot(HaveOccurred())
		})

		It("responds 200/OK",
			expectStatusCode(&httpResponse, http.StatusOK))

		It("sends the request into the tunnel", func() {
			Expect(tunnel.TransferCallCount()).To(Equal(1))
			_, actualInput := tunnel.TransferArgsForCall(0)
			Expect(actualInput).To(Equal(givenInput))
		})

		It("responds to the http request with the response from the tunnel",
			expectRequestBody(&httpResponse, Equal(expectedResponse)))

		It("responds with a wide open CORS header", func() {
			Expect(httpResponse.Header.Get("Access-Control-Allow-Origin")).To(Equal("*"))
		})
	})

	When("looking up the user fails", func() {
		const unexpectedErrorMessage = "sensitive auth error"

		var (
			actualResult *http.Response
		)

		BeforeEach(func() {
			auth.LookupUserReturns(nil, fmt.Errorf(unexpectedErrorMessage))

			var err error
			actualResult, err = http.Post(server.URL, "text/plain", strings.NewReader("user input"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("stops the app from retrying",
			expectFinalAppErrorResponse(&actualResult))

		It("does not report error details",
			expectRequestBody(&actualResult, Not(ContainSubstring(unexpectedErrorMessage))))
	})

	When("there is no ArduMower connected", func() {
		var actualResult *http.Response

		BeforeEach(func() {
			broker.FindReturns(nil, false)

			var err error
			actualResult, err = http.Post(server.URL, "text/plain", strings.NewReader("user input"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("stops the app from retrying",
			expectFinalAppErrorResponse(&actualResult))

		It("reports 'not connected'",
			expectRequestBody(&actualResult, ContainSubstring("not connected")))
	})

	When("the tunnel transfer fails", func() {
		const expectedErrorMessage = "expected tunnel error"
		var (
			httpResponse *http.Response
		)

		BeforeEach(func() {
			tunnel.TransferReturns("", fmt.Errorf(expectedErrorMessage))

			var err error
			httpResponse, err = http.Post(server.URL, "text/plain", strings.NewReader("user input"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("responds with 502/bad gateway",
			expectStatusCode(&httpResponse, http.StatusBadGateway))

		It("reports error details",
			expectRequestBody(&httpResponse, ContainSubstring(expectedErrorMessage)))
	})

	When("the incoming http request is canceled", func() {
		It("cancels the websocket tunnel transfer", func() {
			outerContext, cancelOuterRequest := context.WithCancel(context.Background())
			defer cancelOuterRequest()

			done := &atomic.Value{}
			done.Store(false)

			tunnel.TransferCalls(func(transferContext context.Context, _ string) (string, error) {
				cancelOuterRequest()

				<-transferContext.Done()
				done.Store(true)

				return "", transferContext.Err()
			})

			req, err := http.NewRequestWithContext(outerContext, http.MethodPost, server.URL, strings.NewReader("some input"))
			Expect(err).ToNot(HaveOccurred())

			go func() { _, _ = http.DefaultClient.Do(req) }()

			Eventually(done.Load).Should(BeTrue())
		})
	})

	When("when the tunnel transfer is slower than the configured timeout", func() {
		It("cancels the websocket tunnel transfer", func() {
			done := &atomic.Value{}
			done.Store(false)

			tunnel.TransferCalls(func(transferContext context.Context, _ string) (string, error) {
				<-transferContext.Done()
				done.Store(true)

				return "", transferContext.Err()
			})

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()

				Eventually(done.Load).Should(BeTrue())
			}()

			res, err := http.Post(server.URL, "text/plain", strings.NewReader("some input"))
			Expect(err).ToNot(HaveOccurred())
			Expect(res.StatusCode).To(Equal(http.StatusBadGateway))
			wg.Wait()
		})
	})

	Describe("CORS preflight", func() {
		var httpResponse *http.Response
		BeforeEach(func() {
			req, err := http.NewRequest(http.MethodOptions, server.URL, nil)
			Expect(err).ToNot(HaveOccurred())
			req.Header.Set("access-control-request-headers", "authorization")
			req.Header.Set("access-control-request-method", "POST")

			httpResponse, err = http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
		})

		It("responds 204/No Content",
			expectStatusCode(&httpResponse, http.StatusNoContent))

		It("serves a wide open Access-Control-Allow-Origin HTTP header", func() {
			Expect(httpResponse).To(HaveHTTPHeaderWithValue("Access-Control-Allow-Origin", "*"))
		})

		It("serves an Access-Control-Allow-Methods HTTP header for POST and OPTIONS", func() {
			Expect(httpResponse).To(HaveHTTPHeaderWithValue("Access-Control-Allow-Methods", "POST, OPTIONS"))
		})

		It("serves an Access-Control-Allow-Headers HTTP header for the Authorization header", func() {
			Expect(httpResponse).To(HaveHTTPHeaderWithValue("Access-Control-Allow-Headers", "Authorization"))
		})

		It("serves an Access-Control-Max-Age HTTP header for one hour", func() {
			Expect(httpResponse).To(HaveHTTPHeaderWithValue("Access-Control-Max-Age", "86400"))
		})
	})
})

func expectStatusCode(res **http.Response, expectedStatusCode int) func() {
	return func() {
		Expect((*res).StatusCode).To(Equal(expectedStatusCode))
	}
}

func expectRequestBody(res **http.Response, matcher types.GomegaMatcher) func() {
	return func() {
		defer func() { _ = (*res).Body.Close() }()
		body, err := ioutil.ReadAll((*res).Body)
		Expect(err).ToNot(HaveOccurred())

		Expect(string(body)).To(matcher)
	}
}

func expectFinalAppErrorResponse(res **http.Response) func() {
	return func() {
		defer func() { _ = (*res).Body.Close() }()
		data, err := ioutil.ReadAll((*res).Body)
		Expect(err).ToNot(HaveOccurred())
		body := string(data)

		Expect((*res).StatusCode).To(Equal(http.StatusOK))
		Expect(body).To(HavePrefix("CRC ERROR"))
		Expect(body).To(HaveSuffix("\n"))
	}
}
