package app_endpoint_test

import (
	"code.cloudfoundry.org/lager"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAppEndpoint(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AppEndpoint Suite")
}

func aLogger() lager.Logger {
	return lager.NewLogger("test")
}
