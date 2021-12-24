package daemon_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/timotto/ardumower-relay/internal/daemon"
	. "github.com/timotto/ardumower-relay/internal/model/fake"
	"github.com/timotto/ardumower-relay/internal/util"
	"sync"
	"sync/atomic"
	"time"
)

var _ = Describe("Runner", func() {
	Describe("Run", func() {
		It("starts all the daemons in the order they were added", func() {
			first := &FakeDaemon{}
			second := &FakeDaemon{}
			third := &FakeDaemon{}

			times := &sync.Map{}

			rec := func(name string) func(_ *util.AsyncErr) error {
				return func(err *util.AsyncErr) error {
					times.Store(name, time.Now())

					return nil
				}
			}
			check := func(name string) time.Time {
				if load, ok := times.Load(name); !ok {
					return time.Time{}
				} else if t, ok := load.(time.Time); !ok {
					return time.Time{}
				} else {
					return t
				}
			}

			first.StartCalls(rec("1"))
			second.StartCalls(rec("2"))
			third.StartCalls(rec("3"))

			uut := NewRunner().
				With("z", first).
				With("0", second).
				With("a", third)

			go func() { _ = uut.Run() }()

			Eventually(first.StartCallCount).Should(Equal(1))
			Eventually(second.StartCallCount).Should(Equal(1))
			Eventually(third.StartCallCount).Should(Equal(1))

			Expect(check("1").Before(check("2"))).To(BeTrue())
			Expect(check("2").Before(check("3"))).To(BeTrue())
		})

		When("any daemon fails to start", func() {
			It("stops all the daemons in the reverse order they were added", func() {
				first := &FakeDaemon{}
				second := &FakeDaemon{}
				third := &FakeDaemon{}

				times := &sync.Map{}

				rec := func(name string) func() {
					return func() {
						times.Store(name, time.Now())
					}
				}
				check := func(name string) time.Time {
					if load, ok := times.Load(name); !ok {
						return time.Time{}
					} else if t, ok := load.(time.Time); !ok {
						return time.Time{}
					} else {
						return t
					}
				}

				first.StopCalls(rec("1"))
				second.StopCalls(rec("2"))
				third.StopCalls(rec("3"))

				second.StartReturns(fmt.Errorf("expected error"))

				uut := NewRunner().
					With("z", first).
					With("0", second).
					With("a", third)

				go func() { _ = uut.Run() }()

				Eventually(first.StopCallCount).Should(Equal(1))
				Eventually(second.StopCallCount).Should(Equal(1))
				Eventually(third.StopCallCount).Should(Equal(1))

				Expect(check("1").After(check("2"))).To(BeTrue())
				Expect(check("2").After(check("3"))).To(BeTrue())
			})
		})

		When("any daemon fails asynchronously during runtime", func() {
			It("stops all the daemons in the reverse order they were added", func() {
				first := &FakeDaemon{}
				second := &FakeDaemon{}
				third := &FakeDaemon{}

				times := &sync.Map{}

				rec := func(name string) func() {
					return func() {
						times.Store(name, time.Now())
					}
				}
				check := func(name string) time.Time {
					if load, ok := times.Load(name); !ok {
						return time.Time{}
					} else if t, ok := load.(time.Time); !ok {
						return time.Time{}
					} else {
						return t
					}
				}

				first.StopCalls(rec("1"))
				second.StopCalls(rec("2"))
				third.StopCalls(rec("3"))

				expectedError := fmt.Errorf("expected error")

				second.StartCalls(func(err *util.AsyncErr) error {
					err.AddWriter()
					time.AfterFunc(10*time.Millisecond, func() {
						defer err.WriterDone()
						err.C <- expectedError
					})
					return nil
				})

				uut := NewRunner().
					With("z", first).
					With("0", second).
					With("a", third)

				done := &atomic.Value{}
				go func() {
					err := uut.Run()
					done.Store(err)
				}()

				Eventually(first.StopCallCount).Should(Equal(1))
				Eventually(second.StopCallCount).Should(Equal(1))
				Eventually(third.StopCallCount).Should(Equal(1))

				Expect(check("1").After(check("2"))).To(BeTrue())
				Expect(check("2").After(check("3"))).To(BeTrue())

				Eventually(done.Load).Should(Equal(expectedError))
			})
		})
	})
	Describe("With", func() {
		When("name was used before", func() {
			It("panics", func() {
				uut := NewRunner()
				sameName := "same-name"

				uut.With(sameName, &FakeDaemon{})

				Expect(func() { uut.With(sameName, &FakeDaemon{}) }).To(Panic())
			})
		})
	})
})
