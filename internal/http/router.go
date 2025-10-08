package http

import (
	"time"

	"github.com/gin-gonic/gin"
)

type registrable interface {
	Register(gin.IRoutes)
}

func NewRouter(todoHandler registrable, allowedOrigins []string) *gin.Engine {
	r := gin.Default()
	r.HandleMethodNotAllowed = true
	r.NoMethod(methodNotAllowed)
	r.Use(contextTimeout(2 * time.Second))
	r.Use(setCors(allowedOrigins))

	api := r.Group("/api")
	v := api.Group("/v0")
	todoHandler.Register(v)

	r.GET("/healthz", func(c *gin.Context) { c.Status(204) })
	return r
}
