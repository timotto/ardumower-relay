package mower_endpoint

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"code.cloudfoundry.org/lager"
	"github.com/gorilla/websocket"
	"github.com/timotto/ardumower-relay/internal/model"
	"github.com/timotto/ardumower-relay/internal/tunnel"
	"net/http"
)

type (
	mowerEndpoint struct {
		logger lager.Logger
		param  Parameters
		auth   Auth
		broker Broker

		up *websocket.Upgrader
	}
	Parameters struct {
		ReadBufferSize  int `yaml:"read_buffer_size"`
		WriteBufferSize int `yaml:"write_buffer_size"`

		Tunnel tunnel.Parameters `yaml:"tunnel"`
	}
)

//counterfeiter:generate -o fake . Auth
type Auth interface {
	LookupUser(req *http.Request) (model.User, error)
}

//counterfeiter:generate -o fake . Broker
type Broker interface {
	Open(user model.User, tun model.Tunnel)
}

func NewMowerEndpoint(logger lager.Logger, param Parameters, auth Auth, broker Broker) *mowerEndpoint {
	return &mowerEndpoint{
		logger: logger.Session("mower-endpoint"),
		param:  param,
		auth:   auth,
		broker: broker,

		up: &websocket.Upgrader{
			ReadBufferSize:  param.ReadBufferSize,
			WriteBufferSize: param.WriteBufferSize,
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
	}
}

func (e *mowerEndpoint) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	user, err := e.auth.LookupUser(req)
	if err != nil {
		e.logger.Info("lookup-user", lager.Data{"error": err.Error()})
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := e.up.Upgrade(w, req, nil)
	if err != nil {
		e.logger.Error("upgrade", err)
		return
	}

	mow := tunnel.NewTunnel(e.logger, e.param.Tunnel, conn, user)
	e.broker.Open(user, mow)
}
