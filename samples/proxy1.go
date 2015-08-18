package main
import (
    "github.com/fangdingjun/myproxy"
    "net/http"
    "log"
    . "fmt"
)

func main(){
    port := 8080

    log.Print("Listen on: ", Sprintf("0.0.0.0:%d", port))
    err := http.ListenAndServe(Sprintf(":%d", port), &myproxy.Handler{})
    if err != nil{
        log.Fatal(err)
    }
}
