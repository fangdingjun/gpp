/*
gpp-local is application to convert the https proxy(http proxy over TLS) to a normal http proxy.

Usage

generate a configure file and edit it
    $ gpp-local -dumpflags > client.ini

run it
    $ gpp-local -config client.ini
*/
package main

import (
	"crypto/tls"
	"errors"
	"flag"
	. "fmt"
	"github.com/fangdingjun/net/http2"
	"github.com/vharitonsky/iniflags"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	//"strings"
	//"github.com/fangdingjun/gpp/util"
	"github.com/fangdingjun/handlers"
	"os"
	//"time"
)

var server_name string
var port int = 0

type myhandler struct {
	proxy *proxy
}

type proxy struct {
	handler http.RoundTripper
	index   int
}

func (p *proxy) do(r *http.Request) (*http.Response, error) {
	return p.handler.RoundTrip(r)
}

func (p *proxy) connect(r *http.Request) (*http2.ClientDataConn, error) {
	tr, ok := p.handler.(*http2.Transport)
	if ok {
		return tr.Connect(r)
	}
	return nil, errors.New("wrong http2.Transport")
}

func (p *proxy) dialTLS(network, addr string, cfg *tls.Config) (net.Conn, error) {
	name := addr
	if server_name != "" {
		name = server_name
	}

	c, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	cfg.ServerName = name
	cfg.InsecureSkipVerify = false

	cc := tls.Client(c, cfg)

	err = cc.Handshake()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return cc, nil
}

func (p *proxy) getproxy(req *http.Request) (*url.URL, error) {
	var u *url.URL
	if len(hosts) > 1 {
		j := p.index % len(hosts)
		p.index += 1
		u = &url.URL{
			Scheme: "http",
			Host:   hosts[j],
		}
	} else {
		u = &url.URL{
			Scheme: "http",
			Host:   hosts[0],
		}
	}
	return u, nil
}

func (mhd *myhandler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(503)
		return
	}

	r.Header.Del("proxy-connection")

	s, err := mhd.proxy.connect(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(503)
		return
	}

	w.WriteHeader(s.Res.StatusCode)

	c, _, err := hj.Hijack()
	if err != nil {
		log.Print(err)
		w.WriteHeader(503)
		return
	}

	go func() {
		io.Copy(c, s)
		c.Close()
	}()

	io.Copy(s, c)
	s.Close()
}

func (mhd *myhandler) HandleHttp(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("proxy-connection")
	resp, err := mhd.proxy.do(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(503)
		return
	}

	header := w.Header()
	for k, v := range resp.Header {
		for _, v1 := range v {
			header.Add(k, v1)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	resp.Body.Close()
}

func (mhd *myhandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		mhd.HandleConnect(w, r)
		return
	}

	// local request
	if r.RequestURI[0] == '/' {
		http.DefaultServeMux.ServeHTTP(w, r)
		return
	}

	mhd.HandleHttp(w, r)
}

type myargs []string

func (m *myargs) Set(s string) error {
	*m = append(*m, s)
	return nil
}

func (m *myargs) String() string {
	return ""
}

var hosts myargs
var docroot string

func main() {

	flag.IntVar(&port, "port", 8080, "the port listen to")
	flag.StringVar(&server_name, "server_name", "", "the server name")
	flag.Var(&hosts, "server", "the server connect to")
	flag.StringVar(&docroot, "docroot", ".", "the local http www root")

	iniflags.Parse()

	init_routers()

	if len(hosts) == 0 {
		log.Fatal("you must special a server")
	}
	http2.VerboseLogs = false
	log.Printf("Listening on :%d", port)
	p := &proxy{}
	p.handler = &http2.Transport{
		DialTLS: p.dialTLS,
		Proxy:   p.getproxy,
	}
	hdr := &myhandler{proxy: p}
	err := http.ListenAndServe(Sprintf(":%d", port),
		handlers.CombinedLoggingHandler(os.Stdout, hdr))
	if err != nil {
		log.Fatal(err)
	}
}
