package http

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
)

type Server struct {
	*http.Server
	Network string
}

func NewServer(addr string, handler http.Handler) *Server {
	var s = &Server{}
	s.Server = &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	s.Network = "tcp"
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.Handler.ServeHTTP(w, req)
}

func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen(s.Network, s.Addr)
	if err != nil {
		return err
	}

	if s.TLSConfig != nil {
		listener = tls.NewListener(listener, s.TLSConfig)
	}

	s.BaseContext = func(listener net.Listener) context.Context {
		return ctx
	}

	if err = s.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
