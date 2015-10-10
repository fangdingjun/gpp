/*
gpp-server is a server application support proxy, file serve  and https, http2.

Usage

use this command to generate a configure file and edit it
    $ gpp-server -dumpflags > server.ini


run it with
    $ gpp-server -config server.ini

use this command to show help message
    $ gpp-server -h

http2 is not enabled if you do not provide the tls certificate and private key file.
*/
package main

import (
	//"crypto/tls"
	"flag"
	. "fmt"
	"github.com/fangdingjun/gpp"
	"github.com/fangdingjun/http2"
	"github.com/gorilla/mux"
	"github.com/vharitonsky/iniflags"
	"log"
	//"net"
	"net/http"
	"os"
)

var Router *mux.Router

var docroot string

func main() {
	var port int
	var cert string
	var key string
	var logfile string
	var port1 int
	var local_domain string
	var logger *log.Logger
	//var ssl_cert tls.Certificate
	//var listener, listener1 net.Listener
	//var err error

	http2.VerboseLogs = false
	var srv, srv1 http.Server

	flag.IntVar(&port, "port", 8000, "the listening port")
	flag.IntVar(&port1, "port2", 0, "the second port")
	flag.StringVar(&cert, "cert", "", "the server certificate")
	flag.StringVar(&key, "key", "", "the private key")
	flag.StringVar(&docroot, "docroot", ".", "the www root directory")
	flag.StringVar(&logfile, "logfile", "", "log file")
	flag.StringVar(&local_domain, "domain", "", "local domain name")
	iniflags.Parse()

	Router = mux.NewRouter()

	init_routers()

	var out *os.File = os.Stdout

	if logfile != "" {
		out1, err := os.Create(logfile)
		if err != nil {
			log.Print(err)
		} else {
			out = out1
		}
	}
	log.SetOutput(out)
	logger = log.New(out, "", log.LstdFlags)

	srv.Addr = Sprintf(":%d", port)
	srv.Handler = &gpp.Handler{
		Handler:     Router,
		EnableProxy: true,
		LocalDomain: local_domain,
		Logger:      logger,
	}

	srv1.Addr = Sprintf(":%d", port1)
	srv1.Handler = &gpp.Handler{EnableProxy: false}

	if port1 != 0 {
		go func() {
			logger.Printf("Listen on http://0.0.0.0:%d", port1)
			log.Fatal(srv1.ListenAndServe())
		}()
	}

	if cert != "" && key != "" {
		http2.ConfigureServer(&srv, nil)
		logger.Printf("Listen on https://0.0.0.0:%d", port)
		log.Fatal(srv.ListenAndServeTLS(cert, key))
	} else {
		logger.Printf("Listen on http://0.0.0.0:%d", port)
		log.Fatal(srv.ListenAndServe())
	}
}
