package todo

import "time"

type Todo struct {
	ID        uint32    `json:"id"`
	Text      string    `json:"text"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TodoInput struct {
	Text      *string `json:"text"`
	Completed *bool   `json:"completed"`
}
