package mower_endpoint

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/timotto/ardumower-relay/internal/tunnel"
	"time"
)

var _ = Describe("Parameters", func() {
	It("sets defaults", func() {
		uut := Parameters{}
		Expect(uut.Validate()).ToNot(HaveOccurred())

		Expect(uut.ReadBufferSize).ToNot(BeZero())
		Expect(uut.WriteBufferSize).ToNot(BeZero())

		Expect(uut.Tunnel.PingInterval).ToNot(BeZero())
		Expect(uut.Tunnel.PingTimeout).ToNot(BeZero())
		Expect(uut.Tunnel.PongTimeout).ToNot(BeZero())
	})

	It("does not overwrite values", func() {
		givenReadBufferSize := 111
		givenTunnelPingTimeout := time.Hour

		uut := Parameters{
			ReadBufferSize: givenReadBufferSize,
			Tunnel: tunnel.Parameters{
				PingTimeout: givenTunnelPingTimeout,
			},
		}

		Expect(uut.Validate()).ToNot(HaveOccurred())

		Expect(uut.ReadBufferSize).To(Equal(givenReadBufferSize))
		Expect(uut.Tunnel.PingTimeout).To(Equal(givenTunnelPingTimeout))
	})

})
