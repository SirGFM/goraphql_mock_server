package goraphql_mock_server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Server manages a mock GraphQL server.
type Server interface {
	// Close closes the underlying http server.
	Close()

	// URL returns the address that a client may use to communicate with the mock server.
	URL() string

	// Client returns an initialized HTTP client configured to accept the server's certificates
	// if running with TLS enabled.
	Client() *http.Client

	// RegisterQuery registers a new query with a specific response.
	//
	// identifier is matched with a simple strings.Contains.
	// So, considering the query:
	//
	// 	query {
	// 		ListObjects(page: $page) {
	// 			data {
	// 				id
	// 			}
	// 		}
	// 	}
	//
	// "ListObjects" would be a valid identifier, as well as "ListObjects(page: $page)".
	// Just be sure to match as much as needed to uniquely identify the request.
	//
	// This may be called multiple types with the same identifier,
	// though note that mocked requests are checked in the same order they were registered.
	// So, the most generic mocked request (e.g., one that embeds a KeyOnlyVariables
	// and that matches the query only on the operation)
	// should be registered last.
	RegisterQuery(identifier string, mock MockedRequest)
}

// server holds the mock GraphQL server, implementing Server for interacting with it.
type server struct {
	// The mocked GraphQL server.
	server *httptest.Server
	// Whether the server should be started with TLS enabled.
	useTLS bool
	// Every registered query in this mocked server.
	queries map[string][]MockedRequest
}

// New starts a new mocked GraphQL server.
//
// To communicate with it, either use the pre-initialized client
// (retrieved by calling Client()) or just use the address returned by URL().
// Mocked requests must be registered by calling RegisterQuery().
//
// Be sure to call Close() when done with the server!
func New(opts ...ServerOptions) Server {
	s := server{
		queries: make(map[string][]MockedRequest),
	}

	var mux http.ServeMux
	mux.HandleFunc("/", s.handler)

	s.server = httptest.NewUnstartedServer(&mux)
	for _, fn := range opts {
		fn(&s)
	}

	if s.useTLS {
		s.server.StartTLS()
	} else {
		s.server.Start()
	}

	return &s
}

// Close implements Server for server.
func (s *server) Close() {
	s.server.Close()
}

// Query implements Server for server.
func (s *server) URL() string {
	return s.server.URL
}

// Client implements Server for server.
func (s *server) Client() *http.Client {
	return s.server.Client()
}

// RegisterQuery implements Server for server.
func (s *server) RegisterQuery(identifier string, mock MockedRequest) {
	tmp := s.queries[identifier]
	tmp = append(tmp, mock)
	s.queries[identifier] = tmp
}

// handler decodes and processes a single GraphQL request.
func (s *server) handler(w http.ResponseWriter, r *http.Request) {
	var reqBody Request
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Errorf("goraphql_mock_server: decode request body: %e", err), nil)
		return
	}

	switch {
	case strings.HasPrefix(strings.TrimSpace(reqBody.Query), "query"):
		for id, mockedRequests := range s.queries {
			if strings.Contains(reqBody.Query, id) {
				for _, mockedRequest := range mockedRequests {
					if s.handleQuery(mockedRequest, reqBody, w) {
						return
					}
				}
			}
		}
	}

	respondError(w, http.StatusNotFound, errors.New("goraphql_mock_server: mocked request not found"), nil)
}

// handleQuery checks if the provided request matches the mocked request,
// sending the mocked response and returning true if they match.
func (s *server) handleQuery(mock MockedRequest, req Request, w http.ResponseWriter) bool {
	if !mock.CompareVariables(req.Variables) {
		return false
	}

	respondResponse(w, http.StatusOK, mock.Response())
	return true
}
