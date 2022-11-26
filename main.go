package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/cgi"
	"net/http/fcgi"
	"path/filepath"
)

var useHttp = flag.Bool("http", false, "Use HTTP instead of FCGI")
var port = flag.Int("port", 10101, "Port to bind to for FCGI/HTTP")
var prefix = flag.String("prefix", "/", "Path prefix for the CGI scripts")
var verbose = flag.Bool("v", false, "Verbose mode")

func main() {
	flag.Parse()

	protocolString := "FCGI"
	if *useHttp {
		protocolString = "HTTP"
	}

	path := flag.Arg(0)
	if path == "" {
		log.Fatal("A single argument of the path to the CGI directory is required")
	}

	log.Printf("Starting cgify using %s on port %d serving %s with prefix %s\n", protocolString, *port, path, *prefix)

	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}

	handler := http.StripPrefix(*prefix, makeCgiHandler(path))

	if *useHttp {
		log.Fatal(http.Serve(sock, handler))
	} else {
		log.Fatal(fcgi.Serve(sock, handler))
	}
}

func makeCgiHandler(cgiPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(cgiPath, r.URL.Path)

		var logger *log.Logger = nil
		if *verbose {
			logger = log.Default()
		}

		handler := &cgi.Handler{Path: path, Logger: logger, Root: *prefix}
		handler.ServeHTTP(w, r)
	}
}
