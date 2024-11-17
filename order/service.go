package order

import (
	"context"
	"time"

	"github.com/segmentio/ksuid"
)

type Service interface {
	Post(ctx context.Context, accountId string, products []OrderedProduct) (*Order, error)
	GetByAccountId(ctx context.Context, accountId string) ([]Order, error)
}

type Order struct {
	ID         string
	CreatedAt  time.Time
	TotalPrice float64
	AccountId  string
	Products   []OrderedProduct
}

type OrderedProduct struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Quantity    int
}

type orderService struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &orderService{repository}
}

func (s *orderService) Post(ctx context.Context, accountId string, products []OrderedProduct) (*Order, error) {
	o := Order{
		ID:         ksuid.New().String(),
		CreatedAt:  time.Now().UTC(),
		TotalPrice: 0.0,
		AccountId:  accountId,
		Products:   products,
	}
	for _, p := range products {
		o.TotalPrice += p.Price * float64(p.Quantity)
	}

	err := s.repository.Put(ctx, o)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *orderService) GetByAccountId(ctx context.Context, accountId string) ([]Order, error) {
	return s.repository.GetByAccountId(ctx, accountId)
}
