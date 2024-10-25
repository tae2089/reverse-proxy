package domain

import (
	"bytes"
	"io"
	"net/http"
)

type ResponseCapture struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (r *ResponseCapture) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseCapture) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func NewResponseCapture(w http.ResponseWriter) *ResponseCapture {
	return &ResponseCapture{
		ResponseWriter: w,
		body:           new(bytes.Buffer),
		statusCode:     http.StatusOK,
	}
}

func (r *ResponseCapture) ToHttpResponse(req *http.Request) *http.Response {
	return &http.Response{
		Status:        http.StatusText(r.statusCode),
		StatusCode:    r.statusCode,
		Body:          io.NopCloser(r.body),
		ContentLength: int64(r.body.Len()),
		Request:       req,
		Header:        r.Header(),
	}
}
