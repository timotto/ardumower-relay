package server

import (
	"code.cloudfoundry.org/lager"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/timotto/ardumower-relay/internal/util"
	"net"
	"net/http"
	"sync"
)

type (
	Server struct {
		logger lager.Logger
		param  Parameters
		router *mux.Router

		lHttp, lHttps net.Listener

		stop bool
		wg   *sync.WaitGroup
	}
	Parameters struct {
		Http  HttpParameters  `yaml:"http"`
		Https HttpsParameters `yaml:"https"`
	}
	HttpParameters struct {
		Enabled bool   `yaml:"enabled"`
		Address string `yaml:"address"`
	}
	HttpsParameters struct {
		Enabled  bool   `yaml:"enabled"`
		Address  string `yaml:"address"`
		KeyFile  string `yaml:"key"`
		CertFile string `yaml:"cert"`
	}
)

func NewServer(logger lager.Logger, param Parameters, appEndpoint, mowerEndpoint, statusEndpoint http.Handler) *Server {
	router := mux.NewRouter()

	router.Path("/").Methods(http.MethodGet).HandlerFunc(webSocketHandlerSwitcher(mowerEndpoint, statusEndpoint))
	router.Path("/").Methods(http.MethodPost, http.MethodOptions).Handler(appEndpoint)
	router.Path("/health").HandlerFunc(alwaysHealthy)

	return &Server{
		logger: logger.Session("server"),
		param:  param,
		router: router,

		wg: &sync.WaitGroup{},
	}
}

func (s *Server) Start(errs *util.AsyncErr) error {
	if err := s.param.Validate(); err != nil {
		return err
	}

	if err := s.startHttp(errs); err != nil {
		return fmt.Errorf("failed to start http Server: %w", err)
	}

	if err := s.startHttps(errs); err != nil {
		return fmt.Errorf("failed to start https Server: %w", err)
	}

	s.logStarted()

	return nil
}

func (s *Server) Stop() {
	s.stop = true

	if s.lHttp != nil {
		_ = s.lHttp.Close()
	}

	if s.lHttps != nil {
		_ = s.lHttps.Close()
	}

	s.wg.Wait()
}

func (s *Server) startHttp(errs *util.AsyncErr) error {
	p := s.param.Http
	if !p.Enabled {
		return nil
	}

	serve := func(l net.Listener, h http.Handler) error {
		return http.Serve(l, h)
	}

	var err error
	s.lHttp, err = s.listenAndServe(p.Address, serve, errs)

	return err
}

func (s *Server) startHttps(errs *util.AsyncErr) error {
	p := s.param.Https
	if !p.Enabled {
		return nil
	}

	serve := func(l net.Listener, h http.Handler) error {
		return http.ServeTLS(l, h, p.CertFile, p.KeyFile)
	}

	var err error
	s.lHttps, err = s.listenAndServe(p.Address, serve, errs)

	return err
}

func (s *Server) listenAndServe(address string, serve func(l net.Listener, h http.Handler) error, errs *util.AsyncErr) (net.Listener, error) {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	s.wg.Add(1)
	errs.AddWriter()
	go func() {
		defer errs.WriterDone()
		defer s.wg.Done()

		if err := serve(listen, s.router); err != nil {
			errs.C <- err
		}
	}()

	return listen, nil
}

func alwaysHealthy(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

func (s *Server) logStarted() {
	data := lager.Data{}
	if s.lHttp != nil {
		data["http"] = s.lHttp.Addr().String()
	}

	if s.lHttps != nil {
		data["https"] = s.lHttps.Addr().String()
	}

	s.logger.Info("started", data)
}

func (s *Server) Address() (httpAddress, httpsAddress string) {
	if s.lHttp != nil {
		httpAddress = s.lHttp.Addr().String()
	}

	if s.lHttps != nil {
		httpsAddress = s.lHttps.Addr().String()
	}

	return
}
