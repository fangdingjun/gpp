package main

import (
	"net/http"
)

func init_routers() {
	proxy := NewProxy("127.0.0.1:9090")
	Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir(docroot))))

	Router.HandleFunc("/proxy.pac",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/javascript")
			w.WriteHeader(200)
			w.Write([]byte(`function FindProxyForUrl(url, host){
    return "PRXOY 127.0.0.1:8080";
}
`))
		})

	Router.HandleFunc("/add_route.bat",
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "/home/dingjun/html/add_route.bat")
		})

	Router.HandleFunc("/del_route.bat",
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "/home/dingjun/html/del_route.bat")
		})

	Router.PathPrefix("/").HandlerFunc(proxy.ProxyPass)

	/* defaut router */
	http.HandleFunc("/", proxy.ProxyPass)
	http.Handle("/static",
		http.StripPrefix("/static", http.FileServer(http.Dir(docroot))))

}
