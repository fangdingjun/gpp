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
	"strings"
	"time"
)

var server_name string
var port int = 0

type myhandler struct{}

var dialer *net.Dialer = &net.Dialer{
	Timeout:   time.Second * 10,
	KeepAlive: time.Second * 10,
}
var i = 0

func createconn() (c net.Conn, err error) {
	//u, _ := url.Parse(proxy)
	if len(hosts) > 1 {
		j := i % len(hosts)
		return dial("tcp", hosts[j])
	}
	return dial("tcp", hosts[0])
}

func dial(network string, addr string) (c net.Conn, err error) {
	if server_name != "" {
		return tls.DialWithDialer(dialer, network, addr, &tls.Config{
			ServerName: server_name,
		})
	}
	return tls.DialWithDialer(dialer, network, addr, nil)
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

	s, err := createconn()
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
	resp, err := DefaultTr.RoundTrip(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(503)
		return
	}
	header := w.Header()
	for k, v := range resp.Header {
		header.Set(k, strings.Join(v, ","))
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

func getproxy(req *http.Request) (*url.URL, error) {
	var u *url.URL
	if len(hosts) > 1 {
		j := i % len(hosts)
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

var DefaultTr http.RoundTripper = &http.Transport{
	Proxy: getproxy,
	Dial:  dial,
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
	err := http.ListenAndServe(Sprintf(":%d", port), &myhandler{})
	if err != nil {
		log.Fatal(err)
	}
}
