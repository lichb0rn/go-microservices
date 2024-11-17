package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/lichb0rn/go-microservices/order"
)

var (
	ErrInvalidaParameter = errors.New("invalid parameter")
)

type mutationResolver struct {
	server *Server
}

func (r *mutationResolver) CreateAccount(ctx context.Context, in AccountInput) (*Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	acc, err := r.server.accountClient.Post(ctx, in.Name)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &Account{ID: acc.ID, Name: acc.Name}, nil
}

func (r *mutationResolver) CreateProduct(ctx context.Context, in ProductInput) (*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	product, err := r.server.catalogClient.PostProduct(ctx, in.Name, in.Description, in.Price)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &Product{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
	}, nil
}

func (r *mutationResolver) CreateOrder(ctx context.Context, in OrderInput) (*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var products []order.OrderedProduct
	for _, p := range in.Products {
		if p.Quantity <= 0 {
			return nil, ErrInvalidaParameter
		}

		products = append(products, order.OrderedProduct{
			ID:       p.ID,
			Quantity: p.Quantity,
		})
	}

	order, err := r.server.orderClient.Post(ctx, in.AccountID, products)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &Order{
		ID:         order.ID,
		CreatedAt:  order.CreatedAt,
		TotalPrice: order.TotalPrice,
	}, nil

}
