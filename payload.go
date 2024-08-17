package goraphql_mock_server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Request maps the received GraphQL request into a go structure.
type Request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

// ResponseError maps an error response into a go structure.
type ResponseError struct {
	Message    string   `json:"message"`
	Path       []string `json:"path"`
	Extensions any      `json:"extensions"`
}

// ResponseError maps a successful response into a go structure.
type Response struct {
	Data   any             `json:"data"`
	Errors []ResponseError `json:"errors,omitempty"`
}

// respondError sends a ResponseError with the specified data and no errors.
func respondError(w http.ResponseWriter, status int, err error, extensions any) {
	res := Response{
		Errors: []ResponseError{{
			Message:    err.Error(),
			Extensions: extensions,
		}},
	}

	respond(w, status, res)
}

// respondResponse sends a Response with the specified data and no errors.
func respondResponse(w http.ResponseWriter, status int, data any) {
	res := Response{
		Data:   data,
		Errors: nil,
	}

	respond(w, status, res)
}

// respond sends a response with the specified status code and payload.
func respond(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		panic(fmt.Sprintf("goraphql_mock_server: failed to encode response: %v", err))
	}
}
