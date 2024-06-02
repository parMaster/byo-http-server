package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
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
	defer conn.Close()
	request := Request{}
	err := request.ReadConn(conn)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
			log.Printf("[INFO] Connection closed (EOF)")
			return // connection closed
		} else {
			log.Printf("[ERROR] %v", err)
			return
		}
	}
	log.Printf("[DEBUG] request parsed: %+v", request)

	resp := s.respond(request)
	conn.Write([]byte(resp.format()))
}

// primitive router
func (s *Server) respond(req Request) Response {
	resp := NewResponse()

	req.target = strings.Trim(req.target, "/")

	// Accept-Encoding: gzip
	if encodings, ok := req.headers["Accept-Encoding"]; ok {
		acceptEncodings := []string{
			"gzip",
		}

		for _, encoding := range strings.Split(encodings, ",") {
			encoding = strings.TrimSpace(encoding)
			if slices.Contains(acceptEncodings, encoding) {
				resp.headers["Content-Encoding"] = encoding
				break
			}
		}

	}

	if strings.ToUpper(req.method) == "GET" {

		resp.code = 200
		resp.reason = "OK"
		resp.body = []byte{}

		if req.target == "" {
			return *resp
		}

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
				log.Printf("files GET request: %v", urls)
				fileName := urls[1]
				fileName = filepath.Join(s.dir, fileName)
				_, err := os.Stat(fileName)
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

	if strings.ToUpper(req.method) == "POST" {
		resp.code = 200
		resp.reason = "OK"
		resp.headers = map[string]string{}
		resp.body = []byte{}

		urls := strings.Split(req.target, "/")
		if len(urls) > 0 {
			if urls[0] == "files" && len(urls) > 1 {
				log.Printf("[DEBUG] files POST request: %v", urls)
				fileName := urls[1]
				fileName = filepath.Join(s.dir, fileName)
				log.Printf("[DEBUG] writing to file: %v", fileName)

				_, err := os.Stat(s.dir)
				if err != nil {
					err = fmt.Errorf("error stat dir: %w", err)
					log.Printf("[ERROR] %v", err)
					err = os.MkdirAll(s.dir, 0o644)
					err = fmt.Errorf("error mkdirall: %w", err)
					log.Printf("[ERROR] %v", err)
				}

				err = os.WriteFile(fileName, req.body, 0o644)
				if err != nil {
					err = fmt.Errorf("error writing to file: %w", err)
					log.Printf("[ERROR] %v", err)

					resp.code = http.StatusInternalServerError
					resp.reason = "Internal server error"
				}
				resp.code = 201
				resp.reason = "Created"

				return *resp
			}
			resp.code = http.StatusBadRequest
			resp.reason = "Bad Request"
		}
	}

	return *resp
}
