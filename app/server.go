package main

import (
	"bufio"
	"errors"
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

// primitive router
func (s *Server) respond(req Request) Response {
	resp := NewResponse()

	if req.method == "GET" {

		resp.code = 200
		resp.reason = "OK"
		resp.headers = map[string]string{}
		resp.body = []byte{}

		if req.target == "/" {
			return *resp
		}

		urls := strings.Split(req.target, "/")
		if len(urls) > 1 {
			if urls[1] == "echo" && len(urls) > 2 {
				resp.headers["Content-Type"] = "text/plain"
				resp.headers["Content-Length"] = strconv.Itoa(len(urls[2]))
				resp.body = []byte(urls[2])
				return *resp
			}
		}

		// not found
		resp.code = 404
		resp.reason = "Not Found"
	}

	return *resp
}
