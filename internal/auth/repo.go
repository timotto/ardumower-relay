package auth

import "code.cloudfoundry.org/lager"

func newRepo(logger lager.Logger, param Parameters) repo {
	return newPlaintextRepo(logger, param.Filename)
}
