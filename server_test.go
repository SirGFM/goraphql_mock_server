package goraphql_mock_server

import (
	"context"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
)

// TestMockServer tests a few GraphQL requests.
func TestMockServer(t *testing.T) {
	// Type that exactly matches the variables and uses an untyped, generic response.
	type CoercedExactResponse struct {
		StringResponse
		ExactVariables
	}

	// Types used by the query.
	type ListFoos struct {
		Foo int `json:"foo"`
	}

	type ListFoosQuery struct {
		ListFoos ListFoos `json:"ListFoos"`
	}

	// Configure the mock server and client.
	s := New()
	defer s.Close()

	s.RegisterQuery("ListFoos", CoercedExactResponse{
		StringResponse: StringResponse(`{
			"ListFoos": {
				"foo": 123
			}
		}`),
		ExactVariables: ExactVariables{
			Variables: EncodedVariable{
				Num: 0,
			},
		},
	})

	s.RegisterQuery("ListFoos", CoercedExactResponse{
		StringResponse: StringResponse(`{
			"ListFoos": {
				"foo": 456
			}
		}`),
		ExactVariables: ExactVariables{
			Variables: EncodedVariable{
				Foo: "foo",
			},
		},
	})

	s.RegisterQuery("ListFoos", SimpleMockedRequest{
		StringResponse: StringResponse(`{
			"ListFoos": {
				"foo": 789
			}
		}`),
		KeyOnlyVariables: KeyOnlyVariables{"foo", "num"},
	})

	client := graphql.NewClient(s.URL())

	// Prepare and run the test cases.
	type testCase struct {
		// The GraphQL request to be sent to the mock server.
		request string
		// The expected response.
		want any
		// Variables used to customize the request.
		variables map[string]any
	}

	const query = `query ($num:Integer!) {
		ListFoos(num:$num) {
			foo
		}
	}`

	testCases := []testCase{{
		request: query,
		want: ListFoosQuery{
			ListFoos: ListFoos{
				Foo: 123,
			},
		},
		variables: map[string]any{
			"num": 0,
		},
	}, {
		request: query,
		want: ListFoosQuery{
			ListFoos: ListFoos{
				Foo: 456,
			},
		},
		variables: map[string]any{
			"foo": "foo",
		},
	}, {
		request: query,
		want: ListFoosQuery{
			ListFoos: ListFoos{
				Foo: 789,
			},
		},
		variables: map[string]any{
			"num": 1,
			"foo": "foo",
		},
	}}

	for _, tc := range testCases {
		req := graphql.NewRequest(tc.request)

		for k, v := range tc.variables {
			req.Var(k, v)
		}

		typ := reflect.TypeOf(tc.want)
		ptr := reflect.New(typ)

		ctx := context.Background()
		err := client.Run(ctx, req, ptr.Interface())
		if assert.NoError(t, err, "failed to send request '%s'", tc.request) {
			assert.Equal(t, tc.want, ptr.Elem().Interface(), "response doesn't match the expected")
		}
	}
}

// TestMockServerCustomAddress checks that it's possible to configure the server to a specific address.
func TestMockServerCustomAddress(t *testing.T) {
	type DummyResponse struct {
		StringResponse
		NoVariable
	}

	s := New(WithAddress("127.0.0.2", 8080, false))
	defer s.Close()

	s.RegisterQuery("ListFoos", DummyResponse{
		StringResponse: StringResponse(`{
			"ListFoos": {
				"foo": 123
			}
		}`),
	})

	type testCase struct {
		// The address where the request should be sent to.
		addr string
		// Whether the request should be received successfully.
		ok bool
	}

	testCases := []testCase{{
		addr: "http://127.0.0.1:8080",
		ok:   false,
	}, {
		addr: "http://127.0.0.2:8080",
		ok:   true,
	}}

	for _, tc := range testCases {
		client := graphql.NewClient(tc.addr)

		req := graphql.NewRequest(
			`query ($num:Integer!) {
				ListFoos(num:$num) {
					foo
				}
			}`,
		)

		var resp map[string]any
		ctx := context.Background()
		err := client.Run(ctx, req, resp)
		if tc.ok {
			assert.NoError(t, err, "failed to send request to %s", tc.addr)
		} else {
			assert.Error(t, err, "request sent successfully to %s although it should have failed", tc.addr)
		}
	}
}

// TestTLSMockServer checks that it's possible to configure the server with TLS communication.
func TestTLSMockServer(t *testing.T) {
	// Suppress logs for this test,
	// re-enabling it afterwards.
	defer func(w io.Writer) {
		log.SetOutput(w)
	}(log.Writer())
	log.SetOutput(io.Discard)

	type DummyResponse struct {
		StringResponse
		NoVariable
	}

	s := New(WithTLS())
	defer s.Close()

	s.RegisterQuery("ListFoos", DummyResponse{
		StringResponse: StringResponse(`{
			"ListFoos": {
				"foo": 123
			}
		}`),
	})

	type testCase struct {
		// The address where the request should be sent to.
		addr string
		// Whether the request should be received successfully.
		ok bool
		// An optional, custom underlying HTTP client that accepts the server's certificates.
		client *http.Client
	}

	testCases := []testCase{{
		addr: "http://" + strings.Replace(s.URL(), "https://", "", 1),
		ok:   false,
	}, {
		addr: s.URL(),
		ok:   false,
	}, {
		addr:   s.URL(),
		ok:     true,
		client: s.Client(),
	}}

	for _, tc := range testCases {
		var opts []graphql.ClientOption

		if tc.client != nil {
			opts = append(opts, graphql.WithHTTPClient(tc.client))
		}
		client := graphql.NewClient(tc.addr, opts...)

		req := graphql.NewRequest(
			`query ($num:Integer!) {
				ListFoos(num:$num) {
					foo
				}
			}`,
		)

		var resp map[string]any
		ctx := context.Background()
		err := client.Run(ctx, req, resp)
		if tc.ok {
			assert.NoError(t, err, "failed to send request to %s", tc.addr)
		} else {
			assert.Error(t, err, "request sent successfully to %s although it should have failed", tc.addr)
		}
	}
}
