package todo

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct{ svc Service }

func NewHandler(s Service) *Handler {
	return &Handler{svc: s}
}

func (h *Handler) Register(r gin.IRoutes) {
	r.GET("/todos", h.getAll)
	r.GET("/todos/:id", h.getById)
	r.POST("/todos", h.post)
	r.PUT("/todos/:id", h.put)
	r.DELETE("/todos/:id", h.delete)
}

func (h *Handler) getAll(ctx *gin.Context) {
	c := ctx.Request.Context()

	todos, err := h.svc.GetAll(c)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		return
	}

	ctx.JSON(http.StatusOK, todos)
}

func (h *Handler) getById(ctx *gin.Context) {
	i := ctx.Param("id")
	id, err := strconv.Atoi(i)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID must be an integer"})
		return
	}

	c := ctx.Request.Context()
	t, err := h.svc.GetById(c, uint32(id))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		return
	}

	ctx.JSON(http.StatusOK, t)
}

func (h *Handler) post(ctx *gin.Context) {
	if ctx.ContentType() != "application/json" {
		ctx.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Content-Type must be application/json"})
		return
	}

	var newTodo TodoInput
	err := ctx.ShouldBindJSON(&newTodo)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	if newTodo.Completed == nil {
		val := false
		newTodo.Completed = &val
	}

	c := ctx.Request.Context()
	t, err := h.svc.Create(c, newTodo)
	if err != nil {
		if errors.Is(err, ErrInputInvalid) {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		return
	}

	ctx.Header("Location", fmt.Sprintf("/todos/%d", t.ID))
	ctx.JSON(http.StatusCreated, t)
}

func (h *Handler) put(ctx *gin.Context) {
	if ctx.ContentType() != "application/json" {
		ctx.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Content-Type must be application/json"})
		return
	}

	i := ctx.Param("id")
	id, err := strconv.Atoi(i)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID must be an integer"})
		return
	}

	var updatedTodo TodoInput
	err = ctx.ShouldBindJSON(&updatedTodo)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	if updatedTodo.Text == nil || updatedTodo.Completed == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing field"})
		return
	}

	c := ctx.Request.Context()
	t, err := h.svc.Update(c, uint32(id), updatedTodo)
	if err != nil {
		if errors.Is(err, ErrInputInvalid) {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		return
	}

	ctx.JSON(http.StatusOK, t)
}

func (h *Handler) delete(ctx *gin.Context) {
	i := ctx.Param("id")
	id, err := strconv.Atoi(i)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID must be an integer"})
		return
	}

	c := ctx.Request.Context()
	err = h.svc.Delete(c, uint32(id))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
