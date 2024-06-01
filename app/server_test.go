package main

import (
	"bufio"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OKResponse(t *testing.T) {

	r := NewResponse()
	r.code = 200
	r.reason = "OK"
	r.headers = map[string]string{}
	r.body = []byte{}
	assert.Equal(t, "HTTP/1.1 200 OK\r\n\r\n", r.format())
}

func Test_ReadRequest(t *testing.T) {

	r := "GET /qwe/rty HTTP/1.1\r\n"
	r += "Host: localhost:4221\r\n"
	r += "User-Agent: curl/8.4.0\r\n"
	r += "Accept: */*\r\n\r\n"

	reader := bufio.NewReader(strings.NewReader(r))
	request := Request{}
	err := request.Read(reader)

	assert.NoError(t, err)
	assert.Equal(t, "GET", request.method)
	assert.Equal(t, "/qwe/rty", request.target)
	assert.Equal(t, "HTTP/1.1", request.version)
	assert.Equal(t, "localhost:4221", request.headers["Host"])
	assert.Equal(t, "curl/8.4.0", request.headers["User-Agent"])
	assert.Equal(t, "*/*", request.headers["Accept"])
	assert.Equal(t, []byte{}, request.body)
}

func Test_ReadTwoRequests(t *testing.T) {

	r := "GET / HTTP/1.1\r\n"
	r += "Host: localhost:4221\r\n"
	r += "User-Agent: curl/8.4.0\r\n"
	r += "Accept: */*\r\n\r\n"
	r += "GET /qwe/rty HTTP/1.1\r\n"
	r += "Host: localhost:4221\r\n"
	r += "User-Agent: curl/8.4.0\r\n"
	r += "Accept: */*\r\n\r\n"

	reader := bufio.NewReader(strings.NewReader(r))
	request := Request{}

	err := request.Read(reader)
	assert.NoError(t, err)

	err = request.Read(reader)
	assert.NoError(t, err)

}

func Test_404Response(t *testing.T) {

	r := "GET /qwe/rty HTTP/1.1\r\n"
	r += "Host: localhost:4221\r\n"
	r += "User-Agent: curl/8.4.0\r\n"
	r += "Accept: */*\r\n\r\n"

	reader := bufio.NewReader(strings.NewReader(r))
	request := Request{}
	err := request.Read(reader)

	assert.NoError(t, err)
	assert.Equal(t, "GET", request.method)
	assert.Equal(t, "/qwe/rty", request.target)
	assert.Equal(t, "HTTP/1.1", request.version)
	assert.Equal(t, "localhost:4221", request.headers["Host"])
	assert.Equal(t, "curl/8.4.0", request.headers["User-Agent"])
	assert.Equal(t, "*/*", request.headers["Accept"])
	assert.Equal(t, []byte{}, request.body)

	s := NewServer(0)

	response := s.respond(request)
	assert.Equal(t, 404, response.code)
	assert.Equal(t, "Not Found", response.reason)
}

func Test_Echo(t *testing.T) {

	r := "GET /echo/qwe HTTP/1.1\r\n"
	r += "Host: localhost:4221\r\n"
	r += "User-Agent: curl/8.4.0\r\n"
	r += "Accept: */*\r\n\r\n"

	reader := bufio.NewReader(strings.NewReader(r))
	request := Request{}

	err := request.Read(reader)
	assert.NoError(t, err)

	s := NewServer(0)

	response := s.respond(request)
	assert.Equal(t, response.code, 200)
	assert.Equal(t, response.reason, "OK")
	assert.Equal(t, []byte("qwe"), response.body)

	log.Println(response.format())
}
