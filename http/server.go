package http

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
)

type Server struct {
	*http.Server
}

func NewServer(addr string, handler http.Handler) *Server {
	var s = &Server{}
	s.Server = &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.Handler.ServeHTTP(w, req)
}

func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Printf("HTTP 服务监听失败: %+v\n", err)
		return fmt.Errorf("listen on %s: %w", s.Addr, err)
	}

	if s.TLSConfig != nil {
		listener = tls.NewListener(listener, s.TLSConfig)
	}

	if s.BaseContext == nil {
		s.BaseContext = func(listener net.Listener) context.Context {
			return context.WithoutCancel(ctx)
		}
	}

	log.Printf("HTTP 服务开始运行: %s\n", listener.Addr().String())
	if err = s.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("HTTP 服务运行失败: %+v\n", err)
		return fmt.Errorf("serve http: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	log.Println("HTTP 服务开始关闭")
	if err := s.Server.Shutdown(ctx); err != nil {
		log.Printf("HTTP 服务关闭失败: %+v\n", err)
		return fmt.Errorf("shutdown http server: %w", err)
	}
	log.Println("HTTP 服务关闭完成")
	return nil
}
