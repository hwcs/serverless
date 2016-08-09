package router

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "serverless/server"
)

func Load () http.Handler {
    e := gin.New ()
    e.Use (gin.Recovery ())

    functions := e.Group ("/functions")
    {
        functions.POST("", server.CreateFunction)
    }
    return e
}
