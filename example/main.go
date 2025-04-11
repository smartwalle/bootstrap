package main

import (
	"context"
	"fmt"
	"github.com/smartwalle/bootstrap"
)

func main() {
	var ctx = NewContext(context.Background(), "hello")
	var app = bootstrap.New(
		bootstrap.WithContext(ctx),
		bootstrap.WithServers(&Server{id: "1"}, &Server{id: "2"}, &Server{id: "3"}),
	)
	fmt.Println(app.Run())
}

type Server struct {
	id string
}

func (s *Server) Start(ctx context.Context) error {
	fmt.Printf("%s  start: %+v \n", s.id, FromContext(ctx))
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	fmt.Printf("%s  stop: %+v \n", s.id, FromContext(ctx))
	return nil
}

type contextKey struct{}

func NewContext(ctx context.Context, value interface{}) context.Context {
	return context.WithValue(ctx, contextKey{}, value)
}

func FromContext(ctx context.Context) interface{} {
	return ctx.Value(contextKey{})
}
