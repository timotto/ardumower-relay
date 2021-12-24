package fake_ardumower_test

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFakeArdumower(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FakeArdumower Suite")
}

type websocketHandler struct {
	lock    sync.Mutex
	cons    []*websocket.Conn
	reqs    []*http.Request
	Handler func(receivedMessageType int, receivedData []byte) (bool, int, []byte)
	EchoOff bool
}

func (f *websocketHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	up := websocket.Upgrader{
		HandshakeTimeout: time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
	}

	conn, err := up.Upgrade(w, req, nil)
	if err != nil {
		return
	}

	f.add(conn, req)
	go func() {
		defer func() {
			f.remove(conn)
			_ = conn.Close()
		}()

		for {
			receivedMessageType, receivedData, err := conn.ReadMessage()
			if err != nil {
				return
			}

			respond, responseMessageType, responseData := f.handle(receivedMessageType, receivedData)
			if !respond {
				continue
			}

			if err := conn.WriteMessage(responseMessageType, responseData); err != nil {
				return
			}
		}
	}()
}

func (f *websocketHandler) handle(receivedMessageType int, receivedData []byte) (bool, int, []byte) {
	if f.Handler == nil {
		return !f.EchoOff, receivedMessageType, receivedData
	}

	return f.Handler(receivedMessageType, receivedData)
}

func (f *websocketHandler) Stop() {
	f.lock.Lock()
	defer f.lock.Unlock()

	for _, con := range f.cons {
		_ = con.Close()
	}
}

func (f *websocketHandler) ConnectionCount() int {
	f.lock.Lock()
	defer f.lock.Unlock()

	return len(f.cons)
}

func (f *websocketHandler) Connection(index int) *websocket.Conn {
	f.lock.Lock()
	defer f.lock.Unlock()

	return f.cons[index]
}

func (f *websocketHandler) Request(index int) *http.Request {
	f.lock.Lock()
	defer f.lock.Unlock()

	return f.reqs[index]
}

func (f *websocketHandler) add(conn *websocket.Conn, req *http.Request) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.cons = append(f.cons, conn)
	f.reqs = append(f.reqs, req)
}

func (f *websocketHandler) remove(conn *websocket.Conn) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for i, c := range f.cons {
		if c == conn {
			continue
		}

		if i == 0 {
			f.cons = f.cons[1:]
			f.reqs = f.reqs[1:]
		} else {
			f.cons = append(f.cons[:i], f.cons[i+1:]...)
			f.reqs = append(f.reqs[:i], f.reqs[i+1:]...)
		}
		return
	}
}
