package auth

import (
	"code.cloudfoundry.org/lager"
	"github.com/timotto/ardumower-relay/internal/model"
	"github.com/timotto/ardumower-relay/internal/util"
	"net/http"
)

type (
	freeAuth struct {
		logger lager.Logger
	}
)

func newFreeAuth(logger lager.Logger) *freeAuth {
	return &freeAuth{
		logger: logger.Session("free-auth"),
	}
}

func (f freeAuth) LookupUser(req *http.Request) (model.User, error) {
	username, password, ok := req.BasicAuth()
	if !ok {
		return nil, ErrAuthRequired
	}

	return NewUser(username + ":" + password), nil
}

func (f freeAuth) Start(*util.AsyncErr) error {
	return nil
}

func (f freeAuth) Stop() {

}
