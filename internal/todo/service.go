package todo

import (
	"context"
	"errors"
	"fmt"
)

type Service interface {
	GetAll(ctx context.Context) ([]Todo, error)
	GetById(ctx context.Context, id uint32) (*Todo, error)
	Create(ctx context.Context, in TodoInput) (*Todo, error)
	Update(ctx context.Context, id uint32, in TodoInput) (*Todo, error)
	Delete(ctx context.Context, id uint32) error
}

type service struct {
	repo Repository
}

func NewService(r Repository) Service {
	return &service{repo: r}
}

func (s *service) GetAll(ctx context.Context) ([]Todo, error) {
	todos, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	return todos, nil
}

func (s *service) GetById(ctx context.Context, id uint32) (*Todo, error) {
	t, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("%w: id: %d", ErrNotFound, id)
		}
		return nil, err
	}

	return t, nil
}

func (s *service) Create(ctx context.Context, in TodoInput) (*Todo, error) {
	if in.Text == nil {
		return nil, fmt.Errorf("%w: text field missing", ErrInputInvalid)
	}

	t, err := s.repo.Create(ctx, in)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *service) Update(ctx context.Context, id uint32, in TodoInput) (*Todo, error) {
	if in.Text == nil && in.Completed == nil {
		return nil, fmt.Errorf("%w: text and completed fields missing - provide at least one.", ErrInputInvalid)
	}

	t, err := s.repo.Update(ctx, id, in)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("%w: id: %d", ErrNotFound, id)
		}
		return nil, err
	}

	return t, nil
}

func (s *service) Delete(ctx context.Context, id uint32) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("%w: id: %d", ErrNotFound, id)
		}
		return err
	}

	return nil
}
