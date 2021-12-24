package daemon_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDaemon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Daemon Suite")
}
