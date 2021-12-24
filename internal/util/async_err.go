package util

import "sync"

type AsyncErr struct {
	C  chan error
	wg *sync.WaitGroup
}

func NewAsyncErr() *AsyncErr {
	a := &AsyncErr{
		C:  make(chan error, 1),
		wg: &sync.WaitGroup{},
	}
	a.start()

	return a
}

func (a *AsyncErr) ReaderDone() {
	a.wg.Done()
}

func (a *AsyncErr) WriterDone() {
	a.wg.Done()
}

func (a *AsyncErr) AddWriter() {
	a.wg.Add(1)
}

func (a *AsyncErr) start() {
	a.wg.Add(2)
	go func() {
		a.wg.Wait()
		close(a.C)
	}()
}
