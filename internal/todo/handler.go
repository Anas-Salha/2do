package todo

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct{ repo Repository }

func NewHandler(r Repository) *Handler {
	return &Handler{repo: r}
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

	todos, err := h.repo.List(c)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, todos)
}

func (h *Handler) getById(ctx *gin.Context) {
	i := ctx.Param("id")
	id, err := strconv.Atoi(i)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c := ctx.Request.Context()
	todo, err := h.repo.Get(c, uint32(id))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, todo)
}

func (h *Handler) post(ctx *gin.Context) {
	var newTodo TodoInput
	err := ctx.ShouldBindJSON(&newTodo)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c := ctx.Request.Context()
	id, err := h.repo.Create(c, newTodo)
	if err != nil {
		if errors.Is(err, ErrMissingFields) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Header("Location", fmt.Sprintf("/todo/%d", id))
	ctx.JSON(http.StatusCreated, gin.H{"created": true})
}

func (h *Handler) put(ctx *gin.Context) {
	var updatedTodo TodoInput
	err := ctx.ShouldBindJSON(&updatedTodo)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	i := ctx.Param("id")
	id, err := strconv.Atoi(i)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c := ctx.Request.Context()
	err = h.repo.Update(c, uint32(id), updatedTodo)
	if err != nil {
		if errors.Is(err, ErrMissingFields) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"updated": true})
}

func (h *Handler) delete(ctx *gin.Context) {
	i := ctx.Param("id")
	id, err := strconv.Atoi(i)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c := ctx.Request.Context()
	err = h.repo.Delete(c, uint32(id))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
