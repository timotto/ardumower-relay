package tunnel_test

import (
	"code.cloudfoundry.org/lager"
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/timotto/ardumower-relay/internal/auth"
	"github.com/timotto/ardumower-relay/internal/model"
	"github.com/timotto/ardumower-relay/internal/tunnel"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTunnel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tunnel Suite")
}

func aLogger() lager.Logger {
	return lager.NewLogger("test")
}

func aContext() context.Context {
	return context.Background()
}

func aUser() model.User {
	return auth.NewUser("test")
}

func testParameters() tunnel.Parameters {
	return tunnel.Parameters{
		PingTimeout:  time.Second,
		PongTimeout:  time.Second,
		PingInterval: time.Minute,
	}
}

type testBed struct {
	a, b  *websocket.Conn
	lock  *sync.Mutex
	ready *sync.WaitGroup
	done  *sync.WaitGroup

	listen net.Listener
}

func NewTestBed() *testBed {
	b := &testBed{
		lock:  &sync.Mutex{},
		ready: &sync.WaitGroup{},
		done:  &sync.WaitGroup{},
	}

	b.start()

	return b
}

func (b *testBed) A() *websocket.Conn {
	return b.a
}

func (b *testBed) B() *websocket.Conn {
	return b.b
}

func (b *testBed) Close() {
	if b.listen != nil {
		_ = b.listen.Close()
	}

	_ = b.a.Close()
	_ = b.b.Close()

	b.done.Wait()
}

func (b *testBed) ReadLineFromB() (string, error) {
	typ, data, err := b.b.ReadMessage()
	Expect(typ).To(Equal(websocket.TextMessage))

	return string(data), err
}

func (b *testBed) WriteLineIntoB(line string) error {
	return b.b.WriteMessage(websocket.TextMessage, []byte(line))
}

func (b *testBed) start() {
	b.server()
	b.client()
	b.ready.Wait()
}

func (b *testBed) server() {
	// Done() in ServeHTTP when client socket has been connected
	b.ready.Add(1)

	var err error
	b.listen, err = net.Listen("tcp", "localhost:0")
	Expect(err).ToNot(HaveOccurred())

	go func() { _ = http.Serve(b.listen, b) }()
}

func (b *testBed) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	up := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	conn, err := up.Upgrade(w, req, nil)
	Expect(err).ToNot(HaveOccurred())

	b.lock.Lock()
	defer b.lock.Unlock()

	Expect(b.a).To(BeNil())
	b.a = conn
	b.ready.Done()
}

func (b *testBed) client() {
	b.ready.Add(1)

	dial := websocket.Dialer{ReadBufferSize: 1024, WriteBufferSize: 1024}

	go func() {
		defer b.ready.Done()
		for {
			conn, _, err := dial.Dial(fmt.Sprintf("ws://%v/", b.listen.Addr().String()), nil)
			if err != nil {
				time.Sleep(100)
				continue
			}

			b.b = conn
			return
		}
	}()
}
