package tunnel_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/timotto/ardumower-relay/internal/tunnel"
	"time"
)

var _ = Describe("Parameters", func() {
	Describe("Validate", func() {
		It("sets defaults", func() {
			uut := &tunnel.Parameters{}

			Expect(uut.Validate()).ToNot(HaveOccurred())

			Expect(uut.PingInterval).To(Equal(time.Minute))
			Expect(uut.PingTimeout).To(Equal(10 * time.Second))
			Expect(uut.PongTimeout).To(Equal(10 * time.Second))
		})

		It("keeps existing values", func() {
			uut := &tunnel.Parameters{
				PingInterval: time.Hour,
				PingTimeout:  time.Minute,
				PongTimeout:  time.Millisecond,
			}

			Expect(uut.Validate()).ToNot(HaveOccurred())

			Expect(uut.PingInterval).To(Equal(time.Hour))
			Expect(uut.PingTimeout).To(Equal(time.Minute))
			Expect(uut.PongTimeout).To(Equal(time.Millisecond))
		})
	})
})
