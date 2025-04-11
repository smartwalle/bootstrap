package main

import (
	"context"
	"fmt"
	"github.com/smartwalle/bootstrap"
)

func main() {
	var app = bootstrap.New(bootstrap.WithServers(&Server{id: "1"}, &Server{id: "2"}, &Server{id: "3"}))

	fmt.Println(app.Run())
}

type Server struct {
	id string
}

func (s *Server) Start(ctx context.Context) error {
	fmt.Printf("%s  start \n", s.id)
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	fmt.Printf("%s  stop \n", s.id)
	return nil
}
