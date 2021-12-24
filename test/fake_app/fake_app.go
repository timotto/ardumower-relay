package fake_app

import (
	. "github.com/timotto/ardumower-relay/test/testbed"
	"net/http"
	"strings"
)

type FakeApp struct {
	serverUrl string
	username  string
	password  string

	http *http.Client
}

func NewFakeApp(bed *Testbed) *FakeApp {
	return &FakeApp{
		serverUrl: bed.RelayServerUrl,
		username:  bed.Username,
		password:  bed.Password,

		http: http.DefaultClient,
	}
}

func (f *FakeApp) Send(body string) (*http.Response, error) {
	reqBody := strings.NewReader(body)
	contentType := "application/x-www-form-urlencoded; charset=UTF-8"
	req, err := http.NewRequest(http.MethodPost, f.serverUrl, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", contentType)
	req.SetBasicAuth(f.username, f.password)

	res, err := f.http.Do(req)

	return res, err
}
