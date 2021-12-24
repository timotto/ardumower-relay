package server

import (
	"net/http"
	"strings"
)

func webSocketHandlerSwitcher(websocketHandler, httpHandler http.Handler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if isGetRequest(req) && hasHeaderValue(req, "Connection", "upgrade") && hasHeader(req, "Upgrade") {
			websocketHandler.ServeHTTP(w, req)
		} else {
			httpHandler.ServeHTTP(w, req)
		}
	}
}

func isGetRequest(r *http.Request) bool {
	return r.Method == http.MethodGet
}

func hasHeaderValue(r *http.Request, k, v string) bool {
	val := r.Header.Get(k)
	return strings.EqualFold(val, v)
}

func hasHeader(r *http.Request, k string) bool {
	return r.Header.Get(k) != ""
}
