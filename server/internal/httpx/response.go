package httpx

import (
	"encoding/json"
	"net/http"
)

type errorBody struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteJSON writes v as JSON with the given status code. Passing a nil value
// writes only the status (no body).
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError converts err into the standard error envelope. Unknown errors are
// treated as internal errors to avoid leaking details.
func WriteError(w http.ResponseWriter, err error) {
	e := AsError(err)
	if e == nil {
		e = Internal("unexpected error")
	}
	WriteJSON(w, e.Status, errorBody{Error: errorPayload{Code: e.Code, Message: e.Message}})
}
