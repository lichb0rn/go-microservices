//go:generate protoc --go_out=./ --go-grpc_out=./ order.proto
package order

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/lichb0rn/go-microservices/account"
	"github.com/lichb0rn/go-microservices/catalog"
	"github.com/lichb0rn/go-microservices/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedOrderServiceServer
	service       Service
	accountClient *account.Client
	catalogClient *catalog.Client
}

func ListendGRPC(s Service, accountUrl, catalogUrl string, port int) error {
	accountClient, err := account.NewClient(accountUrl)
	if err != nil {
		return err
	}
	catalogClient, err := catalog.NewClient(catalogUrl)
	if err != nil {
		accountClient.Close()
		return err
	}

	lis, err := net.Listen("tcp", ":"+fmt.Sprint(port))
	if err != nil {
		accountClient.Close()
		catalogClient.Close()
		return err
	}
	server := grpc.NewServer()
	pb.RegisterOrderServiceServer(server, &grpcServer{
		service:       s,
		accountClient: accountClient,
		catalogClient: catalogClient,
	})
	reflection.Register(server)
	return server.Serve(lis)
}

func (s *grpcServer) PostOrder(ctx context.Context, r *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {
	_, err := s.accountClient.GetOne(ctx, r.AccountId)
	if err != nil {
		log.Panicln("Account not found: ", err)
		return nil, errors.New("account not found")
	}

	productIds := make([]string, 0, len(r.Products))
	orderedProducts, err := s.catalogClient.GetProducts(ctx, 0, 0, productIds, "")
	if err != nil {
		log.Panicln("Products not found: ", err)
		return nil, errors.New("products not found")
	}

	products := make([]OrderedProduct, 0, len(orderedProducts))
	for _, p := range orderedProducts {
		product := OrderedProduct{
			ID:          p.ID,
			Price:       p.Price,
			Name:        p.Name,
			Description: p.Description,
			Quantity:    0,
		}

		for _, rp := range r.Products {
			if rp.ProductId == p.ID {
				product.Quantity = int(rp.Quantity)
				break
			}
		}

		if product.Quantity > 0 {
			products = append(products, product)
		}
	}

	order, err := s.service.Post(ctx, r.AccountId, products)
	if err != nil {
		log.Panicln("Order not created: ", err)
		return nil, errors.New("order not created")
	}

	orderProto := &pb.Order{
		Id:         order.ID,
		AccountId:  order.AccountId,
		TotalPrice: order.TotalPrice,
		Products:   []*pb.Order_OrderProduct{},
	}
	orderProto.CreatedAt, _ = order.CreatedAt.MarshalBinary()

	for _, p := range order.Products {
		orderProto.Products = append(orderProto.Products, &pb.Order_OrderProduct{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    uint32(p.Quantity),
		})
	}

	return &pb.PostOrderResponse{Order: orderProto}, nil

}

func (s *grpcServer) GetOrdersForAccount(ctx context.Context, r *pb.GetOrdersForAccountRequest) (*pb.GetOrdersForAccountResponse, error) {
	accountOrders, err := s.service.GetByAccountId(ctx, r.AccountId)
	if err != nil {
		log.Println("Orders not found: ", err)
		return nil, errors.New("orders not found")
	}

	productIdMap := map[string]bool{}
	for _, o := range accountOrders {
		for _, p := range o.Products {
			productIdMap[p.ID] = true
		}
	}

	productIds := make([]string, 0, len(productIdMap))
	for id := range productIdMap {
		productIds = append(productIds, id)
	}

	products, err := s.catalogClient.GetProducts(ctx, 0, 0, productIds, "")
	if err != nil {
		log.Println("Products not found: ", err)
		return nil, errors.New("products not found")
	}

	orders := make([]*pb.Order, 0, len(accountOrders))
	for _, o := range accountOrders {
		op := &pb.Order{
			Id:         o.ID,
			AccountId:  o.AccountId,
			TotalPrice: o.TotalPrice,
			Products:   []*pb.Order_OrderProduct{},
		}
		op.CreatedAt, _ = o.CreatedAt.MarshalBinary()

		for _, product := range o.Products {
			for _, p := range products {
				if product.ID == p.ID {
					product.Name = p.Name
					product.Description = p.Description
					product.Price = p.Price
					break
				}
			}
			op.Products = append(op.Products, &pb.Order_OrderProduct{
				Id:          product.ID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				Quantity:    uint32(product.Quantity),
			})
		}

		orders = append(orders, op)
	}

	return &pb.GetOrdersForAccountResponse{Orders: orders}, nil
}
