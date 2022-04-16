package auth

import (
	"code.cloudfoundry.org/lager"
	"fmt"
	"github.com/timotto/ardumower-relay/internal/model"
	"github.com/timotto/ardumower-relay/internal/util"
	"net/http"
)

type Auth interface {
	LookupUser(req *http.Request) (model.User, error)
	Start(*util.AsyncErr) error
	Stop()
}

var (
	ErrUnauthorized = fmt.Errorf("unauthorized")
	ErrAuthRequired = fmt.Errorf("authentication required")
)

func NewAuth(logger lager.Logger, param Parameters) Auth {
	if param.Enabled {
		logger.Info("authentication", lager.Data{"access": "basic-auth"})
		return newBasicAuth(logger, param)
	}

	if param.FreeForAll {
		logger.Info("authentication", lager.Data{"access": "free-for-all"})
		return newFreeAuth(logger)
	}

	logger.Info("authentication", lager.Data{"access": "anonymous"})
	return newAnonAuth()
}
