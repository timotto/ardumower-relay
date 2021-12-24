package server_test

import (
	"crypto/tls"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"
)

type fakeMowerEndpoint struct {
	lock     sync.Mutex
	received [][]byte
}

func (f *fakeMowerEndpoint) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	up := websocket.Upgrader{
		HandshakeTimeout: time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
	}

	conn, err := up.Upgrade(w, req, nil)
	if err != nil {
		return
	}

	go func() {
		defer func() { _ = conn.Close() }()
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				return
			}

			if err := conn.WriteMessage(messageType, data); err != nil {
				return
			}

			f.remember(data)
		}
	}()
}

func (f *fakeMowerEndpoint) remember(data []byte) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.received = append(f.received, data)
}

func (f *fakeMowerEndpoint) ReceivedData() [][]byte {
	f.lock.Lock()
	defer f.lock.Unlock()

	return f.received
}

var _ = Describe("fakeMowerEndpoint", func() {
	var (
		uut    *fakeMowerEndpoint
		server *httptest.Server
		con    *websocket.Conn

		newRequest = testServerRequestMaker(&server)
		expectEcho = echoExpector(&con)
	)

	BeforeEach(func() {
		uut = &fakeMowerEndpoint{}
		server = httptest.NewServer(uut)
	})
	AfterEach(func() {
		server.Close()
		if con != nil {
			_ = con.Close()
		}
	})

	It("is a WebSocket echo server", func() {
		con = connectWebsocket(server.URL)

		expectEcho(websocket.TextMessage, "hello 1")

		expectEcho(websocket.BinaryMessage, "hello 2")
	})

	It("records the data received", func() {
		data1 := "hello 1"
		data2 := "hello 2"

		chk := &websocketEchoClient{}
		chk.Check(server.URL, data1, data2)

		Eventually(uut.ReceivedData).Should(HaveLen(2))
		Eventually(uut.ReceivedData).Should(Equal([][]byte{[]byte(data1), []byte(data2)}))
	})

	When("the request is not a WebSocket request", func() {
		It("returns 400/Bad Request", func() {
			requests := []*http.Request{
				newRequest(http.MethodGet, nil),
				newRequest(http.MethodPost, strings.NewReader("{}")),
				newRequest(http.MethodHead, nil),
				newRequest(http.MethodOptions, nil),
			}
			for _, req := range requests {
				res, err := http.DefaultClient.Do(req)
				Expect(err).ToNot(HaveOccurred())

				Expect(res).To(HaveHTTPStatus(http.StatusBadRequest))
			}
		})
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

func echoExpector(con **websocket.Conn) func(messageType int, data string) {
	return func(messageType int, data string) {
		dataSent := []byte(data)
		Expect((*con).WriteMessage(websocket.TextMessage, dataSent)).ToNot(HaveOccurred())

		messageType, dataReceived, err := (*con).ReadMessage()
		Expect(err).ToNot(HaveOccurred())
		Expect(messageType).To(Equal(websocket.TextMessage))

		Expect(dataReceived).To(Equal(dataSent))
	}
}

func connectWebsocket(url string) *websocket.Conn {
	url = strings.Replace(url, "http", "ws", -1)

	dial := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	con, res, err := dial.Dial(url, nil)
	Expect(err).ToNot(HaveOccurred())
	Expect(res).To(HaveHTTPStatus(http.StatusSwitchingProtocols))

	return con
}

type websocketEchoClient struct {
}

func (c *websocketEchoClient) Check(url string, messages ...string) {
	con := connectWebsocket(url)
	defer func() { _ = con.Close() }()

	expectEcho := echoExpector(&con)

	Expect(messages).ToNot(BeEmpty())
	for _, data := range messages {
		expectEcho(websocket.TextMessage, data)
	}
}
