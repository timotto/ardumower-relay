package fake_ardumower_test

import (
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/timotto/ardumower-relay/test/fake_ardumower"
	. "github.com/timotto/ardumower-relay/test/testbed"
	"net/http/httptest"
	"sync/atomic"
)

var _ = Describe("FakeArdumower", func() {
	var rec *websocketHandler
	var server *httptest.Server
	var newTestbed = testbedMaker(&server)
	BeforeEach(func() {
		rec = &websocketHandler{}
		server = httptest.NewServer(rec)
	})
	AfterEach(func() {
		server.Close()
	})
	Describe("Start", func() {
		It("connects to a relay server", func() {
			uut := NewFakeArdumower(newTestbed())
			defer uut.Stop()

			Expect(uut.Start()).ToNot(HaveOccurred())
			Eventually(rec.ConnectionCount).Should(Equal(1))
		})

		It("supplies basic auth creds", func() {
			givenUsername := "a-user"
			givenPassword := "a-password"
			uut := NewFakeArdumower(newTestbed(givenUsername, givenPassword))
			defer uut.Stop()

			Expect(uut.Start()).ToNot(HaveOccurred())
			Eventually(rec.ConnectionCount).Should(Equal(1))
			req := rec.Request(0)
			actualUsername, actualPassword, ok := req.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(actualUsername).To(Equal(givenUsername))
			Expect(actualPassword).To(Equal(givenPassword))
		})

		When("it fails to connect", func() {
			It("returns an error", func() {
				uut := NewFakeArdumower(&Testbed{
					RelayServerUrl: "bad url",
					Username:       "1",
					Password:       "2",
				})

				Expect(uut.Start()).To(HaveOccurred())
			})
		})
	})

	Describe("Behavior", func() {
		var uut *FakeArdumower
		BeforeEach(func() {
			uut = NewFakeArdumower(&Testbed{
				RelayServerUrl: server.URL,
				Username:       "a-user",
				Password:       "a-password",
			})
			uut = NewFakeArdumower(newTestbed())

			Expect(uut.Start()).ToNot(HaveOccurred())

			Eventually(rec.ConnectionCount).Should(Equal(1))
		})
		AfterEach(func() {
			uut.Stop()
		})
		When(`a received message starts with "AT+" and ends with "\n"`, func() {
			It(`responds with "OK=" and the message content after "AT+" including "\n"`, func() {
				received := &atomic.Value{}
				rec.Handler = func(t int, data []byte) (bool, int, []byte) {
					received.Store(string(data))
					return false, 0, nil
				}

				con := rec.Connection(0)
				Expect(con.WriteMessage(websocket.TextMessage, []byte("AT+Hello\n"))).ToNot(HaveOccurred())

				Eventually(received.Load).Should(Equal("OK=Hello\n"))
			})
		})
	})
})

func testbedMaker(server **httptest.Server) func(creds ...string) *Testbed {
	return func(creds ...string) *Testbed {
		Expect(*server).ToNot(BeNil())
		username := "some-username"
		if len(creds) > 0 {
			username = creds[0]

		}
		password := "some password"
		if len(creds) > 1 {
			password = creds[1]
		}

		return &Testbed{
			RelayServerUrl: (*server).URL,
			Username:       username,
			Password:       password,
		}
	}
}
