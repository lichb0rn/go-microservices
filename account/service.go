package account

import (
	"context"

	"github.com/segmentio/ksuid"
)

type Service interface {
	Post(ctx context.Context, name string) (*Account, error)
	GetOne(ctx context.Context, id string) (*Account, error)
	GetMany(ctx context.Context, skip uint64, take uint64) ([]Account, error)
}

type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type accountService struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &accountService{repository: r}
}

func (s *accountService) Post(ctx context.Context, name string) (*Account, error) {
	a := &Account{
		Name: name,
		ID:   ksuid.New().String(),
	}

	if err := s.repository.Put(ctx, *a); err != nil {
		return nil, err
	}

	return a, nil
}

func (s *accountService) GetOne(ctx context.Context, id string) (*Account, error) {
	return s.repository.GetById(ctx, id)
}

func (s *accountService) GetMany(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}

	return s.repository.List(ctx, skip, take)
}
