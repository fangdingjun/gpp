package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var proxy = NewProxy("127.0.0.1:9090")

func roothandler(w http.ResponseWriter, r *http.Request) {
	fullpath := filepath.Join(docroot, r.URL.Path)

	/* local file not exists */
	if _, err := os.Stat(fullpath); err != nil {
		proxy.ProxyPass(w, r)
		return
	}

	/* file out of docroot, path may contains .. */
	if b := strings.HasPrefix(filepath.Clean(fullpath),
		filepath.Clean(docroot)); !b {
		w.WriteHeader(404)
		w.Write([]byte("<h1>Not Found</h1>"))
		return
	}

	/* serve local file */
	http.ServeFile(w, r, fullpath)
}

func init_routers() {

	Router.PathPrefix("/").HandlerFunc(roothandler)

	/* defaut router */
	http.HandleFunc("/", roothandler)
}
