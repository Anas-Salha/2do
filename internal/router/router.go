package router

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
)

func NewRouter(*sql.DB) *gin.Engine {
	router := gin.Default()
	router.GET("/todo", getAllTodos)
	router.GET("/todo/:id", getTodo)
	router.POST("/todo", postTodo)
	router.DELETE("/todo/:id", deleteTodo)
	router.PUT("/todo/:id", updateTodo)

	return router
}

func getAllTodos(c *gin.Context) {
	fmt.Println("getAllTodos called")
}

func getTodo(c *gin.Context) {
	fmt.Println("getTodo called")
}

func postTodo(c *gin.Context) {
	fmt.Println("postTodo called")
}

func deleteTodo(c *gin.Context) {
	fmt.Println("deleteTodo called")
}

func updateTodo(c *gin.Context) {
	fmt.Println("updateTodo called")
}
