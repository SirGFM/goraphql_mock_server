# G(o)raphQL Mock Server

GraphQL Mock server, highly based in [github.com/getoutreach/goql/graphql\_test](https://pkg.go.dev/github.com/getoutreach/goql/graphql_test).

## Quickguide

1. Start a new server:

```go
	s := goraphql_mock_server.New()
	defer s.Close()
```

2. Register the desired operations:

```go
	s.RegisterQuery("ListFoos", goraphql_mock_server.SimpleMockedRequest{
		StringResponse: goraphql_mock_server.StringResponse(`{
			"ListFoos": {
				"foo": 123
			}
		}`),
		KeyOnlyVariables: goraphql_mock_server.KeyOnlyVariables{"foo", "num"},
	})
```

3. Send requests with your preferred GraphQL client to `s.URL()`.

## Customizing responses

A request must implement the interface `goraphql_mock_server.MockedRequest`.
In most cases, simply embedding one of the partial implementations
(for example, `goraphql_mock_server.StringResponse` and `goraphql_mock_server.KeyOnlyVariables`)
should be enough in most cases.

To match variables exactly (e.g., to mock pagination), `goraphql_mock_server.ExactVariables` may be used,
however using a custom type for `Variables` that implementations `goraphql_mock_server.VariableDecoder` is highly advised!
Otherwise, variables may unexpectedly get mismatched depending on how Go decodes values
(and, for example, a number in a map would end up getting decoded as a `float64`).

## Changes from `graphql_test`

* Currently, only `query` is supported
* Variables may be matched identically
* The response may be specified as a string
