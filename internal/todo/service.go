package todo

import (
	"context"
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
		return nil, err
	}

	return t, nil
}

func (s *service) Create(ctx context.Context, in TodoInput) (*Todo, error) {
	t, err := s.repo.Create(ctx, in)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *service) Update(ctx context.Context, id uint32, in TodoInput) (*Todo, error) {
	t, err := s.repo.Update(ctx, id, in)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *service) Delete(ctx context.Context, id uint32) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
