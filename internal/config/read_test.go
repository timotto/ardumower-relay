package config_test

import (
	"code.cloudfoundry.org/lager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/timotto/ardumower-relay/internal/config"
	"os"
	"time"
)

var _ = Describe("ReadConfig", func() {
	It("reads the config file", func() {
		filename := createAConfigFile(`
log: fatal
server:
  http:
    address: 1.2.3.4:5

app_endpoint:
  timeout: 60s

mower_endpoint:
  read_buffer_size: 16384
  write_buffer_size: 8192
  tunnel:
    ping_interval: 10m
    ping_timeout: 10s
    pong_timeout: 10s
`)
		defer func() { _ = os.Remove(filename) }()

		actualResult, err := config.ReadConfig(filename)
		Expect(err).ToNot(HaveOccurred())

		Expect(actualResult.LogLevel()).To(Equal(lager.FATAL))
		Expect(actualResult.Server.Http.Address).To(Equal("1.2.3.4:5"))

		Expect(actualResult.AppEndpoint.Timeout).To(Equal(time.Minute))
		Expect(actualResult.MowerEndpoint.ReadBufferSize).To(Equal(16384))
		Expect(actualResult.MowerEndpoint.WriteBufferSize).To(Equal(8192))
	})
})

var _ = Describe("Filename", func() {
	When("no extra os args are supplied", func() {
		It("returns the default filename", func() {
			Expect(config.Filename([]string{"something"})).To(Equal("config.yml"))
		})
	})
	When("extra os args are supplied", func() {
		It("returns the first parameter", func() {
			Expect(config.Filename([]string{"something", "expected-value"})).To(Equal("expected-value"))
		})
	})
})

func createAConfigFile(config string) string {
	file, err := os.CreateTemp(os.TempDir(), "ardumower-relay-config-read-test-*")
	Expect(err).ToNot(HaveOccurred())
	defer func() { _ = file.Close() }()

	_, err = file.WriteString(config)
	Expect(err).ToNot(HaveOccurred())
	Expect(file.Close()).ToNot(HaveOccurred())

	return file.Name()
}
