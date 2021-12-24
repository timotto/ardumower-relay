package mower_endpoint_test

import (
	"code.cloudfoundry.org/lager"
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	. "github.com/timotto/ardumower-relay/internal/model/fake"
	. "github.com/timotto/ardumower-relay/internal/mower_endpoint"
	. "github.com/timotto/ardumower-relay/internal/mower_endpoint/fake"
	"github.com/timotto/ardumower-relay/internal/tunnel"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

var _ = Describe("MowerEndpoint", func() {
	var (
		auth   *FakeAuth
		broker *FakeBroker
		user   *FakeUser

		server *httptest.Server
	)
	BeforeEach(func() {
		auth = &FakeAuth{}
		broker = &FakeBroker{}
		user = &FakeUser{}

		auth.LookupUserReturns(user, nil)

		uut := NewMowerEndpoint(aLogger(), testParams(), auth, broker)
		server = httptest.NewServer(uut)
	})
	AfterEach(func() {
		server.Close()
	})
	var wsUrl = func() string {
		return strings.Replace(server.URL, "http", "ws", 1)
	}

	Describe("happy path", func() {
		It("registers the web socket for the user", func() {
			const intoTunnel = "into tunnel"
			const outOfTunnel = "from tunnel"

			con, res, err := websocket.DefaultDialer.Dial(wsUrl(), nil)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = con.Close() }()
			Expect(res.StatusCode).To(Equal(http.StatusSwitchingProtocols))
			go func() {
				defer GinkgoRecover()
				typ, data, err := con.ReadMessage()
				Expect(err).ToNot(HaveOccurred())
				Expect(typ).To(Equal(websocket.TextMessage))
				Expect(string(data)).To(Equal(intoTunnel))

				err = con.WriteMessage(websocket.TextMessage, []byte(outOfTunnel))
				Expect(err).ToNot(HaveOccurred())
			}()

			Eventually(broker.OpenCallCount).Should(Equal(1))
			actualUser, remoteCon := broker.OpenArgsForCall(0)

			Expect(actualUser).To(Equal(user))
			tunnelResponse, err := remoteCon.Transfer(context.Background(), intoTunnel)
			Expect(err).ToNot(HaveOccurred())
			Expect(tunnelResponse).To(Equal(outOfTunnel))
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

		It("responds with 401/unauthorized",
			expectStatusCode(&actualResult, http.StatusUnauthorized))

		It("does not report error details",
			expectRequestBody(&actualResult, Not(ContainSubstring(unexpectedErrorMessage))))
	})

	When("the request is not a websocket request", func() {
		var actualResult *http.Response

		BeforeEach(func() {
			var err error
			actualResult, err = http.Post(server.URL, "text/plain", strings.NewReader("user input"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("responds with 400/bad request",
			expectStatusCode(&actualResult, http.StatusBadRequest))
	})
})

func testParams() Parameters {
	return Parameters{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		Tunnel: tunnel.Parameters{
			PingInterval: time.Minute,
			PingTimeout:  time.Second,
			PongTimeout:  time.Second,
		},
	}
}

func aLogger() lager.Logger {
	return lager.NewLogger("test")
}

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
