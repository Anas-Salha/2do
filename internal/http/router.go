package http

import (
	"context"
	"time"

	"github.com/anas-salha/2do/internal/todo"
	"github.com/gin-gonic/gin"
)

func ContextTimeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func NewRouter(todoHandler *todo.Handler) *gin.Engine {
	r := gin.Default()
	r.Use(ContextTimeout(2 * time.Second))

	api := r.Group("/api")
	v := api.Group("/v0")
	todoHandler.Register(v)

	r.GET("/healthz", func(c *gin.Context) { c.Status(204) })
	return r
}
