package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/cgi"
	"net/http/fcgi"
	"os"
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

	cgiPath := flag.Arg(0)
	if cgiPath == "" {
		log.Fatal("A single argument of the path to the CGI directory is required")
	}
	cgiPath, err := filepath.Abs(cgiPath)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(
		"Starting cgify using %s on port %d serving %s with prefix %s\n",
		protocolString,
		*port,
		cgiPath,
		*prefix,
	)

	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}

	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		var path string
		if r.URL.Path != "" {
			path = filepath.Join(cgiPath, r.URL.Path)
		} else {
			path = filepath.Join(cgiPath, "index.html")
		}

		var logger *log.Logger = nil
		if *verbose {
			logger = log.Default()
		}

		stat, err := os.Stat(path)
		if err != nil {
			if logger != nil {
				logger.Print(err.Error())
			}
			http.Error(w, err.Error(), 500)
			return
		}

		if stat.Mode()&0111 != 0 {
			handler := &cgi.Handler{Path: path, Logger: logger, Root: *prefix}
			handler.ServeHTTP(w, r)
		} else {
			if logger != nil {
				log.Printf("Serving static file: %s", path)
			}
			http.ServeFile(w, r, path)
		}
	}

	handler := http.StripPrefix(*prefix, http.HandlerFunc(handlerFunc))

	if *useHttp {
		log.Fatal(http.Serve(sock, handler))
	} else {
		log.Fatal(fcgi.Serve(sock, handler))
	}
}
