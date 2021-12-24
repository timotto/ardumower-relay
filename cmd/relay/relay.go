package main

import (
	"code.cloudfoundry.org/lager"
	"github.com/timotto/ardumower-relay/internal/app_endpoint"
	"github.com/timotto/ardumower-relay/internal/auth"
	"github.com/timotto/ardumower-relay/internal/broker"
	"github.com/timotto/ardumower-relay/internal/config"
	"github.com/timotto/ardumower-relay/internal/daemon"
	"github.com/timotto/ardumower-relay/internal/monitoring"
	"github.com/timotto/ardumower-relay/internal/mower_endpoint"
	"github.com/timotto/ardumower-relay/internal/server"
	"github.com/timotto/ardumower-relay/internal/status_endpoint"
	"os"
)

func main() {
	log, cfg := setup()

	aut := auth.NewAuth(log, cfg.Auth)
	bro := broker.NewBroker(log)

	appEp := app_endpoint.NewAppEndpoint(log, cfg.AppEndpoint, aut, bro)
	mowEp := mower_endpoint.NewMowerEndpoint(log, cfg.MowerEndpoint, aut, bro)
	staEp := status_endpoint.NewStatusEndpoint(log, aut, bro)
	srv := server.NewServer(log, cfg.Server, appEp, mowEp, staEp)
	mon := monitoring.NewMonitoring(log, cfg.Monitoring)

	err := daemon.NewRunner().
		With("auth", aut).
		With("server", srv).
		With("monitoring", mon).
		Run()

	must(log, "run", err)
}

func setup() (lager.Logger, *config.Configuration) {
	cfg, err := config.Get(os.Args)
	log := logger(cfg)
	must(log, "read-config", err)

	return log, cfg
}

func must(l lager.Logger, action string, err error) {
	if err == nil {
		return
	}

	l.Error(action, err)
	os.Exit(1)
}

func logger(cfg *config.Configuration) lager.Logger {
	l := lager.NewLogger("ardumower-relay")
	l.RegisterSink(lager.NewPrettySink(os.Stdout, cfg.LogLevel()))

	return l
}
