package main

import (
	"fmt"
	"net/http"
	"net/url"

	"os"
	//"path/filepath"
	//"strings"
)

func initRouters() {
	for _, r := range cfg.Routes {
		switch r.URLType {
		case "file":
			func(r URLRoute) {
				http.HandleFunc(r.URLPrefix, func(w http.ResponseWriter, req *http.Request) {
					http.ServeFile(w, req, r.Path)
				})
			}(r)
		case "dir":
			func(r URLRoute) {
				http.Handle(r.URLPrefix, http.FileServer(http.Dir(r.DocRoot)))
			}(r)
		case "uwsgi":
			func(r URLRoute) {
				_u1, err := url.Parse(r.Path)
				if err != nil {
					fmt.Printf("invalid path: %s\n", r.Path)
					os.Exit(-1)
				}
				_p := _u1.Path
				switch _u1.Scheme {
				case "unix":
				case "tcp":
					_p = _u1.Host
				default:
					fmt.Printf("invalid schemd: %s, only support unix, tcp", _u1.Scheme)
					os.Exit(-1)
				}
				_u := NewUwsgi(_u1.Scheme, _p, r.URLPrefix)
				http.Handle(r.URLPrefix, _u)
			}(r)
		case "fastcgi":
			func(r URLRoute) {
				_u1, err := url.Parse(r.Path)
				if err != nil {
					fmt.Printf("invalid path: %s\n", r.Path)
					os.Exit(-1)
				}
				_p := _u1.Path
				switch _u1.Scheme {
				case "unix":
				case "tcp":
					_p = _u1.Host
				default:
					fmt.Printf("invalid schemd: %s, only support unix, tcp", _u1.Scheme)
					os.Exit(-1)
				}
				_u, _ := NewFastCGI(_u1.Scheme, _p, r.DocRoot, r.URLPrefix)
				http.Handle(r.URLPrefix, _u)
			}(r)
		default:
			fmt.Printf("invalid type: %s\n", r.URLType)
		}
	}
}
