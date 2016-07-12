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
	"fmt"
	"github.com/fangdingjun/gpp"
	// "github.com/fangdingjun/net/http2"
	"github.com/gorilla/mux"
	"github.com/vharitonsky/iniflags"
	"log"
	//"net"
	"encoding/base64"
	"github.com/fangdingjun/gpp/util"
	"github.com/fangdingjun/handlers"
	"net/http"
	"os"
	"strings"
	"time"
)

// Router is a global router for http.Server
var Router *mux.Router

var docroot string
var proxyUser, proxyPass string

func main() {
	var port int
	var cert string
	var key string
	var logfile string
	var port1 int
	var localDomain string
	var logger *log.Logger
	var proxyAuth bool
	var enableProxy, enableProxyHTTP11 bool
	//var ssl_cert tls.Certificate
	//var listener, listener1 net.Listener
	//var err error
	util.DialTimeout = 2 * time.Second
	//http2.VerboseLogs = false
	var srv, srv1 http.Server

	flag.IntVar(&port, "port", 8000, "the listening port")
	flag.IntVar(&port1, "port2", 0, "the second port")
	flag.StringVar(&cert, "cert", "", "the server certificate")
	flag.StringVar(&key, "key", "", "the private key")
	flag.StringVar(&docroot, "docroot", ".", "the www root directory")
	flag.StringVar(&logfile, "logfile", "", "log file")
	flag.StringVar(&localDomain, "domain", "", "local domain name")
	flag.StringVar(&proxyUser, "proxy_user", "", "proxy username")
	flag.StringVar(&proxyPass, "proxy_pass", "", "proxy password")
	flag.BoolVar(&enableProxy, "enable_proxy", false, "enable proxy support")
	flag.BoolVar(&enableProxyHTTP11, "enable_proxy_http11", false, "when proxy and http2 eanbled, enable proxy on http/1.1")
	flag.BoolVar(&proxyAuth, "proxy_auth", false, "proxy need auth or not")
	iniflags.Parse()

	Router = mux.NewRouter()

	init_routers()

	var out = os.Stdout

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

	srv.Addr = fmt.Sprintf(":%d", port)
	hdr1 := &gpp.Handler{
		Handler:           Router,
		EnableProxy:       enableProxy,
		EnableProxyHTTP11: enableProxyHTTP11,
		LocalDomain:       localDomain,
		Logger:            logger,
		ProxyAuth:         proxyAuth,
		ProxyAuthFunc:     proxyAuthFunc,
	}

	srv.Handler = handlers.CombinedLoggingHandler(out, hdr1)

	srv1.Addr = fmt.Sprintf(":%d", port1)
	hdr2 := &gpp.Handler{
		EnableProxy: false,
		Logger:      logger,
	}

	srv1.Handler = handlers.CombinedLoggingHandler(out, hdr2)

	if port1 != 0 {
		go func() {
			logger.Printf("Listen on http://0.0.0.0:%d", port1)
			log.Fatal(srv1.ListenAndServe())
		}()
	}

	if cert != "" && key != "" {
		//http2.ConfigureServer(&srv, nil)
		logger.Printf("Listen on https://0.0.0.0:%d", port)
		log.Fatal(srv.ListenAndServeTLS(cert, key))
	} else {
		logger.Printf("Listen on http://0.0.0.0:%d", port)
		log.Fatal(srv.ListenAndServe())
	}
}

func proxyAuthFunc(w http.ResponseWriter, r *http.Request) bool {
	user, pass := getBasicUserpass(r)
	if user == "" && pass == "" {
		authFailed(w)
		return false
	}

	if user == proxyUser && pass == proxyPass {
		return true
	}

	authFailed(w)
	return false
}

func authFailed(w http.ResponseWriter) {
	w.Header().Add("Proxy-Authenticate", "Basic realm=\"xxxx.com\"")
	w.WriteHeader(407)
	w.Write([]byte("<h1>unauthenticate</h1>"))
}

func getBasicUserpass(r *http.Request) (string, string) {
	proxyHeader := r.Header.Get("Proxy-Authorization")
	if proxyHeader == "" {
		return "", ""
	}

	r.Header.Del("Proxy-Authorization")

	ss := strings.Split(proxyHeader, " ")
	if len(ss) != 2 {
		return "", ""
	}

	if strings.ToLower(ss[0]) != "basic" {
		return "", ""
	}

	data, err := base64.StdEncoding.DecodeString(ss[1])
	if err != nil {
		log.Printf("%s\n", err.Error())
		return "", ""
	}

	uu := strings.SplitN(string(data), ":", 2)

	return uu[0], uu[1]
}
