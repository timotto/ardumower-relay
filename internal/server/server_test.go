package server_test

import (
	"crypto/tls"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/timotto/ardumower-relay/internal/server"
	"github.com/timotto/ardumower-relay/internal/util"
	"net/http"
	"sync"
)

var _ = Describe("Server", func() {
	var (
		uut    *Server
		params Parameters
		wg     *sync.WaitGroup

		appEndpoint    *fakeEndpoint
		mowerEndpoint  *fakeMowerEndpoint
		statusEndpoint *fakeEndpoint

		client *http.Client
	)
	BeforeEach(func() {
		wg = &sync.WaitGroup{}
		appEndpoint = &fakeEndpoint{Response: "app"}
		mowerEndpoint = &fakeMowerEndpoint{}
		statusEndpoint = &fakeEndpoint{Response: "status"}

		client = &http.Client{}
	})
	Describe("http endpoint", func() {
		var urlFn = httpUrl
		var errs *util.AsyncErr
		BeforeEach(func() {
			errs = util.NewAsyncErr()
			params = Parameters{Http: HttpParameters{Enabled: true, Address: "localhost:0"}}
			uut = NewServer(aLogger(), params, appEndpoint, mowerEndpoint, statusEndpoint)
			Expect(uut.Start(errs)).ToNot(HaveOccurred())

			wg.Add(1)
			go func() {
				defer wg.Done()
				defer errs.ReaderDone()
				<-errs.C
			}()
		})
		AfterEach(func() {
			uut.Stop()
			wg.Wait()
		})

		It("serves the app endpoint at /",
			expectAppEndpoint(&client, &uut, urlFn))

		It("serves the ArduMower endpoint at /",
			expectMowerEndpoint(&client, &uut, urlFn))

		It("serves the status endpoint at /",
			expectStatusEndpoint(&client, &uut, urlFn))

		It("serves a health endpoint at /health",
			expectSpecificEndpoint(http.MethodGet, "/health", "OK", &client, &uut, urlFn))
	})

	Describe("https endpoint", func() {
		var urlFn = httpsUrl
		var errs *util.AsyncErr
		var pki *tempPki
		BeforeEach(func() {
			errs = util.NewAsyncErr()
			pki = newTempPki()

			client.Transport = &http.Transport{TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			}}

			params = Parameters{Https: HttpsParameters{
				Enabled:  true,
				Address:  "localhost:0",
				CertFile: pki.Cert,
				KeyFile:  pki.Key,
			}}
			uut = NewServer(aLogger(), params, appEndpoint, mowerEndpoint, statusEndpoint)
			Expect(uut.Start(errs)).ToNot(HaveOccurred())

			wg.Add(1)
			go func() {
				defer wg.Done()
				defer errs.ReaderDone()
				<-errs.C
			}()
		})
		AfterEach(func() {
			uut.Stop()
			wg.Wait()
			pki.Close()
		})

		It("serves the app endpoint at /",
			expectAppEndpoint(&client, &uut, urlFn))

		It("serves the ArduMower endpoint at /",
			expectMowerEndpoint(&client, &uut, urlFn))

		It("serves the status endpoint at /",
			expectStatusEndpoint(&client, &uut, urlFn))

		It("serves a health endpoint at /health",
			expectSpecificEndpoint(http.MethodGet, "/health", "OK", &client, &uut, urlFn))
	})
})

func expectAppEndpoint(client **http.Client, uut **Server, urlFn func(*Server, string) string) func() {
	post := expectSpecificEndpoint(http.MethodPost, "/", "app", client, uut, urlFn)
	options := expectSpecificEndpoint(http.MethodOptions, "/", "app", client, uut, urlFn)
	return func() {
		post()
		options()
	}
}

func expectMowerEndpoint(_ **http.Client, uut **Server, urlFn func(*Server, string) string) func() {
	return func() {
		url := urlFn(*uut, "/")

		chk := &websocketEchoClient{}
		chk.Check(url, "test1", "test2")
	}
}

func expectStatusEndpoint(client **http.Client, uut **Server, urlFn func(*Server, string) string) func() {
	return expectSpecificEndpoint(http.MethodGet, "/", "status", client, uut, urlFn)

}

func expectSpecificEndpoint(httpMethod, expectedPath, expectedResponse string, client **http.Client, uut **Server, urlFn func(*Server, string) string) func() {
	return func() {
		url := urlFn(*uut, expectedPath)
		req, err := http.NewRequest(httpMethod, url, nil)
		Expect(err).ToNot(HaveOccurred())
		res, err := (*client).Do(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.StatusCode).To(Equal(http.StatusOK))
		Expect(res).To(HaveHTTPBody(expectedResponse))
	}
}
