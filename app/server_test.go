package main

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OKResponse(t *testing.T) {

	r := NewResponse()
	r.code = 200
	r.reason = "OK"
	r.headers = map[string]string{}
	r.body = []byte{}
	assert.Equal(t, "HTTP/1.1 200 OK\r\n\r\n", r.render())
}

func Test_ReadRequest(t *testing.T) {

	r := "GET /qwe/rty HTTP/1.1\r\n"
	r += "Host: localhost:4221\r\n"
	r += "User-Agent: curl/8.4.0\r\n"
	r += "Accept: */*\r\n\r\n"

	request := Request{}
	err := request.Parse(r)

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

	request := Request{}
	err := request.Parse(r)
	assert.NoError(t, err)

	r = "GET /qwe/rty HTTP/1.1\r\n"
	r += "Host: localhost:4221\r\n"
	r += "User-Agent: curl/8.4.0\r\n"
	r += "Accept: */*\r\n\r\n"
	err = request.Parse(r)
	assert.NoError(t, err)

}

func Test_404Response(t *testing.T) {

	r := "GET /qwe/rty HTTP/1.1\r\n"
	r += "Host: localhost:4221\r\n"
	r += "User-Agent: curl/8.4.0\r\n"
	r += "Accept: */*\r\n\r\n"

	request := Request{}
	err := request.Parse(r)

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

	request := Request{}
	err := request.Parse(r)
	assert.NoError(t, err)

	s := NewServer(0)

	response := s.respond(request)
	assert.Equal(t, response.code, 200)
	assert.Equal(t, response.reason, "OK")
	assert.Equal(t, []byte("qwe"), response.body)

	log.Println(response.render())
}

func Test_SetPort(t *testing.T) {

	s := NewServer(12345)
	assert.Equal(t, "0.0.0.0:12345", s.addr)
}

func Test_WithDirectory(t *testing.T) {
	s := NewServer(0)
	err := s.WithDirectory("")
	assert.Error(t, err)

	os.Mkdir("testdir", 0644)
	err = s.WithDirectory("testdir")
	assert.NoError(t, err)
	os.Remove("testdir")
}

func Test_Files(t *testing.T) {
	s := NewServer(0)

	os.Mkdir("testdir", 0755)
	err := os.WriteFile("testdir/testfile.oct", []byte("content"), 0755)
	assert.NoError(t, err)
	err = s.WithDirectory("testdir")
	assert.NoError(t, err)

	r := "GET /files/testfile.oct HTTP/1.1\r\n"
	r += "Host: localhost:4221\r\n"
	r += "User-Agent: curl/8.4.0\r\n"
	r += "Accept: */*\r\n\r\n"

	request := Request{}
	err = request.Parse(r)
	assert.NoError(t, err)

	response := s.respond(request)
	assert.Equal(t, response.code, 200)
	assert.Equal(t, response.reason, "OK")
	assert.Equal(t, []byte("content"), response.body)

	log.Println(response.render())
}

func Test_PostFiles(t *testing.T) {
	s := NewServer(0)

	os.Mkdir("testdir", 0755)
	err := s.WithDirectory("testdir")
	assert.NoError(t, err)

	r := "POST /files/received_file.txt HTTP/1.1\r\n"
	r += "Host: localhost:4221\r\n"
	r += "User-Agent: curl/8.4.0\r\n"
	r += "Accept: */*\r\n\r\n"
	r += "received"

	request := Request{}
	err = request.Parse(r)
	assert.NoError(t, err)

	response := s.respond(request)
	assert.Equal(t, response.code, 201)
	assert.Equal(t, response.reason, "Created")

	created, err := os.ReadFile("testdir/received_file.txt")
	assert.NoError(t, err)
	assert.Equal(t, "received", string(created))

	log.Println(response.render())
}

func Test_AcceptEncoding(t *testing.T) {
	s := NewServer(0)

	// // valid encoding
	r := "GET /echo/qwe HTTP/1.1\r\n"
	r += "Accept-Encoding: gzip\r\n"

	request := Request{}
	err := request.Parse(r)
	assert.NoError(t, err)

	response := s.respond(request)
	assert.Equal(t, response.code, 200)
	assert.Equal(t, response.reason, "OK")
	strResponse := response.render()
	assert.Contains(t, strResponse, "Content-Encoding: gzip")

	// invalid encoding
	r = "GET /echo/qwe HTTP/1.1\r\n"
	r += "Accept-Encoding: invalid-encoding\r\n"
	err = request.Parse(r)
	assert.NoError(t, err)
	response = s.respond(request)
	assert.Equal(t, response.code, 200)
	assert.Equal(t, response.reason, "OK")
	strResponse = response.render()
	assert.NotContains(t, strResponse, "Content-Encoding")
}

func Test_AcceptMultipleEncoding(t *testing.T) {
	s := NewServer(0)

	// // valid encoding
	r := "GET /echo/qwe HTTP/1.1\r\n"
	r += "Accept-Encoding: encoding-1, gzip, encoding-2\r\n"

	request := Request{}
	err := request.Parse(r)
	assert.NoError(t, err)

	response := s.respond(request)
	assert.Equal(t, response.code, 200)
	assert.Equal(t, response.reason, "OK")
	strResponse := response.render()
	assert.Contains(t, strResponse, "Content-Encoding: gzip")

	// invalid encoding
	r = "GET /echo/qwe HTTP/1.1\r\n"
	r += "Accept-Encoding: invalid-encoding-1, invalid-encoding-2\r\n"
	err = request.Parse(r)
	assert.NoError(t, err)
	response = s.respond(request)
	assert.Equal(t, response.code, 200)
	assert.Equal(t, response.reason, "OK")
	strResponse = response.render()
	assert.NotContains(t, strResponse, "Content-Encoding")
}

func Test_GzipEncoding(t *testing.T) {
	s := NewServer(0)

	// // valid encoding
	r := "GET /echo/foo HTTP/1.1\r\n"
	r += "Accept-Encoding: gzip\r\n"

	request := Request{}
	err := request.Parse(r)
	assert.NoError(t, err)

	response := s.respond(request)
	assert.Equal(t, response.code, 200)
	assert.Equal(t, response.reason, "OK")
	strResponse := response.render()
	assert.Contains(t, strResponse, "Content-Encoding: gzip")

	log.Println(strResponse)
}
