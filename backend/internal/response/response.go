package response

import (
	"encoding/json"
	"net/http"
)

type Envelope struct {
	OK     bool        `json:"ok"`
	Data   interface{} `json:"data,omitempty"`
	Cursor string      `json:"cursor,omitempty"`
	Error  *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func WriteOK(w http.ResponseWriter, status int, data interface{}) {
	write(w, status, Envelope{OK: true, Data: data})
}

func WriteOKWithCursor(w http.ResponseWriter, status int, data interface{}, cursor string) {
	write(w, status, Envelope{OK: true, Data: data, Cursor: cursor})
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	write(w, status, Envelope{OK: false, Error: &APIError{Message: message, Code: code}})
}

func write(w http.ResponseWriter, status int, env Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}
