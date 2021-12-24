package auth

import (
	"code.cloudfoundry.org/lager"
	"github.com/timotto/ardumower-relay/internal/model"
	"github.com/timotto/ardumower-relay/internal/util"
	"net/http"
)

type (
	basicAuth struct {
		logger lager.Logger
		param  Parameters
		repo   repo
	}
	repo interface {
		Stop()
		Start() error
		Lookup(username, password string) (model.User, error)
	}
)

func newBasicAuth(logger lager.Logger, param Parameters) *basicAuth {
	logger = logger.Session("basic-auth")

	return &basicAuth{
		logger: logger,
		param:  param,
		repo:   newRepo(logger, param),
	}
}

func (b *basicAuth) LookupUser(req *http.Request) (model.User, error) {
	username, password, ok := req.BasicAuth()
	if !ok {
		return nil, ErrAuthRequired
	}

	return b.repo.Lookup(username, password)
}

func (b *basicAuth) Start(*util.AsyncErr) error {
	return b.repo.Start()
}

func (b *basicAuth) Stop() {
	b.repo.Stop()
}
