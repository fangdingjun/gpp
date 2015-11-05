/*
gpp is a http proxy handler, it can act as a proxy server and a http server.

Use it as a normal http.Handler. it determines the proxy request and the local request automatically.

It handle the proxy request itself, and route the local request to http.DefaultServerMux.

you can set its Handler options to yourself handler.

you can set EnableProxy to false to disable proxy function.

Example

a proxy example
    package main

    import (
        . "fmt"
        "github.com/fangdingjun/gpp"
        "log"
        "net/http"
    )

    func main() {
        port := 8080

        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(200)
            w.Write([]byte("<h1>welcome!</h1>"))
        })

        log.Print("Listen on: ", Sprintf("0.0.0.0:%d", port))
        err := http.ListenAndServe(Sprintf(":%d", port), &gpp.Handler{EnableProxy:true})
        if err != nil {
            log.Fatal(err)
        }
    }
Run above example and use curl to test it.

Run the follow command you will see a welcome message
    $ curl http://127.0.0.1:8080/
Run the follow command to test proxy function
    $ curl --proxy 127.0.0.1:8080 http://httpbin.org/ip

*/
package gpp

import (
	"github.com/fangdingjun/gpp/util"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

/*
This is proxy handler, you can use this as a http.Handler.
*/
type Handler struct {
	// the handler to process local path request
	Handler http.Handler

	// enable proxy or not
	EnableProxy bool

	// the local domain name, only required when http2 enabled
	LocalDomain string

	// the RoundTripper for http proxy
	Transport http.RoundTripper

	// log instance
	Logger *log.Logger

	// proxy require auth
	ProxyAuth bool

	/*
	   if ProxyAuth is true, ProxyAuthFunc used to check the user authorization,
	   return true if success, false if failed,

	   when failed, the response must be replyed to the client before the
	   function ProxyAuthFunc return
	*/
	ProxyAuthFunc func(w http.ResponseWriter, r *http.Request) bool
}

/*
Log the http request, if the h.Logger is nil, this function does nothing.

r is the client request

status is a http status code reply to client.

The log format look like this
    2015/09/09 15:21:41 59.44.39.234 "GET /proxy.pac HTTP/1.1" 200 "Mozilla/5.0 (compatible; MSIE 10.0; Win32; Trident/6.0)"
*/
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

		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		h.Logger.Printf("%s \"%s %s %s\" %03d \"%s\"\n",
			ip, r.Method, uri, r.Proto, status, ua,
		)
	}
}

/*
This is a shortcut for log.Printf, if h.Logger is nil this does nothing.
*/
func (h *Handler) Log(format string, args ...interface{}) {
	if h.Logger != nil {
		h.Logger.Printf(format, args...)
	}
}

/*
Impelemnt the http.Handler inferface.

It determimes the proxy request and the local page request automitically.

If the h.EnableProxy is false, all proxy requests will be denied.

If the h.Handler is nil, the local page request will be routed to http.DefaultServerMux.

If the h.Handler is not nil, will use h.Handler to handle the request.
*/
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

	if h.ProxyAuth {
		if h.ProxyAuthFunc == nil {
			panic("ProxyAuth is true but ProxyAuthFunc is nil")
		}

		if !h.ProxyAuthFunc(w, r) {
			return
		}
	}

	if r.Method == "CONNECT" {
		h.HandleConnect(w, r)
		return
	}

	h.HandleHTTP(w, r)
}

/*
handle the CONNECT request
*/
func (h *Handler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		h.Log("connection not support hijack\n")
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
		srv = net.JoinHostPort(srv, "443")
	}

	server_conn, err := util.Dial("tcp", srv)
	if err != nil {
		h.Log("dial to server: %s\n", err.Error())

		w.WriteHeader(503)

		w.Write([]byte(err.Error()))

		h.LogReq(r, 503)
		return
	}

	w.WriteHeader(200)
	h.LogReq(r, 200)

	client_conn, _, _ := hj.Hijack()

	go func() {
		io.Copy(server_conn, client_conn)
		server_conn.Close()
	}()

	io.Copy(client_conn, server_conn)
	client_conn.Close()
}

/*
Handle the other http proxy request, like GET, POST, HEAD.

If h.Transport is nil, will use http.DefaultTransport to process the request.

*/
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

	if r.Method != "POST" && r.Method != "PUT" {
		r.ContentLength = 0
		r.Body = nil
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

	hdr := w.Header()
	for k, v := range resp.Header {
		//h.Log("header: %s = %s\n", k, v)
		if strings.ToLower(k) != "connection" {
			for _, v1 := range v {
				hdr.Add(k, v1)
			}
		}
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
