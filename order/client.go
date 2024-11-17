package order

import (
	"context"
	"log"
	"time"

	"github.com/lichb0rn/go-microservices/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.OrderServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:    conn,
		service: pb.NewOrderServiceClient(conn),
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) Post(ctx context.Context, accountId string, products []OrderedProduct) (*Order, error) {
	pbProducts := make([]*pb.PostOrderRequest_OrderProduct, 0, len(products))
	for _, p := range products {
		pbProducts = append(pbProducts, &pb.PostOrderRequest_OrderProduct{
			ProductId: p.ID,
			Quantity:  uint32(p.Quantity),
		})
	}
	r, err := c.service.PostOrder(ctx, &pb.PostOrderRequest{
		AccountId: accountId,
		Products:  pbProducts,
	})
	if err != nil {
		return nil, err
	}

	newOrder := r.Order
	createdAt := time.Time{}
	createdAt.UnmarshalBinary(newOrder.CreatedAt)

	return &Order{
		ID:         newOrder.Id,
		CreatedAt:  createdAt,
		TotalPrice: newOrder.TotalPrice,
		AccountId:  newOrder.AccountId,
		Products:   products,
	}, nil
}

func (c *Client) GetByAccountId(ctx context.Context, accountId string) ([]Order, error) {
	r, err := c.service.GetByAccountId(ctx, &pb.GetOrdersForAccountRequest{AccountId: accountId})
	if err != nil {
		log.Println("Orders not found: ", err)
		return nil, err
	}

	orders := make([]Order, 0, len(r.Orders))
	for _, o := range r.Orders {
		newOrder := Order{
			ID:         o.Id,
			TotalPrice: o.TotalPrice,
			AccountId:  o.AccountId,
		}
		newOrder.CreatedAt = time.Time{}
		newOrder.CreatedAt.UnmarshalBinary(o.CreatedAt)

		products := make([]OrderedProduct, 0, len(o.Products))
		for _, p := range o.Products {
			products = append(products, OrderedProduct{
				ID:          p.Id,
				Price:       p.Price,
				Name:        p.Name,
				Description: p.Description,
				Quantity:    int(p.Quantity),
			})
		}

		newOrder.Products = products
		orders = append(orders, newOrder)
	}

	return orders, nil
}
