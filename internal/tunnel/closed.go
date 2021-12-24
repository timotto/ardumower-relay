package tunnel

func (t *tunnel) setClosed() {
	t.clLock.Lock()
	defer t.clLock.Unlock()

	t.closed = true
}

func (t *tunnel) isClosed() bool {
	t.clLock.Lock()
	defer t.clLock.Unlock()

	return t.closed
}
