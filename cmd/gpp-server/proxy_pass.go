package main

import (
	"net"
	"net/http"
	//"bufio"
	"io"
	"log"
	"strings"
)

type Proxy struct {
	transport *http.Transport
	addr      string
	//prefix string
}

func NewProxy(addr string) *Proxy {
	p := &Proxy{addr: addr}
	p.transport = &http.Transport{
		Dial: p.Dial,
	}
	/*
	   p.prefix = prefix
	   if p.prefix[len(prefix)-1] == '/'{
	       p.prefix = strings.TrimRight(p.prefix, "/")
	   }
	*/
	return p
}

func (p *Proxy) Dial(network string, addr string) (conn net.Conn, err error) {
	return net.Dial("tcp", p.addr)
}

func (p *Proxy) ProxyPass(w http.ResponseWriter, r *http.Request) {
	r.Header.Add("X-Forwarded-For", strings.Split(r.RemoteAddr, ":")[0])
	r.RequestURI = ""
	r.URL.Scheme = "http"
	r.URL.Host = r.Host
	//r.URL.Path = strings.TrimLeft(r.URL.Path, p.prefix)
	resp, err := p.transport.RoundTrip(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(502)
		w.Write([]byte("<h1>502 Bad Gateway</h1>"))
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
