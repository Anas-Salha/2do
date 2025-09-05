package router

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const todosTable = "todos"

func ContextTimeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func NewRouter(db *sql.DB) *gin.Engine {
	router := gin.Default()
	router.Use(ContextTimeout(2 * time.Second))
	router.GET("/todo", func(ctx *gin.Context) {
		getAllTodos(ctx, db)
	})
	router.GET("/todo/:id", func(ctx *gin.Context) {
		getTodo(ctx, db)
	})
	router.POST("/todo", func(ctx *gin.Context) {
		postTodo(ctx, db)
	})
	router.PUT("/todo/:id", func(ctx *gin.Context) {
		updateTodo(ctx, db)
	})
	router.DELETE("/todo/:id", func(ctx *gin.Context) {
		deleteTodo(ctx, db)
	})

	return router
}

func getAllTodos(ctx *gin.Context, db *sql.DB) {
	query := fmt.Sprintf("SELECT * FROM `%s`", todosTable)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	todos := []Todo{}

	for rows.Next() {
		var t Todo

		if err := rows.Scan(&t.ID, &t.Todo, &t.Completed, &t.CreatedAt, &t.UpdatedAt); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		todos = append(todos, t)
	}

	if err := rows.Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, todos)
}

func getTodo(ctx *gin.Context, db *sql.DB) {
	id := ctx.Param("id")
	query := fmt.Sprintf("SELECT id, todo, completed, created_at, updated_at FROM `%s` WHERE id=?", todosTable)

	var t Todo

	err := db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.Todo, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	switch {
	case err == sql.ErrNoRows:
		ctx.JSON(http.StatusNotFound, gin.H{"error": "requested row not found in database"})
	case err != nil:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	default:
		ctx.JSON(http.StatusOK, t)
	}
}

func postTodo(ctx *gin.Context, db *sql.DB) {
	var newTodo TodoInserter
	if err := ctx.ShouldBindJSON(&newTodo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if newTodo.Todo == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "provide todo"})
		return
	}

	query := fmt.Sprintf("INSERT INTO `%s` (todo) VALUES ('%s')", todosTable, *newTodo.Todo)

	result, err := db.ExecContext(ctx, query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rows != 1 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("expected to affect 1 row, affected %d", rows)})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch updated row
	var readTodo Todo
	readQuery := fmt.Sprintf("SELECT id, todo, completed, created_at, updated_at FROM `%s` WHERE id=?", todosTable)

	if err = db.QueryRowContext(ctx, readQuery, id).Scan(&readTodo.ID, &readTodo.Todo, &readTodo.Completed, &readTodo.CreatedAt, &readTodo.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "inserted row not found in database"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond
	ctx.Header("Location", fmt.Sprintf("/todo/%d", id))
	ctx.JSON(http.StatusCreated, readTodo)
}

func updateTodo(ctx *gin.Context, db *sql.DB) {
	var updatedTodo TodoInserter
	if err := ctx.ShouldBindJSON(&updatedTodo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updatedTodo.Todo == nil && updatedTodo.Completed == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "provide at least one of: todo, completed"})
		return
	}

	query := fmt.Sprintf("UPDATE `%s` SET todo = IFNULL(?, todo), completed = IFNULL(?, completed) WHERE id=?", todosTable)

	id := ctx.Param("id")
	_, err := db.ExecContext(ctx, query, updatedTodo.Todo, updatedTodo.Completed, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch updated row
	var readTodo Todo
	readQuery := fmt.Sprintf("SELECT id, todo, completed, created_at, updated_at FROM `%s` WHERE id=?", todosTable)

	if err = db.QueryRowContext(ctx, readQuery, id).Scan(&readTodo.ID, &readTodo.Todo, &readTodo.Completed, &readTodo.CreatedAt, &readTodo.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "inserted row not found in database"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond
	ctx.JSON(http.StatusOK, readTodo)
}

func deleteTodo(ctx *gin.Context, db *sql.DB) {
	id := ctx.Param("id")
	query := fmt.Sprintf("DELETE FROM `%s` WHERE id=?", todosTable)

	result, err := db.ExecContext(ctx, query, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rows != 1 {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("expected to affect 1 row, affected %d", rows)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted": true})
}
