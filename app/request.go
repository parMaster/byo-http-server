package main

import (
	"bufio"
	"fmt"
	"log"
	"strings"
)

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
