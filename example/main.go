package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/smartwalle/bootstrap"
	nhttp "github.com/smartwalle/bootstrap/http"
)

func main() {
	var mux = http.NewServeMux()
	mux.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		NewResponse(1, "", "hello").Write(http.StatusOK, writer)
	})

	var httpServer = nhttp.NewServer("127.0.0.1:9090", mux)

	var ctx = NewContext(context.Background(), "这是来自 main 函数的信息")

	//ctx, cancel := context.WithCancel(ctx)
	//time.AfterFunc(time.Second*3, cancel)

	var app = bootstrap.New(
		bootstrap.WithContext(ctx),
		bootstrap.WithServers(&SimpleServer{id: "服务A"}, &SimpleServer{id: "服务B"}),
		bootstrap.WithServers(httpServer),
		bootstrap.WithBeforeStart(func(ctx context.Context) {
			fmt.Println("before start")
		}),
		bootstrap.WithAfterStop(func(ctx context.Context) {
			fmt.Println("afte stop")
		}),
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

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewResponse(code int, message string, data interface{}) *Response {
	var r = &Response{}
	r.Code = code
	r.Message = message
	r.Data = data
	return r
}

func (r *Response) Write(statusCode int, writer http.ResponseWriter) error {
	writer.WriteHeader(statusCode)
	return json.NewEncoder(writer).Encode(r)
}
