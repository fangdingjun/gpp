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
	"flag"
	. "fmt"
	"github.com/vharitonsky/iniflags"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	//"strings"
	"github.com/fangdingjun/gpp/util"
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

func (p *proxy) connect(r *http.Request) (net.Conn, error) {
	if len(hosts) > 1 {
		j := p.index % len(hosts)
		p.index += 1
		return p.dial("tcp", hosts[j])
	}
	return p.dial("tcp", hosts[0])
}

func (p *proxy) dial(network string, addr string) (net.Conn, error) {
	name := addr
	if server_name != "" {
		name = server_name
	}
	c, err := util.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	cc := tls.Client(c, &tls.Config{ServerName: name})
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

func forward(src, dst net.Conn) {
	io.Copy(dst, src)
	dst.Close()
}

func (mhd *myhandler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(503)
		return
	}

	s, err := mhd.proxy.connect(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(503)
		return
	}

	c, _, err := hj.Hijack()
	if err != nil {
		log.Print(err)
		w.WriteHeader(503)
		return
	}

	r.WriteProxy(s)
	go forward(c, s)
	forward(s, c)
}

func (mhd *myhandler) HandleHttp(w http.ResponseWriter, r *http.Request) {
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

func main() {
	flag.IntVar(&port, "port", 8080, "the port listen to")
	flag.StringVar(&server_name, "server_name", "", "the server name")
	flag.Var(&hosts, "server", "the server connect to")

	iniflags.Parse()

	if len(hosts) == 0 {
		log.Fatal("you must special a server")
	}

	log.Printf("Listening on :%d", port)
	p := &proxy{}
	p.handler = &http.Transport{
		Dial:  p.dial,
		Proxy: p.getproxy,
	}

	err := http.ListenAndServe(Sprintf(":%d", port), &myhandler{proxy: p})
	if err != nil {
		log.Fatal(err)
	}
}
