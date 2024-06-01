package main

import (
	"bufio"
	"fmt"
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

func (s *Server) handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		request, err := s.readPacket(reader)
		if err != nil {
			log.Printf("[ERROR] %v", err)
			continue
		}

		if request.method == "GET" {
			r := NewResponse()
			code := 200
			reason := "OK"
			headers := []string{}
			body := []byte{}
			response := r.format(code, reason, headers, body)
			conn.Write([]byte(response))
		}

		log.Printf("[DEBUG] request parsed: %+v", request)
	}
}

type Packet struct {
	method  string
	request string
	version string
	headers map[string]string
	body    []byte
}

func (s *Server) readPacket(reader *bufio.Reader) (Packet, error) {
	p := Packet{}
	// https://datatracker.ietf.org/doc/html/rfc9112#name-message-format
	// GET /qwe/rty HTTP/1.1
	startLine, err := reader.ReadString('\n')
	if err != nil {
		err := fmt.Errorf("error reading start-line: %w", err)
		return p, err
	}
	startLine = strings.Trim(startLine, "\r\n")
	startLineParts := strings.Split(startLine, " ")

	p.method = startLineParts[0]
	p.request = startLineParts[1]
	p.version = startLineParts[2]

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
			p.headers = headers
			break
		}
		headerLineParts := strings.SplitN(headerLine, ":", 2)
		if len(headerLineParts) == 2 {
			headerLineParts[1] = strings.Trim(headerLineParts[1], "\r\n")
			headers[headerLineParts[0]] = strings.TrimSpace(headerLineParts[1])
		}
	}

	p.body = []byte{}

	log.Printf("[DEBUG] startLineParts: %v", startLineParts)
	log.Printf("[DEBUG] headers: %v", headers)

	return p, nil
}

type Response struct {
	version string
}

func NewResponse() *Response {
	r := Response{
		version: "HTTP/1.1",
	}
	return &r
}

func (r *Response) format(code int, reason string, headers []string, body []byte) string {
	result := ""

	result = fmt.Sprintf("%s%s %d %s\r\n", result, r.version, code, reason)
	result = fmt.Sprintf("%s%s\r\n", result, strings.Join(headers, ""))

	for _, v := range body {
		result += string(v)
	}

	return result
}
