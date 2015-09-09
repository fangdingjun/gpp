package gpp

import (
	. "fmt"
	"github.com/fangdingjun/gpp/util"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

type Handler struct {
	Handler     http.Handler
	EnableProxy bool
	LocalDomain string
	Transport   http.RoundTripper
	Logger      *log.Logger
}

/* log request */
func (h *Handler) LogReq(r *http.Request, status int) {
	if h.Logger != nil {
		ua := r.Header.Get("user-agent")
		if ua == "" {
			ua = "-"
		}

		uri := r.RequestURI

		if r.ProtoMajor == 2 {
			uri = r.URL.String()
		}

		if r.Method == "CONNECT" {
			uri = r.URL.Host
		}

		ip := strings.Split(r.RemoteAddr, ":")[0]

		h.Logger.Printf("%s \"%s %s %s\" %03d \"%s\" ",
			ip, r.Method, uri, r.Proto, status, ua,
		)
	}
}

/* log */
func (h *Handler) Log(format string, args ...interface{}) {
	if h.Logger != nil {
		h.Logger.Printf(format, args...)
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if b := h.is_local_request(r); b {
		h.LogReq(r, 0)
		if h.Handler != nil {
			/* invoke handler */
			h.Handler.ServeHTTP(w, r)
			return
		}

		/* invoke default handler */
		http.DefaultServeMux.ServeHTTP(w, r)
		return
	}

	if !h.EnableProxy {
		/* proxy not enabled */
		w.WriteHeader(404)
		w.Write([]byte("<h1>Not Found</h1>"))
		h.LogReq(r, 404)
		return
	}

	/* proxy */

	if r.Method == "CONNECT" {
		h.HandleConnect(w, r)
		return
	}

	h.HandleHTTP(w, r)
}

func (h *Handler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		h.Log("connection not support hijack\n")
		w.WriteHeader(503)
		h.LogReq(r, 503)
		return
	}

	client_conn, _, err := hj.Hijack()
	if err != nil {
		h.Log("%s\n", err.Error())
		w.WriteHeader(503)
		h.LogReq(r, 503)
		return
	}

	srv := r.RequestURI
	if r.ProtoMajor == 2 {
		/* http/2.0 */
		srv = r.URL.Host
	}

	if strings.Index(srv, ":") == -1 {
		/* no port specialed, set port to 443 */
		srv = Sprintf("%s:443", srv)
	}

	server_conn, err := util.Dial("tcp", srv)
	if err != nil {
		h.Log("dial to server: %s\n", err.Error())
		w.WriteHeader(503)
		client_conn.Close()
		h.LogReq(r, 503)
		return
	}

	if r.ProtoMajor == 2 {
		w.WriteHeader(200)
	} else {
		client_conn.Write([]byte("HTTP/1.1 200 ok\r\n\r\n"))
	}

	h.LogReq(r, 200)

	go pipe(server_conn, client_conn)
	pipe(client_conn, server_conn)
}

func (h *Handler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var err error

	/* delete proxy-connection header */
	r.Header.Del("proxy-connection")

	/* set URL.Scheme, URL.Host for http/2.0 */
	if r.ProtoMajor == 2 {
		r.URL.Scheme = "http"
		r.URL.Host = r.Host
	}

	if h.Transport != nil {
		/* invoke user defined transport */
		resp, err = h.Transport.RoundTrip(r)
	} else {
		/* invoke default transport */
		resp, err = http.DefaultTransport.RoundTrip(r)
	}

	if err != nil {
		h.Log("proxy err: %s\n", err.Error())
		w.WriteHeader(503)
		h.LogReq(r, 503)
		return
	}

	for k, v := range resp.Header {
		w.Header().Set(k, strings.Join(v, ","))
	}

	w.WriteHeader(resp.StatusCode)

	h.LogReq(r, resp.StatusCode)

	io.Copy(w, resp.Body)

	resp.Body.Close()
}

func (h *Handler) is_local_request(r *http.Request) bool {
	/* http/1.x */
	if r.ProtoMajor == 1 {
		if r.RequestURI[0] == '/' {
			return true
		}
	}

	/* http/2.x */
	if r.ProtoMajor == 2 {
		host := r.Host
		if strings.Index(r.Host, ":") != -1 {
			host, _, _ = net.SplitHostPort(r.Host)
		}
		if h.LocalDomain != "" &&
			strings.HasSuffix(host, h.LocalDomain) {
			return true
		}
	}

	return false
}

/* copy and close */
func pipe(dst, src net.Conn) {
	io.Copy(dst, src)
	dst.Close()
}
