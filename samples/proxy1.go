package main

import (
	. "fmt"
	"github.com/fangdingjun/myproxy"
	"log"
	"net/http"
)

func main() {
	port := 8080

	log.Print("Listen on: ", Sprintf("0.0.0.0:%d", port))
	err := http.ListenAndServe(Sprintf(":%d", port), &myproxy.Handler{})
	if err != nil {
		log.Fatal(err)
	}
}
