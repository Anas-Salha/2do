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
	Get(ctx context.Context, id uint32) (*Todo, error)
	Create(ctx context.Context, in TodoInput) (*Todo, error)
	Update(ctx context.Context, id uint32, in TodoInput) (*Todo, error)
	Delete(ctx context.Context, id uint32) error
}

type sqlrepo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repository {
	return &sqlrepo{db: db}
}

func (r *sqlrepo) List(ctx context.Context) ([]Todo, error) {
	query := fmt.Sprintf("SELECT id, text, completed, created_at, updated_at FROM `%s`", table)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Text, &t.Completed, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func (r *sqlrepo) Get(ctx context.Context, id uint32) (*Todo, error) {
	query := fmt.Sprintf("SELECT id, text, completed, created_at, updated_at FROM `%s` WHERE id=?", table)

	var t Todo
	err := r.db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.Text, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &t, nil
}

func (r *sqlrepo) Create(ctx context.Context, in TodoInput) (*Todo, error) {
	query := fmt.Sprintf("INSERT INTO `%s` (text, completed) VALUES (?, ?)", table)

	result, err := r.db.ExecContext(ctx, query, in.Text, in.Completed)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	var t Todo
	getQuery := fmt.Sprintf("SELECT id, text, completed, created_at, updated_at FROM `%s` WHERE id=?", table)
	err = r.db.QueryRowContext(ctx, getQuery, id).Scan(&t.ID, &t.Text, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *sqlrepo) Update(ctx context.Context, id uint32, in TodoInput) (*Todo, error) {
	query := fmt.Sprintf("UPDATE `%s` SET text = IFNULL(?, text), completed = IFNULL(?, completed) WHERE id=?", table)

	result, err := r.db.ExecContext(ctx, query, in.Text, in.Completed, id)
	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows > 1 {
		return nil, ErrMultipleRowsAffected
	}

	var t Todo
	getQuery := fmt.Sprintf("SELECT id, text, completed, created_at, updated_at FROM `%s` WHERE id=?", table)
	err = r.db.QueryRowContext(ctx, getQuery, id).Scan(&t.ID, &t.Text, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &t, nil
}

func (r *sqlrepo) Delete(ctx context.Context, id uint32) error {
	query := fmt.Sprintf("DELETE FROM `%s` WHERE id=?", table)
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	} else if rows != 1 {
		return ErrMultipleRowsAffected
	}

	return nil
}
