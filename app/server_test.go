package main

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OKResponse(t *testing.T) {

	r := NewResponse()
	code := 200
	reason := "OK"
	headers := []string{}
	body := []byte{}
	assert.Equal(t, "HTTP/1.1 200 OK\r\n\r\n", r.format(code, reason, headers, body))
}

func Test_ReadPacket(t *testing.T) {

	r := "GET /qwe/rty HTTP/1.1\r\n"
	r += "Host: localhost:4221\r\n"
	r += "User-Agent: curl/8.4.0\r\n"
	r += "Accept: */*\r\n\r\n"

	s := NewServer(0)

	reader := bufio.NewReader(strings.NewReader(r))

	p, err := s.readPacket(reader)
	assert.NoError(t, err)
	assert.Equal(t, "GET", p.method)
	assert.Equal(t, "/qwe/rty", p.request)
	assert.Equal(t, "HTTP/1.1", p.version)
	assert.Equal(t, "localhost:4221", p.headers["Host"])
	assert.Equal(t, "curl/8.4.0", p.headers["User-Agent"])
	assert.Equal(t, "*/*", p.headers["Accept"])
	assert.Equal(t, []byte{}, p.body)

}
