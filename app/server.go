package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/go-pkgz/lgr"
)

func main() {

	// Logger setup
	logOpts := []lgr.Option{
		lgr.LevelBraces,
		lgr.StackTraceOnError,
	}
	logOpts = append(logOpts, lgr.Debug)
	lgr.SetupStdLogger(logOpts...)

	log.Printf("Starting server...")
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	_, err = l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	log.Printf("Server stopped")
}
