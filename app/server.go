package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Server struct {
	addr string
	dir  string
}

func NewServer(port int) *Server {
	s := Server{
		addr: net.JoinHostPort("0.0.0.0", strconv.Itoa(port)),
	}
	return &s
}

func (s *Server) WithDirectory(dir string) error {
	if dir != "" {
		_, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("error finding directory: %w", err)
		}
		s.dir = dir
		return nil
	}
	return fmt.Errorf("directory name is empty")
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", s.addr)
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

		req.target = strings.Trim(req.target, "/")

		urls := strings.Split(req.target, "/")
		if len(urls) > 0 {
			if urls[0] == "echo" && len(urls) > 1 {
				resp.headers["Content-Type"] = "text/plain"
				resp.headers["Content-Length"] = strconv.Itoa(len(urls[1]))
				resp.body = []byte(urls[1])
				return *resp
			}
			if urls[0] == "user-agent" {
				if ua, ok := req.headers["User-Agent"]; ok {
					resp.headers["Content-Type"] = "text/plain"
					resp.headers["Content-Length"] = strconv.Itoa(len(ua))
					resp.body = []byte(ua)
					return *resp
				}
			}
			if urls[0] == "files" && len(urls) > 1 {
				log.Printf("files request: %v", urls)
				fileName := urls[1]
				wd, err := os.Getwd()
				if err != nil {
					resp.code = http.StatusInternalServerError
					resp.reason = "error getting work directory"
				}
				log.Printf("wd: %v, s.dir: %s", wd, s.dir)
				fileName = filepath.Join(s.dir, fileName)
				_, err = os.Stat(fileName)
				if err != nil {
					resp.code = http.StatusNotFound
					resp.reason = "Not Found"
					return *resp
				}
				cont, err := os.ReadFile(fileName)
				if err != nil {
					resp.code = http.StatusInternalServerError
					resp.reason = "error reading file" + fileName
				}
				resp.headers["Content-Type"] = "application/octet-stream"
				resp.headers["Content-Length"] = strconv.Itoa(len(cont))
				resp.body = []byte(cont)
				return *resp
			}
			resp.code = http.StatusBadRequest
			resp.reason = "Bad Request"
		}

		// not found
		resp.code = 404
		resp.reason = "Not Found"
	}

	return *resp
}
