package app_endpoint

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"code.cloudfoundry.org/lager"
	"context"
	"fmt"
	"github.com/timotto/ardumower-relay/internal/model"
	"io/ioutil"
	"net/http"
)

type appEndpoint struct {
	logger lager.Logger
	param  Parameters
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

func NewAppEndpoint(logger lager.Logger, param Parameters, auth Auth, broker Broker) *appEndpoint {
	return &appEndpoint{
		logger: logger.Session("app-endpoint"),
		param:  param,
		auth:   auth,
		broker: broker,
	}
}

func (e *appEndpoint) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if done := e.handleCors(w, req); done {
		return
	}

	user, err := e.auth.LookupUser(req)
	if err != nil {
		e.logger.Info("lookup-user", lager.Data{"error": err.Error()})
		onFinalError(w, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	reqBody, err := readBody(req)
	if err != nil {
		e.logger.Error("read-body", err)
		onFinalError(w, http.StatusBadRequest, err)
		return
	}

	tun, found := e.broker.Find(user)
	if !found {
		onFinalError(w, http.StatusNotFound, fmt.Errorf("not connected"))
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), e.param.Timeout)
	defer cancel()

	res, err := tun.Transfer(ctx, reqBody)
	if err != nil {
		e.logger.Error("transfer", err)
		onRetryError(w, http.StatusBadGateway, err)
		return
	}

	_, _ = w.Write([]byte(res))
}

func readBody(req *http.Request) (string, error) {
	defer func() { _ = req.Body.Close() }()
	data, err := ioutil.ReadAll(req.Body)

	return string(data), err
}

func onRetryError(w http.ResponseWriter, statusCode int, err error) {
	// When the app receives a response code other than 200/OK it hammers the endpoint with retries without backoff
	w.WriteHeader(statusCode)

	// When the response starts with "CRC ERROR" and terminates with "\n" the app stops calling the endpoint
	res := fmt.Sprintf("ERROR: %v\n", err.Error())
	_, _ = w.Write([]byte(res))
}

func onFinalError(w http.ResponseWriter, statusCode int, err error) {
	// When the app receives a response code other than 200/OK it hammers the endpoint with retries without backoff
	w.WriteHeader(http.StatusOK)

	// When the response starts with "CRC ERROR" and terminates with "\n" the app stops calling the endpoint
	res := fmt.Sprintf("CRC ERROR %v %v\n", statusCode, err.Error())
	_, _ = w.Write([]byte(res))
}
