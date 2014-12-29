package main

import (
    "net/http"
    "./fileservice"
    "./responsehandler"
)

func main() {
    http.HandleFunc("/", responsehandler.Handler)
    go func() {
      http.ListenAndServe(":8888", nil)
    }()
    fileservice.NewWatcher()
}
