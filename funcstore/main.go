package main

import (
	"net/http"
	"serverless/config"
	"serverless/router"
	"serverless/util"
)

func main() {

	if config.ParseConfig() != nil {
		return
	}
	if util.Init() != nil {
		return
	}

	/*
		if util.InitLogger() != nil {
			return
		}
	*/
	handler := router.Load()

	http.ListenAndServe(":8888", handler)
}

