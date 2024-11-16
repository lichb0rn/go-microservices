//go:generate protoc --go_out=./ --go-grpc_out=./ catalog.proto
package catalog

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/lichb0rn/go-microservices/catalog/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedCatalogServiceServer
	service Service
}

func ListendGRPC(s Service, port int) error {
	lis, err := net.Listen("tcp", ":"+fmt.Sprint(port))
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	pb.RegisterCatalogServiceServer(server, &grpcServer{service: s})
	reflection.Register(server)
	return server.Serve(lis)
}

func (s *grpcServer) PostProduct(ctx context.Context, r *pb.PostProductRequest) (*pb.PostProductResponse, error) {
	p, err := s.service.Put(ctx, r.Name, r.Description, r.Price)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &pb.PostProductResponse{Product: &pb.Product{Id: p.ID, Name: p.Name, Description: p.Description, Price: p.Price}}, nil
}

func (s *grpcServer) GetProduct(ctx context.Context, r *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	p, err := s.service.GetOne(ctx, r.Id)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &pb.GetProductResponse{Product: &pb.Product{Id: p.ID, Name: p.Name, Description: p.Description, Price: p.Price}}, nil
}

func (s *grpcServer) GetProducts(ctx context.Context, r *pb.GetProductsRequest) (*pb.GetProductsResponse, error) {
	var res []Product
	var err error
	if r.Query != "" {
		res, err = s.service.Search(ctx, r.Query, r.Skip, r.Take)
	} else if len(r.Ids) > 0 {
		res, err = s.service.GetManyByIDs(ctx, r.Ids)
	} else {
		res, err = s.service.GetMany(ctx, r.Skip, r.Take)
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}

	products := make([]*pb.Product, 0, len(res))
	for _, p := range res {
		products = append(products, &pb.Product{Id: p.ID, Name: p.Name, Description: p.Description, Price: p.Price})
	}
	return &pb.GetProductsResponse{Products: products}, err
}
