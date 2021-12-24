package fake_app_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFakeApp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FakeApp Suite")
}

type recordingHandler struct {
	ResponseBody   []byte
	ResponseStatus int

	lock   sync.Mutex
	reqs   []*http.Request
	bodies []string
}

func (r *recordingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.lock.Lock()
	defer r.lock.Unlock()

	body := ""
	if req.Method == http.MethodPost {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		body = string(data)
	}

	r.reqs = append(r.reqs, req)
	r.bodies = append(r.bodies, body)

	w.WriteHeader(r.responseStatus())
	_, _ = w.Write(r.ResponseBody)
}

func (r *recordingHandler) RequestCount() int {
	r.lock.Lock()
	defer r.lock.Unlock()

	return len(r.reqs)
}

func (r *recordingHandler) Request(index int) *http.Request {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.reqs[index]
}

func (r *recordingHandler) RequestBody(index int) string {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.bodies[index]
}

func (r *recordingHandler) responseStatus() int {
	if r.ResponseStatus == 0 {
		return http.StatusOK
	}

	return r.ResponseStatus
}

var _ = Describe("recordingHandler", func() {
	var (
		uut        *recordingHandler
		server     *httptest.Server
		newRequest = testServerRequestMaker(&server)
	)
	BeforeEach(func() {
		uut = &recordingHandler{}
		server = httptest.NewServer(uut)
	})
	AfterEach(func() {
		server.Close()
	})

	It("accepts POST requests", func() {
		req := newRequest(http.MethodPost, strings.NewReader("something"))

		expectSuccessfulRequest(req)
	})
	It("accepts OPTIONS requests", func() {
		req := newRequest(http.MethodOptions, nil)

		expectSuccessfulRequest(req)
	})
	It("records requests", func() {
		givenRequestBody := "given request body"
		req1 := newRequest(http.MethodPost, strings.NewReader(givenRequestBody))

		givenHeaderKey := "header-key"
		givenHeaderValue := "given header value"
		req2 := newRequest(http.MethodOptions, nil)
		req2.Header.Set(givenHeaderKey, givenHeaderValue)

		expectSuccessfulRequest(req1)
		expectSuccessfulRequest(req2)

		Eventually(uut.RequestCount).Should(Equal(2))

		actualReq1Body := uut.RequestBody(0)
		Expect(actualReq1Body).To(Equal(givenRequestBody))

		actualReq2 := uut.Request(1)
		Expect(actualReq2.Header.Get(givenHeaderKey)).To(Equal(givenHeaderValue))
	})
	It("can fake error responses", func() {
		expectedResponseCode := http.StatusFailedDependency
		uut.ResponseStatus = expectedResponseCode

		req := newRequest(http.MethodPost, strings.NewReader("something"))

		res, err := http.DefaultClient.Do(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(HaveHTTPStatus(expectedResponseCode))
	})
	It("can fake responses", func() {
		expectedResponseBody := []byte("expected response")
		uut.ResponseBody = expectedResponseBody

		req := newRequest(http.MethodPost, strings.NewReader("something"))

		res, err := http.DefaultClient.Do(req)
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(HaveHTTPBody(expectedResponseBody))
	})
})

func testServerRequestMaker(server **httptest.Server) func(method string, body io.Reader) *http.Request {
	return func(method string, body io.Reader) *http.Request {
		Expect(*server).ToNot(BeNil())
		req, err := http.NewRequest(method, (*server).URL, body)
		Expect(err).ToNot(HaveOccurred())

		return req
	}
}

func expectSuccessfulRequest(req *http.Request) {
	res, err := http.DefaultClient.Do(req)
	Expect(err).ToNot(HaveOccurred())
	Expect(res).To(HaveHTTPStatus(http.StatusOK))
}
