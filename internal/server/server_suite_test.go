package server_test

import (
	"code.cloudfoundry.org/lager"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/timotto/ardumower-relay/internal/server"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Server Suite")
}

func aLogger() lager.Logger {
	return lager.NewLogger("test")
}

type fakeEndpoint struct {
	Response  string
	CallCount int
}

func (e *fakeEndpoint) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	e.CallCount++

	_, _ = w.Write([]byte(e.Response))
}

func httpUrl(s *server.Server, path string) string {
	address := s.Address()
	//goland:noinspection HttpUrlsUsage
	return fmt.Sprintf("http://%v%v", address, path)
}
