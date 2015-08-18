package myproxy

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI[0] == '/' {
		http.DefaultServeMux.ServeHTTP(w, r)
		return
	}

	if r.Method == "CONNECT" {
		h.HandleConnect(w, r)
		return
	}

	h.HandleHTTP(w, r)
}

func (h *Handler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		log.Print("connection not support hijack")
		w.WriteHeader(503)
		return
	}

	client_conn, _, err := hj.Hijack()
	if err != nil {
		log.Print(err)
		w.WriteHeader(503)
		return
	}

	server_conn, err := net.Dial("tcp", r.RequestURI)
	if err != nil {
		log.Print("dial to server: ", err)
		client_conn.Write([]byte("HTTP/1.1 503 Service Unaviable\r\n\r\n"))
		return
	}

	client_conn.Write([]byte("HTTP/1.1 200 ok\r\n\r\n"))

	go pipe(server_conn, client_conn)
	go pipe(client_conn, server_conn)
}

func (h *Handler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	r.RequestURI = ""

	r.Header.Del("proxy-connection")

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		log.Print("proxy err: ", err)
		w.WriteHeader(503)
		return
	}

	for k, v := range resp.Header {
		w.Header().Set(k, strings.Join(v, ","))
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)

	resp.Body.Close()
}

func pipe(dst, src net.Conn) {
	io.Copy(dst, src)
	dst.Close()
}
