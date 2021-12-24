package monitoring

import (
	"code.cloudfoundry.org/lager"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/timotto/ardumower-relay/internal/util"
	"net"
	"net/http"
	"sync"
)

type monitoring struct {
	logger lager.Logger
	param  Parameters
	listen net.Listener
	wg     *sync.WaitGroup
}

func NewMonitoring(logger lager.Logger, param Parameters) *monitoring {
	return &monitoring{
		logger: logger.Session("monitoring"),
		param:  param,
		wg:     &sync.WaitGroup{},
	}
}

func (m *monitoring) Start(errs *util.AsyncErr) error {
	if !m.param.Enabled {
		return nil
	}

	var err error
	if m.listen, err = net.Listen("tcp", m.param.Address); err != nil {
		return fmt.Errorf("failed to listen at %v: %w", m.param.Address, err)
	}

	m.wg.Add(1)
	errs.AddWriter()
	go func() {
		defer errs.WriterDone()
		defer m.wg.Done()

		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())

		if err := http.Serve(m.listen, mux); err != nil {
			errs.C <- fmt.Errorf("monitoring server failed: %w", err)
		}
	}()

	m.logger.Info("started", lager.Data{"address": m.listen.Addr().String()})

	return nil
}

func (m *monitoring) Stop() {
	if m.listen != nil {
		_ = m.listen.Close()
	}

	m.wg.Wait()
}

func (m *monitoring) Address() string {
	if m.listen != nil {
		return m.listen.Addr().String()
	}

	return ""
}
