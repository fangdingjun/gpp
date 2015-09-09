package main

import (
	"crypto/tls"
	"flag"
	. "fmt"
	"github.com/fangdingjun/gpp"
	"github.com/fangdingjun/http2"
	"github.com/gorilla/mux"
	"github.com/vharitonsky/iniflags"
	"log"
	"net"
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
	var ssl_cert tls.Certificate
	var listener, listener1 net.Listener
	var err error

	//http2.VerboseLogs = true
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

	srv.Handler = &gpp.Handler{
		Handler:     Router,
		EnableProxy: true,
		LocalDomain: local_domain,
		Logger:      logger,
	}

	srv1.Handler = &gpp.Handler{EnableProxy: false}

	if cert != "" && key != "" {
		ssl_cert, err = tls.LoadX509KeyPair(cert, key)
		if err != nil {
			logger.Fatal(err)
		}
		logger.Printf("Listen on https://0.0.0.0:%d", port)
		http2.ConfigureServer(&srv, nil)
		//log.Println("init http2 support..")
		srv.TLSConfig.Certificates = append(srv.TLSConfig.Certificates,
			ssl_cert)
		listener, err = tls.Listen("tcp", Sprintf(":%d", port),
			srv.TLSConfig)
		if err != nil {
			logger.Fatal(err)
		}
	} else {
		logger.Printf("Listen on http://0.0.0.0:%d", port)
		listener, err = net.Listen("tcp", Sprintf(":%d", port))
		if err != nil {
			logger.Fatal(err)
		}
	}

	if port1 != 0 {
		logger.Printf("Listen on http://0.0.0.0:%d", port1)
		listener1, err = net.Listen("tcp", Sprintf(":%d", port1))
	}

	if port1 != 0 {
		go func() {
			err := srv1.Serve(listener1)
			if err != nil {
				logger.Fatal(err)
			}
		}()
	}

	err = srv.Serve(listener)
	if err != nil {
		logger.Fatal(err)
	}
}
