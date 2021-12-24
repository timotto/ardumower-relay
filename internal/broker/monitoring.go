package broker

import (
	. "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "ardumower_relay"
	subsystem = "broker"
)

type metrics struct {
	countFindSuccess Counter
	countFindError   Counter
	countOpen        Counter
	countReplace     Counter
	countRemove      Counter
	countRemoveStale Counter
}

func (b *broker) setupMonitoring() {
	promauto.NewCounterFunc(counterOpts(
		"modem_connected_count",
		"Count of currently connected ArduMowers"),
		func() float64 {
			b.lock.RLock()
			defer b.lock.RUnlock()

			return float64(len(b.tuns))
		})

	b.mon = metrics{
		countFindSuccess: promauto.NewCounter(counterOpts(
			"op_find_success",
			"Count of Find method calls with result")),
		countFindError: promauto.NewCounter(counterOpts(
			"op_find_failure",
			"Count of Find method calls without result")),
		countOpen: promauto.NewCounter(counterOpts(
			"op_open",
			"Count of Open method calls")),
		countReplace: promauto.NewCounter(counterOpts(
			"op_open_replace",
			"Count of Open method calls replacing an existing tunnel")),
		countRemove: promauto.NewCounter(counterOpts(
			"op_remove",
			"Count of Remove method calls")),
		countRemoveStale: promauto.NewCounter(counterOpts(
			"op_remove_stale",
			"Count of Remove method calls but the tunnel does not exist")),
	}
}

func counterOpts(name, help string) CounterOpts {
	return CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}
}
