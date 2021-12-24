package model

import "github.com/timotto/ardumower-relay/internal/util"

//counterfeiter:generate -o fake . Daemon
type Daemon interface {
	Start(err *util.AsyncErr) error
	Stop()
}
