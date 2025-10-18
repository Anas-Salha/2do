package todo

import (
	"errors"
	"time"
)

var (
	ErrTodoNotFound = errors.New("todo_not_found")
	ErrInputInvalid = errors.New("input_invalid") //Placeholder - replace with more meaningful errors. e.g. todo_too_long
	ErrUnexpected   = errors.New("unexpected")
)

var (
	ErrBadJson              = errors.New("bad_json")
	ErrBadId                = errors.New("bad_id")
	ErrUnsupportedMediaType = errors.New("unsupported_media_type")
)

type ErrorResponse struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message,omitempty"`
		Timestamp string `json:"timestamp"` //RFC3339
	} `json:"error"`
}

func NewErrorResponse(code, msg string) *ErrorResponse {
	e := &ErrorResponse{}
	e.Error.Code = code
	e.Error.Timestamp = time.Now().UTC().Format(time.RFC3339)

	if msg != "" {
		e.Error.Message = msg
	}
	return e
}
