package broker_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/timotto/ardumower-relay/internal/broker"
	"net/http"
	"net/http/httptest"
	"strings"
)

var _ = Describe("Monitoring", func() {
	BeforeEach(func() {
		resetPrometheus()
	})
	It(`reports open tunnels as "modem_connected_count"`, func() {
		uut := broker.NewBroker(aLogger())
		Expect(metricValue("modem_connected_count")).To(Equal("0"))
		uut.Open(aUser("user-a"), aTunnel())
		Expect(metricValue("modem_connected_count")).To(Equal("1"))
		uut.Open(aUser("user-b"), aTunnel())
		uut.Open(aUser("user-c"), aTunnel())
		Expect(metricValue("modem_connected_count")).To(Equal("3"))
	})
})

func metricValue(name string) string {
	req, err := http.NewRequest(http.MethodGet, "/metrics", nil)
	Expect(err).ToNot(HaveOccurred())

	rec := httptest.NewRecorder()
	promhttp.Handler().ServeHTTP(rec, req)

	Expect(rec.Code).To(Equal(http.StatusOK))
	lines := strings.Split(rec.Body.String(), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}

		if strings.Contains(line, name) {

			parts := strings.Split(line, " ")
			return parts[len(parts)-1]
		}
	}

	return ""
}
