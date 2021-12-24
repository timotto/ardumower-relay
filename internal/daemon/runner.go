package daemon

import (
	"fmt"
	. "github.com/timotto/ardumower-relay/internal/model"
	"github.com/timotto/ardumower-relay/internal/util"
)

type runner struct {
	errs    *util.AsyncErr
	daemons map[string]Daemon
	order   []string
}

func NewRunner() *runner {
	return &runner{
		errs:    util.NewAsyncErr(),
		daemons: make(map[string]Daemon),
	}
}

func (r *runner) With(name string, d Daemon) *runner {
	if _, exists := r.daemons[name]; exists {
		panic("daemon names must be unique")
	}

	r.order = append(r.order, name)
	r.daemons[name] = d

	return r
}

func (r *runner) Run() error {
	defer r.stop()
	defer r.errs.ReaderDone()

	if err := r.start(); err != nil {
		return err
	}

	return <-r.errs.C
}

func (r *runner) start() error {
	for _, name := range r.order {
		d := r.daemons[name]

		if err := d.Start(r.errs); err != nil {
			return fmt.Errorf("failed to start %v: %w", name, err)
		}
	}

	return nil
}

func (r *runner) stop() {
	n := len(r.order)
	for i := n - 1; i >= 0; i-- {
		name := r.order[i]
		daemon := r.daemons[name]
		daemon.Stop()
	}
}
