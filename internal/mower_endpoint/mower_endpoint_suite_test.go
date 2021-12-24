package mower_endpoint_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMowerEndpoint(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MowerEndpoint Suite")
}
