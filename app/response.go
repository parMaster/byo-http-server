package main

import "fmt"

type Response struct {
	code    int
	reason  string
	headers map[string]string
	body    []byte
	version string
}

func NewResponse() *Response {
	r := Response{
		version: "HTTP/1.1",
	}
	return &r
}

func (r *Response) format() string {
	result := ""

	result = fmt.Sprintf("%s%s %d %s\r\n", result, r.version, r.code, r.reason)
	for k, v := range r.headers {
		result = fmt.Sprintf("%s%s: %s\r\n", result, k, v)
	}
	result += "\r\n"

	result += string(r.body)

	return result
}
