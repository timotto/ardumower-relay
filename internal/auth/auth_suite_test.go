package auth_test

import (
	"code.cloudfoundry.org/lager"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auth Suite")
}

func aLogger() lager.Logger {
	return lager.NewLogger("test")
}

func aGetRequest() *http.Request {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	Expect(err).ToNot(HaveOccurred())

	return req
}

func aPostRequest() *http.Request {
	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader("something"))
	Expect(err).ToNot(HaveOccurred())

	return req
}
