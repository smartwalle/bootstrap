package http

import (
	"encoding/json"
	"net/http"
)

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
