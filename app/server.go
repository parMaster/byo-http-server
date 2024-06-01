package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

type Server struct {
	Addr string
}

func NewServer(port int) *Server {
	s := Server{
		Addr: net.JoinHostPort("0.0.0.0", strconv.Itoa(port)),
	}
	return &s
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		log.Printf("[INFO] New connection from: %v (%v)", conn.RemoteAddr(), conn)
		go s.handleConnection(conn)
	}
}

// primitive router
func (s *Server) respond(req Request) Response {
	resp := NewResponse()

	if req.method == "GET" {

		resp.code = 200
		resp.reason = "OK"
		resp.headers = map[string]string{}
		resp.body = []byte{}

		// hardcoded routing, 404 for rabdom target
		if req.target != "/" {
			resp.code = 404
			resp.reason = "Not Found"
		}

	}

	return *resp
}

func (s *Server) handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		request := Request{}
		err := request.Read(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("[INFO] Connection closed (EOF)")
				return // connection closed
			} else {
				log.Printf("[ERROR] %v", err)
				continue
			}
		}
		log.Printf("[DEBUG] request parsed: %+v", request)

		resp := s.respond(request)
		conn.Write([]byte(resp.format()))
	}
}

type Request struct {
	method  string
	target  string
	version string
	headers map[string]string
	body    []byte
}

func (req *Request) Read(reader *bufio.Reader) error {
	// https://datatracker.ietf.org/doc/html/rfc9112#name-message-format
	// GET /qwe/rty HTTP/1.1
	startLine, err := reader.ReadString('\n')
	if err != nil {
		err := fmt.Errorf("error reading start-line: %w", err)
		return err
	}
	startLine = strings.Trim(startLine, "\r\n")
	startLineParts := strings.Split(startLine, " ")

	req.method = startLineParts[0]
	req.target = startLineParts[1]
	req.version = startLineParts[2]

	// headers are optional. example:
	// Host: localhost:4221
	// User-Agent: curl/8.4.0
	// Accept: */*
	headers := map[string]string{}
	for {
		headerLine, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("[ERROR] error reading string: %e", err)
		}
		// reached the end of the headers section?
		if headerLine == "\r\n" {
			req.headers = headers
			break
		}
		headerLineParts := strings.SplitN(headerLine, ":", 2)
		if len(headerLineParts) == 2 {
			headerLineParts[1] = strings.Trim(headerLineParts[1], "\r\n")
			headers[headerLineParts[0]] = strings.TrimSpace(headerLineParts[1])
		}
	}

	req.body = []byte{}

	log.Printf("[DEBUG] startLineParts: %v", startLineParts)
	log.Printf("[DEBUG] headers: %v", headers)

	return nil
}

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
	result = fmt.Sprintf("%s\r\n", result)

	// temporary stub
	for _, v := range r.body {
		result += string(v)
	}

	return result
}
