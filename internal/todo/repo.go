package todo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const table = "todos"

type Repository interface {
	List(ctx context.Context) ([]Todo, error)
	Get(ctx context.Context, id uint32) (Todo, error)
	Create(ctx context.Context, in TodoInput) (uint32, error)
	Update(ctx context.Context, id uint32, in TodoInput) error
	Delete(ctx context.Context, id uint32) error
}

type SQLRepo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *SQLRepo {
	return &SQLRepo{db: db}
}

func (r *SQLRepo) List(ctx context.Context) ([]Todo, error) {
	query := fmt.Sprintf("SELECT * FROM `%s`", table)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("List todos: %v", err)
	}
	defer rows.Close()

	todos := []Todo{}

	for rows.Next() {
		var t Todo

		if err := rows.Scan(&t.ID, &t.Todo, &t.Completed, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("Scan todo: %v", err)
		}

		todos = append(todos, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func (r *SQLRepo) Get(ctx context.Context, id uint32) (Todo, error) {
	query := fmt.Sprintf("SELECT id, todo, completed, created_at, updated_at FROM `%s` WHERE id=?", table)

	var t Todo

	err := r.db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.Todo, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Todo{}, fmt.Errorf("Row not found id=%d: %v", id, err)
	}
	if err != nil {
		return Todo{}, fmt.Errorf("Get todo id=%d: %v", id, err)
	}

	return t, nil
}

func (r *SQLRepo) Create(ctx context.Context, in TodoInput) (uint32, error) {
	if in.Todo == nil {
		return 0, fmt.Errorf("Empty todo")
	}

	query := fmt.Sprintf("INSERT INTO `%s` (todo) VALUES ('%s')", table, *in.Todo)

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("Error inserting row: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("Error reading affected rows: %v", err)
	}
	if rows != 1 {
		return 0, fmt.Errorf("Expected to affect 1 row, affected %d: %v", rows, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Error fetching ID of inserted row: %v", err)
	}

	return uint32(id), nil
}

func (r *SQLRepo) Update(ctx context.Context, id uint32, in TodoInput) error {
	if in.Todo == nil && in.Completed == nil {
		return fmt.Errorf("No changes provided in input")
	}

	query := fmt.Sprintf("UPDATE `%s` SET todo = IFNULL(?, todo), completed = IFNULL(?, completed) WHERE id=?", table)

	result, err := r.db.ExecContext(ctx, query, in.Todo, in.Completed, id)
	if err != nil {
		return fmt.Errorf("Error updating row: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Error reading affected rows: %v", err)
	}
	if rows != 1 {
		return fmt.Errorf("Expected to affect 1 row, affected %d: %v", rows, err)
	}

	return nil
}

func (r *SQLRepo) Delete(ctx context.Context, id uint32) error {
	query := fmt.Sprintf("DELETE FROM `%s` WHERE id=?", table)
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("Error deleting row: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Error reading affected rows: %v", err)
	}
	if rows != 1 {
		return fmt.Errorf("Expected to affect 1 row, affected %d: %v", rows, err)
	}

	return nil
}
