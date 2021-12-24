package fake_ardumower

import (
	"github.com/gorilla/websocket"
	. "github.com/timotto/ardumower-relay/test/testbed"
	"net/http"
	"strings"
)

//FakeArdumower connects to a relay server and behaves similar to an ArduMower
//except it does not use the encryption and different AT commands, responses
type FakeArdumower struct {
	serverUrl string
	username  string
	password  string

	con *websocket.Conn
}

func NewFakeArdumower(bed *Testbed) *FakeArdumower {
	relayServerUrl := bed.RelayServerUrl
	if strings.HasPrefix(relayServerUrl, "http") {
		relayServerUrl = strings.TrimPrefix(relayServerUrl, "http")
		relayServerUrl = "ws" + relayServerUrl
	}

	return &FakeArdumower{
		serverUrl: relayServerUrl,
		username:  bed.Username,
		password:  bed.Password,
	}
}

func (f *FakeArdumower) Start() error {
	req, _ := http.NewRequest(http.MethodGet, "http://localhost/", nil)
	req.SetBasicAuth(f.username, f.password)

	con, _, err := websocket.DefaultDialer.Dial(f.serverUrl, req.Header)
	if err != nil {
		return err
	}

	f.con = con

	go f.reader(con)

	return nil
}

func (f *FakeArdumower) Stop() {
	if f.con != nil {
		_ = f.con.Close()
	}
}

func (f *FakeArdumower) reader(con *websocket.Conn) {
	for {
		t, d, err := con.ReadMessage()
		if err != nil {
			return
		}

		if t != websocket.TextMessage {
			continue
		}

		line := string(d)
		if !strings.HasPrefix(line, "AT+") {
			continue
		}

		if !strings.HasSuffix(line, "\n") {
			continue
		}

		payload := strings.TrimPrefix(line, "AT+")
		response := "OK=" + payload

		_ = con.WriteMessage(websocket.TextMessage, []byte(response))
	}
}
