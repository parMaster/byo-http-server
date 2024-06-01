package main

import (
	"log"
	"os"

	"github.com/go-pkgz/lgr"
	"github.com/jessevdk/go-flags"
)

var Options struct {
	Port      int    `long:"port" short:"p" env:"PORT" description:"redis port" default:"4221"`
	Directory string `long:"directory" short:"d" env:"DIRECTORY" description:"directory to serve files from" default:""`
}

func main() {
	if _, err := flags.Parse(&Options); err != nil {
		os.Exit(1)
	}

	// Logger setup
	logOpts := []lgr.Option{
		lgr.LevelBraces,
		lgr.StackTraceOnError,
	}
	logOpts = append(logOpts, lgr.Debug)
	lgr.SetupStdLogger(logOpts...)

	s := NewServer(4221)
	if Options.Directory != "" {
		err := s.WithDirectory(Options.Directory)
		if err != nil {
			log.Printf("[DEBUG] error running with directory: %e", err)
		}
	}
	s.ListenAndServe()

	log.Printf("Server stopped")
}
