package main

import (
	"net/http"
	"serverless/router"
	"serverless/util"
)

func main() {

	if util.Init() != nil {
		return
	}

	handler := router.Load()

	http.ListenAndServe(":8888", handler)
}

