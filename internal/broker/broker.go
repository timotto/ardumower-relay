package broker

import (
	"code.cloudfoundry.org/lager"
	"github.com/timotto/ardumower-relay/internal/model"
	"sync"
)

type broker struct {
	logger lager.Logger
	lock   *sync.RWMutex
	tuns   map[string]model.Tunnel
	mon    metrics
}

func NewBroker(logger lager.Logger) *broker {
	b := &broker{
		logger: logger.Session("broker"),
		lock:   &sync.RWMutex{},
		tuns:   make(map[string]model.Tunnel),
	}
	b.setupMonitoring()

	return b
}

func (b *broker) Find(user model.User) (model.Tunnel, bool) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	tun, exist := b.tuns[user.Id()]
	if exist {
		b.mon.countFindSuccess.Inc()
	} else {
		b.mon.countFindError.Inc()
	}

	return tun, exist
}

func (b *broker) Open(user model.User, tun model.Tunnel) {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.mon.countOpen.Inc()
	if _, exist := b.tuns[user.Id()]; exist {
		_ = b.tuns[user.Id()].Close()
		b.mon.countReplace.Inc()
	}

	b.tuns[user.Id()] = tun
	tun.SetListener(b)
}

func (b *broker) RemoveTunnel(user model.User, tun model.Tunnel) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if t, exist := b.tuns[user.Id()]; exist && t == tun {
		b.mon.countRemove.Inc()
		delete(b.tuns, user.Id())
	} else {
		b.mon.countRemoveStale.Inc()
	}
}
