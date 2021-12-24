package status_endpoint

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"code.cloudfoundry.org/lager"
	"encoding/json"
	"fmt"
	"github.com/timotto/ardumower-relay/internal/model"
	"net/http"
)

type statusEndpoint struct {
	logger lager.Logger
	auth   Auth
	broker Broker
}

//counterfeiter:generate -o fake . Auth
type Auth interface {
	LookupUser(req *http.Request) (model.User, error)
}

//counterfeiter:generate -o fake . Broker
type Broker interface {
	Find(user model.User) (model.Tunnel, bool)
}

func NewStatusEndpoint(logger lager.Logger, auth Auth, broker Broker) *statusEndpoint {
	return &statusEndpoint{
		logger: logger.Session("status-endpoint"),
		auth:   auth,
		broker: broker,
	}
}

func (s *statusEndpoint) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	user, err := s.auth.LookupUser(req)
	if err != nil {
		s.logger.Info("lookup-user", lager.Data{"error": err.Error()})
		w.Header().Set("www-authenticate", `Basic realm="ArduMower Relay"`)
		onError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	tun, found := s.broker.Find(user)
	if !found {
		onError(w, http.StatusNotFound, fmt.Errorf("not connected"))
		return
	}

	stats := tun.Stats()
	bytes, err := json.Marshal(&stats)
	if err != nil {
		onError(w, http.StatusInternalServerError, fmt.Errorf("encode stats failed: %w", err))
		return
	}

	_, _ = w.Write(bytes)
}

func onError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(err.Error()))
}
