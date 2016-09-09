package router

import (
	"net/http"
	"serverless/server"
        "github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
)

func Load() http.Handler {
	e := gin.New() // returns a new blank Engine instance without any middleware attached

	// Use attachs a global middleware to the router. ie. the middleware attached though Use() will be
	// included in the handlers chain for every single request. Even 404, 405, static files...
	// For example, this is the right place for a logger or error management middleware.
	e.Use(gin.Recovery()) // Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.

	// Group creates a new router group. You should add all the routes that have common middlwares or the same path prefix.
	// For example, all the routes that use a common middlware for authorization could be grouped.
	functions := e.Group("/functions")
	{
		functions.POST("/", server.CreateFunction)
		functions.GET("/:name", server.GetFunction)
		functions.GET("/", server.GetFunctionList)
		functions.GET("/:name/code", server.GetFunctionCode)
		functions.PUT("/:name/config", server.UpdateFunctionConfig)
		functions.PUT("/:name/code", server.UpdateFunctionCode)
		functions.DELETE("/:name", server.DeleteFunction)

		functions.POST("/:name/invocations", server.InvocateFunction)
	}

        ginpprof.Wrapper(e)
	return e
}

