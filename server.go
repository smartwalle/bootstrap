package bootstrap

import "context"

type Server interface {
	Start(ctx context.Context) error

	Stop(ctx context.Context) error
}

type ServerWrapper struct {
	server  Server
	running bool
}

func (s *ServerWrapper) Start(ctx context.Context) error {
	s.running = true
	return s.server.Start(ctx)
}

func (s *ServerWrapper) Stop(ctx context.Context) error {
	if !s.running {
		return nil
	}
	s.running = false
	return s.server.Stop(ctx)
}
