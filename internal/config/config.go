package config

import (
	"code.cloudfoundry.org/lager"
	"github.com/timotto/ardumower-relay/internal/app_endpoint"
	"github.com/timotto/ardumower-relay/internal/auth"
	"github.com/timotto/ardumower-relay/internal/monitoring"
	"github.com/timotto/ardumower-relay/internal/mower_endpoint"
	"github.com/timotto/ardumower-relay/internal/server"
	"strings"
)

type Configuration struct {
	Loglevel string            `yaml:"log"`
	Server   server.Parameters `yaml:"server"`
	Auth     auth.Parameters   `yaml:"auth"`

	AppEndpoint   app_endpoint.Parameters   `yaml:"app_endpoint"`
	MowerEndpoint mower_endpoint.Parameters `yaml:"mower_endpoint"`

	Monitoring monitoring.Parameters `yaml:"monitoring"`
}

func (c *Configuration) Validate() error {
	return validate([]namedValidator{
		{"auth", &c.Auth},
		{"server", &c.Server},
		{"app endpoint", &c.AppEndpoint},
		{"mower endpoint", &c.MowerEndpoint},
	})
}

func (c *Configuration) LogLevel() lager.LogLevel {
	if c == nil {
		return lager.DEBUG
	}

	switch strings.ToLower(c.Loglevel) {
	case "debug":
		return lager.DEBUG

	case "info":
		return lager.INFO

	case "error":
		return lager.ERROR

	case "fatal":
		return lager.FATAL

	default:
		return lager.INFO
	}
}
