package tunnel_test

import (
	"context"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/timotto/ardumower-relay/internal/model/fake"
	. "github.com/timotto/ardumower-relay/internal/tunnel"
	"time"
)

var _ = Describe("Tunnel", func() {
	var bed *testBed

	BeforeEach(func() {
		bed = NewTestBed()
	})
	AfterEach(func() {
		bed.Close()
	})
	Describe("Transfer", func() {
		It("sends a line", func() {
			givenInput := "expected-content\r\n"

			bed.done.Add(1)
			go func() {
				defer GinkgoRecover()
				defer bed.done.Done()

				uut := NewTunnel(aLogger(), testParameters(), bed.A(), aUser())
				_, _ = uut.Transfer(aContext(), givenInput)
			}()

			Eventually(bed.ReadLineFromB).Should(Equal(givenInput))
		})

		It("receives a line", func() {
			expectedResponse := "expected-response\r\n"

			bed.done.Add(1)
			go func() {
				defer GinkgoRecover()
				defer bed.done.Done()

				_, err := bed.ReadLineFromB()
				Expect(err).ToNot(HaveOccurred())
				Expect(bed.WriteLineIntoB(expectedResponse)).ToNot(HaveOccurred())
			}()

			uut := NewTunnel(aLogger(), testParameters(), bed.A(), aUser())
			actualResponse, err := uut.Transfer(aContext(), "some input")
			Expect(err).ToNot(HaveOccurred())
			Expect(actualResponse).To(Equal(expectedResponse))
		})

		It("forwards errors", func() {
			uut := NewTunnel(aLogger(), testParameters(), bed.A(), aUser())

			bed.Close()

			_, err := uut.Transfer(aContext(), "send into closed tunnel")
			Expect(err).To(HaveOccurred())
		})

		It("ignores websocket messages other than text messages", func() {
			expectedResponse := "expected-response"
			unexpectedResponse := "not-the-" + expectedResponse

			uut := NewTunnel(aLogger(), testParameters(), bed.A(), aUser())

			bed.done.Add(1)
			go func() {
				defer GinkgoRecover()
				defer bed.done.Done()

				_, err := bed.ReadLineFromB()
				Expect(err).ToNot(HaveOccurred())

				err = bed.B().WriteMessage(websocket.BinaryMessage, []byte(unexpectedResponse))
				Expect(err).ToNot(HaveOccurred())

				err = bed.B().WriteMessage(websocket.TextMessage, []byte(expectedResponse))
				Expect(err).ToNot(HaveOccurred())
			}()

			actualResponse, err := uut.Transfer(aContext(), "something")
			Expect(err).ToNot(HaveOccurred())
			Expect(actualResponse).To(Equal(expectedResponse))
		})

		It("respects a canceled Context", func() {
			ctx, cancel := context.WithCancel(aContext())
			givenInput := "expected-content\r\n"

			bed.done.Add(1)
			go func() {
				defer GinkgoRecover()
				defer bed.done.Done()

				Eventually(bed.ReadLineFromB).Should(Equal(givenInput))
				cancel()
			}()

			uut := NewTunnel(aLogger(), testParameters(), bed.A(), aUser())
			_, err := uut.Transfer(ctx, givenInput)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Close", func() {
		It("closes the tunnel", func() {
			uut := NewTunnel(aLogger(), testParameters(), bed.A(), aUser())
			Expect(uut.Close()).ToNot(HaveOccurred())

			_, err := uut.Transfer(aContext(), "some input")
			Expect(err).To(HaveOccurred())
		})
	})

	When("there is an unexpected incoming message", func() {
		It("is dropped", func() {
			expectedResponse := "expected-response\r\n"

			bed.done.Add(1)
			go func() {
				defer GinkgoRecover()
				defer bed.done.Done()

				_, err := bed.ReadLineFromB()
				Expect(err).ToNot(HaveOccurred())
				Expect(bed.WriteLineIntoB(expectedResponse)).ToNot(HaveOccurred())
			}()

			uut := NewTunnel(aLogger(), testParameters(), bed.A(), aUser())

			unexpectedResponse := "other than " + expectedResponse
			Expect(bed.WriteLineIntoB(unexpectedResponse)).ToNot(HaveOccurred())
			time.Sleep(time.Millisecond)

			actualResponse, err := uut.Transfer(aContext(), "some input")
			Expect(err).ToNot(HaveOccurred())
			Expect(actualResponse).To(Equal(expectedResponse))
		})
	})

	Describe("Listener", func() {
		When("the tunnel is closed", func() {
			It("informs the listener", func() {
				givenUser := aUser()
				uut := NewTunnel(aLogger(), testParameters(), bed.A(), givenUser)

				stub := &FakeTunnelListener{}
				uut.SetListener(stub)

				_ = uut.Close()

				Eventually(stub.RemoveTunnelCallCount).Should(Equal(1))
				actualUser, actualTunnel := stub.RemoveTunnelArgsForCall(0)
				Expect(actualTunnel).To(Equal(uut))
				Expect(actualUser).To(Equal(givenUser))
			})
		})
	})
})
