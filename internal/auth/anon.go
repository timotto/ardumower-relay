package auth

import (
	"github.com/timotto/ardumower-relay/internal/model"
	"github.com/timotto/ardumower-relay/internal/util"
	"net/http"
)

type anonAuth struct {
}

func newAnonAuth() *anonAuth {
	return &anonAuth{}
}

func (a *anonAuth) LookupUser(_ *http.Request) (model.User, error) {
	return NewUser("anon"), nil
}

func (a *anonAuth) Start(*util.AsyncErr) error {
	return nil
}

func (a *anonAuth) Stop() {
}
