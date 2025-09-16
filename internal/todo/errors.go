package todo

import "errors"

var ErrNotFound = errors.New("todo not found")
var ErrInputInvalid = errors.New("todo input invalid")
