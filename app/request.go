package main

import (
	"fmt"
	"log"
	"net"
	"slices"
	"strings"
)

type Request struct {
	method  string
	target  string
	version string
	headers map[string]string
	body    []byte
}

func (req *Request) Parse(buffStr string) error {
	parts := strings.Split(buffStr, "\r\n\r\n")
	if len(parts) == 0 {
		return fmt.Errorf("empty request")
	}

	// headers
	header := strings.Split(parts[0], "\r\n")

	// body
	if len(parts) > 1 {
		req.body = []byte(parts[1])
		log.Printf("[DEBUG] body read: %v", string(req.body))
	}

	// GET /qwe/rty HTTP/1.1
	startLine := strings.Trim(header[0], "\r\n")
	startLineParts := strings.Split(startLine, " ")
	req.method = startLineParts[0]
	req.target = startLineParts[1]
	req.version = startLineParts[2]

	headers := map[string]string{}
	if len(header) > 1 {
		for _, line := range header[1:] {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				headers[parts[0]] = strings.Trim(parts[1], " \r\n")
			}
		}
	}
	req.headers = headers

	log.Printf("[DEBUG] startLineParts: %v", startLineParts)
	log.Printf("[DEBUG] headers: %v", headers)

	return nil
}

func (req *Request) ReadConn(conn net.Conn) error {
	buff := []byte{}
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if n == 0 {
		return err
	}
	if err != nil {
		return err // connection closed
	}
	buff = slices.Concat(buff, buf)
	buffStr := strings.Trim(string(buff), "\x00")

	return req.Parse(buffStr)
}
