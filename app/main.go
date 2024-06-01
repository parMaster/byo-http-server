package main

import (
	"log"

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

	s := NewServer(4221)
	s.ListenAndServe()

	log.Printf("Server stopped")
}
