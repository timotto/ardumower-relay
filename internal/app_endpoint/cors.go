package app_endpoint

import "net/http"

func (e *appEndpoint) handleCors(w http.ResponseWriter, req *http.Request) bool {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	if req.Method != http.MethodOptions {

		return false
	}

	w.Header().Add("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Add("Access-Control-Allow-Headers", "Authorization")
	w.Header().Add("Access-Control-Max-Age", "86400")

	w.WriteHeader(http.StatusNoContent)

	return true
}
