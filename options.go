package goraphql_mock_server

import (
	"fmt"
	"net"
)

// ServerOptions defines a function used to configure the server.
type ServerOptions func(s *server)

// WithAddress defines a specific address which the mock server should listen on.
func WithAddress(ip string, port uint16, ipv6 bool) ServerOptions {
	return func(s *server) {
		var err error

		if s.server.Listener != nil {
			err = s.server.Listener.Close()
			if err != nil {
				panic(fmt.Sprintf("goraphql_mock_server: failed to close the original listener: %v", err))
			}
		}

		protocol := "tcp"
		if ipv6 {
			protocol = "tcp6"
		}
		addr := fmt.Sprintf("%s:%d", ip, port)

		s.server.Listener, err = net.Listen(protocol, addr)
		if err != nil {
			panic(fmt.Sprintf("goraphql_mock_server: failed to listen on %s: %v", addr, err))
		}
	}
}

// WithTLS causes the mock server to start with TLS enabled.
func WithTLS() ServerOptions {
	return func(s *server) {
		s.useTLS = true
	}
}
