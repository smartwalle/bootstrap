package main

import (
	"context"
	"fmt"
	"github.com/smartwalle/bootstrap"
	nhttp "github.com/smartwalle/bootstrap/http"
	"net/http"
)

func main() {
	var mux = http.NewServeMux()
	mux.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		nhttp.NewResponse(1, "hello").Write(writer)
	})

	var httpServer = nhttp.NewServer("127.0.0.1:9090", mux)

	var ctx = NewContext(context.Background(), "这是来自 main 函数的信息")
	var app = bootstrap.New(
		bootstrap.WithContext(ctx),
		bootstrap.WithServers(&SimpleServer{id: "服务A"}, &SimpleServer{id: "服务B"}),
		bootstrap.WithServers(httpServer),
	)
	fmt.Println(app.Run())
}

type SimpleServer struct {
	id string
}

func (s *SimpleServer) Start(ctx context.Context) error {
	fmt.Printf("%s Start: %+v \n", s.id, FromContext(ctx))
	return nil
}

func (s *SimpleServer) Stop(ctx context.Context) error {
	fmt.Printf("%s Stop: %+v \n", s.id, FromContext(ctx))
	return nil
}

type contextKey struct{}

func NewContext(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, contextKey{}, value)
}

func FromContext(ctx context.Context) string {
	var value, _ = ctx.Value(contextKey{}).(string)
	return value
}
