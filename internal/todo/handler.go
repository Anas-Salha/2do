package todo

import (
	"bytes"
	"encoding/json"
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
	r.PATCH("/todos/:id", h.patch)
	r.DELETE("/todos/:id", h.delete)
}

func (h *Handler) getAll(ctx *gin.Context) {
	c := ctx.Request.Context()

	todos, err := h.svc.GetAll(c)
	if err != nil {
		r := NewErrorResponse(ErrUnexpected.Error(), "")
		ctx.JSON(http.StatusInternalServerError, r)
		return
	}

	ctx.JSON(http.StatusOK, todos)
}

func (h *Handler) getById(ctx *gin.Context) {
	i := ctx.Param("id")
	id, err := strconv.ParseUint(i, 10, 32)
	if err != nil {
		r := NewErrorResponse(ErrBadId.Error(), "ID must be an integer")
		ctx.JSON(http.StatusBadRequest, r)
		return
	}

	c := ctx.Request.Context()
	t, err := h.svc.GetById(c, uint32(id))
	if err != nil {
		if errors.Is(err, ErrTodoNotFound) {
			msg := fmt.Sprintf("No resource found with ID = %d", id)
			r := NewErrorResponse(ErrTodoNotFound.Error(), msg)
			ctx.JSON(http.StatusNotFound, r)
			return
		}
		r := NewErrorResponse(ErrUnexpected.Error(), "")
		ctx.JSON(http.StatusInternalServerError, r)
		return
	}

	ctx.JSON(http.StatusOK, t)
}

func (h *Handler) post(ctx *gin.Context) {
	if ctx.ContentType() != "application/json" {
		r := NewErrorResponse(ErrUnsupportedMediaType.Error(), "Content-Type must be application/json")
		ctx.JSON(http.StatusUnsupportedMediaType, r)
		return
	}

	var newTodo TodoInput
	err := decodeIntoInput(ctx, &newTodo)
	if err != nil {
		r := NewErrorResponse(ErrBadJson.Error(), err.Error())
		ctx.JSON(http.StatusBadRequest, r)
		return
	}

	if newTodo.Text == nil {
		r := NewErrorResponse(ErrBadJson.Error(), "missing required `text` field")
		ctx.JSON(http.StatusBadRequest, r)
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
			r := NewErrorResponse(ErrInputInvalid.Error(), err.Error())
			ctx.JSON(http.StatusUnprocessableEntity, r)
			return
		}
		r := NewErrorResponse(ErrUnexpected.Error(), "")
		ctx.JSON(http.StatusInternalServerError, r)
		return
	}

	ctx.Header("Location", fmt.Sprintf("/todos/%d", t.ID))
	ctx.JSON(http.StatusCreated, t)
}

func (h *Handler) put(ctx *gin.Context) {
	if ctx.ContentType() != "application/json" {
		r := NewErrorResponse(ErrUnsupportedMediaType.Error(), "Content-Type must be application/json")
		ctx.JSON(http.StatusUnsupportedMediaType, r)
		return
	}

	i := ctx.Param("id")
	id, err := strconv.ParseUint(i, 10, 32)
	if err != nil {
		r := NewErrorResponse(ErrBadId.Error(), "ID must be an integer")
		ctx.JSON(http.StatusBadRequest, r)
		return
	}

	var updatedTodo TodoInput
	err = decodeIntoInput(ctx, &updatedTodo)
	if err != nil {
		r := NewErrorResponse(ErrBadJson.Error(), err.Error())
		ctx.JSON(http.StatusBadRequest, r)
		return
	}

	if updatedTodo.Text == nil || updatedTodo.Completed == nil {
		r := NewErrorResponse(ErrBadJson.Error(), "missing required `text` and `completed` fields")
		ctx.JSON(http.StatusBadRequest, r)
		return
	}

	c := ctx.Request.Context()
	t, err := h.svc.Update(c, uint32(id), updatedTodo)
	if err != nil {
		if errors.Is(err, ErrInputInvalid) {
			r := NewErrorResponse(ErrInputInvalid.Error(), err.Error())
			ctx.JSON(http.StatusUnprocessableEntity, r)
			return
		}
		if errors.Is(err, ErrTodoNotFound) {
			msg := fmt.Sprintf("No resource found with ID = %d", id)
			r := NewErrorResponse(ErrTodoNotFound.Error(), msg)
			ctx.JSON(http.StatusNotFound, r)
			return
		}
		r := NewErrorResponse(ErrUnexpected.Error(), "")
		ctx.JSON(http.StatusInternalServerError, r)
		return
	}

	ctx.JSON(http.StatusOK, t)
}

