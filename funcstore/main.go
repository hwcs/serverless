package main

import (
    "net/http"
    "serverless/router"
)

func main() {
    handler := router.Load ()

    http.ListenAndServe (
        ":8888", handler )
}
