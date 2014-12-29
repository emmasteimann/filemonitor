package main

import (
    "net/http"
    "./fileservice"
    "./responsehandler"
)

const logfile string = "/tmp/foo/mylog.log"

func main() {
    http.HandleFunc("/", responsehandler.Handler)
    go func() {
      http.ListenAndServe(":8888", nil)
    }()
    fileservice.NewWatcher()
}
