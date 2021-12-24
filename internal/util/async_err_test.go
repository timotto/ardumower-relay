package util_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/timotto/ardumower-relay/internal/util"
)

var _ = Describe("AsyncErr", func() {
	It("creates a buffered channel for errors", func() {
		uut := NewAsyncErr()

		givenError := fmt.Errorf("something")

		// write without reader
		uut.C <- givenError

		// read from buffer
		actualResult := <-uut.C

		Expect(actualResult).To(Equal(givenError))
	})

	It("closes the channel when Reader & all Writers are done", func() {
		uut := NewAsyncErr()

		uut.AddWriter()

		uut.ReaderDone()
		uut.WriterDone()
		uut.WriterDone()

		_, open := <-uut.C

		Expect(open).To(BeFalse())
	})
})
