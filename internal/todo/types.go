package todo

import "time"

type Todo struct {
	ID        uint32    `json:"id"`
	Todo      string    `json:"todo"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TodoInput struct {
	Todo      *string `json:"todo"`
	Completed *bool   `json:"completed"`
}
