package broker_test

import (
	"code.cloudfoundry.org/lager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/timotto/ardumower-relay/internal/broker"
	"github.com/timotto/ardumower-relay/internal/model"
	. "github.com/timotto/ardumower-relay/internal/model/fake"
)

var _ = Describe("Broker", func() {
	var (
		theUser       model.User
		differentUser model.User
	)
	BeforeEach(func() {
		resetPrometheus()
		theUser = aUser("1")
		differentUser = aUser("2")
	})

	Describe("Open & Find", func() {
		It("stores the given Tunnel for the given User to return it on subsequent Find calls for the same user", func() {
			givenTunnel := aTunnel()
			differentTunnel := aTunnel()

			uut := NewBroker(aLogger())
			uut.Open(theUser, givenTunnel)
			uut.Open(differentUser, differentTunnel)

			actualResult, found := uut.Find(theUser)
			Expect(found).To(BeTrue())
			Expect(actualResult).To(Equal(givenTunnel))

			actualResult, found = uut.Find(differentUser)
			Expect(found).To(BeTrue())
			Expect(actualResult).To(Equal(differentTunnel))
		})
	})

	Describe("Find", func() {
		It("returns false if there is no Tunnel for the given User", func() {
			uut := NewBroker(aLogger())

			_, found := uut.Find(theUser)
			Expect(found).To(BeFalse())
		})
	})

	Describe("Open", func() {
		When("there already is a Tunnel for the given user", func() {
			It("closes the existing tunnel", func() {
				uut := NewBroker(aLogger())

				existingTunnel := aTunnel()
				uut.Open(theUser, existingTunnel)

				newTunnel := aTunnel()
				uut.Open(theUser, existingTunnel)

				Expect(existingTunnel.CloseCallCount()).To(Equal(1))
				Expect(newTunnel.CloseCallCount()).To(Equal(0))
			})
		})

		It("registers the listener", func() {
			tun := aTunnel()

			uut := NewBroker(aLogger())
			uut.Open(theUser, tun)

			Expect(tun.SetListenerCallCount()).To(Equal(1))
			actualListener := tun.SetListenerArgsForCall(0)
			Expect(actualListener).To(Equal(uut))
		})
	})

	Describe("RemoveTunnel", func() {
		It("removes the tunnel", func() {
			uut := NewBroker(aLogger())

			existingTunnel := aTunnel()
			uut.Open(theUser, existingTunnel)

			uut.RemoveTunnel(theUser, existingTunnel)

			actualResult, ok := uut.Find(theUser)
			Expect(ok).To(BeFalse())
			Expect(actualResult).To(BeNil())
		})
		It("does not panic if the tunnel does not exist", func() {
			uut := NewBroker(aLogger())

			staleTunnel := aTunnel()
			uut.RemoveTunnel(theUser, staleTunnel)

			_, ok := uut.Find(theUser)
			Expect(ok).To(BeFalse())
		})
	})
})

func aUser(id string) model.User {
	u := &FakeUser{}
	u.IdReturns(id)

	return u
}

func aTunnel() *FakeTunnel {
	return &FakeTunnel{}
}

func aLogger() lager.Logger {
	return lager.NewLogger("test")
}
