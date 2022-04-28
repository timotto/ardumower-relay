package auth_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/timotto/ardumower-relay/internal/auth"
)

var _ = Describe("Param", func() {
	Describe("Validate", func() {
		Describe("Enabled == true", func() {
			When("no filename is configured", func() {
				It("returns an error", func() {
					uut := Parameters{Enabled: true, Filename: ""}

					Expect(uut.Validate()).To(HaveOccurred())
				})
			})
		})
	})
})