func (h *Handler) patch(ctx *gin.Context) {
	if ctx.ContentType() != "application/json" {
		r := NewErrorResponse(ErrUnsupportedMediaType.Error(), "Content-Type must be application/json")
		ctx.JSON(http.StatusUnsupportedMediaType, r)
		return
	}

	i := ctx.Param("id")
	id, err := strconv.ParseUint(i, 10, 32)
	if err != nil {
		r := NewErrorResponse(ErrBadId.Error(), "ID must be an integer")
		ctx.JSON(http.StatusBadRequest, r)
		return
	}

	var updatedTodo TodoInput
	err = decodeIntoInput(ctx, &updatedTodo)
	if err != nil {
		r := NewErrorResponse(ErrBadJson.Error(), err.Error())
		ctx.JSON(http.StatusBadRequest, r)
		return
	}

	if updatedTodo.Text == nil && updatedTodo.Completed == nil {
		msg := "missing required `text` or `completed` field"
		r := NewErrorResponse(ErrBadJson.Error(), msg)
		ctx.JSON(http.StatusBadRequest, r)
		return
	}
	c := ctx.Request.Context()
	t, err := h.svc.Update(c, uint32(id), updatedTodo)
	if err != nil {
		if errors.Is(err, ErrInputInvalid) {
			r := NewErrorResponse(ErrInputInvalid.Error(), err.Error())
			ctx.JSON(http.StatusUnprocessableEntity, r)
			return
		}
		if errors.Is(err, ErrTodoNotFound) {
			msg := fmt.Sprintf("No resource found with ID = %d", id)
			r := NewErrorResponse(ErrTodoNotFound.Error(), msg)
			ctx.JSON(http.StatusNotFound, r)
			return
		}
		r := NewErrorResponse(ErrUnexpected.Error(), "")
		ctx.JSON(http.StatusInternalServerError, r)
		return
	}

	ctx.JSON(http.StatusOK, t)
}

func (h *Handler) delete(ctx *gin.Context) {
	i := ctx.Param("id")
	id, err := strconv.ParseUint(i, 10, 32)
	if err != nil {
		r := NewErrorResponse(ErrBadId.Error(), "ID must be an integer")
		ctx.JSON(http.StatusBadRequest, r)
		return
	}

	c := ctx.Request.Context()
	err = h.svc.Delete(c, uint32(id))
	if err != nil {
		if errors.Is(err, ErrTodoNotFound) {
			msg := fmt.Sprintf("No resource found with ID = %d", id)
			r := NewErrorResponse(ErrTodoNotFound.Error(), msg)
			ctx.JSON(http.StatusNotFound, r)
			return
		}
		r := NewErrorResponse(ErrUnexpected.Error(), "")
		ctx.JSON(http.StatusInternalServerError, r)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// decodeIntoInput decodes the JSON request body from the provided gin.Context
// into the given TodoInput struct. It returns an error if the input is not valid
// JSON, contains unknown fields, contains fields explicitly set to null, or if
// there is extra data after the first JSON object. The function has the side
// effect of populating the provided TodoInput argument with the decoded data.
func decodeIntoInput(ctx *gin.Context, t *TodoInput) error {
	// Decode into a map to check for explicit nulls
	var raw map[string]*json.RawMessage
	if err := ctx.ShouldBindJSON(&raw); err != nil {
		return errors.New("invalid json input")
	}

	// Check for explicit nulls
	for k, v := range raw {
		if v == nil {
			return fmt.Errorf("field %q cannot be null", k)
		}
		var tmp any
		if err := json.Unmarshal(*v, &tmp); err == nil && tmp == nil {
			return fmt.Errorf("field %q cannot be null", k)
		}
	}

	// Re-marshal the map back to JSON bytes
	buf, err := json.Marshal(raw)
	if err != nil {
		return errors.New("invalid json input")
	}

	// Use decoder to decode into the struct, protecting against unknown fields
	decoder := json.NewDecoder(bytes.NewReader(buf))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(t); err != nil {
		return errors.New("invalid json input")
	}

	if decoder.More() {
		return errors.New("invalid json input")
	}

	return nil
}
