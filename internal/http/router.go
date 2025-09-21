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
	r.Use(ContextTimeout(2 * time.Second))
	r.Use(Cors(allowedOrigins))

	api := r.Group("/api")
	v := api.Group("/v0")
	todoHandler.Register(v)

	r.GET("/healthz", func(c *gin.Context) { c.Status(204) })
	return r
}
