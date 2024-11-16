package catalog

import (
	"context"

	"github.com/segmentio/ksuid"
)

type Service interface {
	Put(ctx context.Context, name, description string, price float64) (*Product, error)
	GetOne(ctx context.Context, id string) (*Product, error)
	GetMany(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	GetManyByIDs(ctx context.Context, ids []string) ([]Product, error)
	Search(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
}

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type catalogService struct {
	reposiotry Repository
}

func NewService(r Repository) Service {
	return &catalogService{r}
}

func (s *catalogService) Put(ctx context.Context, name, description string, price float64) (*Product, error) {
	p := &Product{
		Name:        name,
		Description: description,
		Price:       price,
		ID:          ksuid.New().String(),
	}

	if err := s.reposiotry.Put(ctx, *p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *catalogService) GetOne(ctx context.Context, id string) (*Product, error) {
	return s.reposiotry.GetById(ctx, id)
}

func (s *catalogService) GetMany(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}

	return s.reposiotry.List(ctx, skip, take)
}

func (s *catalogService) GetManyByIDs(ctx context.Context, ids []string) ([]Product, error) {
	return s.reposiotry.ListWithIDs(ctx, ids)
}

func (s *catalogService) Search(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}

	return s.reposiotry.Search(ctx, query, skip, take)
}
