package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"strconv"
)

type Response struct {
	code        int
	reason      string
	headers     map[string]string
	body        []byte
	version     string
	compression string
}

func NewResponse() *Response {
	r := Response{
		version: "HTTP/1.1",
		headers: map[string]string{},
	}
	return &r
}

// compress compresses response body with gzip
func (r *Response) compress() {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(r.body)
	w.Close()
	r.body = b.Bytes()
}

// render renders response to string, compresses if needed, adds headers
func (r *Response) render() string {
	result := ""
	result = fmt.Sprintf("%s%s %d %s\r\n", result, r.version, r.code, r.reason)

	if r.compression == "gzip" {
		r.headers["Content-Encoding"] = "gzip"
		r.compress()
	}
	if len(r.body) > 0 {
		r.headers["Content-Length"] = strconv.Itoa(len(r.body))
	}

	for k, v := range r.headers {
		result = fmt.Sprintf("%s%s: %s\r\n", result, k, v)
	}
	result += "\r\n"

	result += string(r.body)

	return result
}
