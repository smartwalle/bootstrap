package main

import (
	"context"
	"fmt"
	"github.com/smartwalle/bootstrap"
)

func main() {
	var app = bootstrap.New(&Service{id: "1"}, &Service{id: "2"}, &Service{id: "3"})

	fmt.Println(app.Run())
}

type Service struct {
	id string
}

func (s *Service) Start(ctx context.Context) error {
	fmt.Printf("%s  start \n", s.id)
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	fmt.Printf("%s  stop \n", s.id)
	return nil
}
