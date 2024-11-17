//go:generate protoc --go_out=./ --go-grpc_out=./ account.proto
package account

import (
	"context"
	"fmt"
	"net"

	"github.com/lichb0rn/go-microservices/account/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedAccountServiceServer
	service Service
}

func ListendGRPC(s Service, port int) error {
	lis, err := net.Listen("tcp", ":"+fmt.Sprint(port))
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	pb.RegisterAccountServiceServer(server, &grpcServer{service: s})
	reflection.Register(server)
	return server.Serve(lis)
}

func (s *grpcServer) PostAccount(ctx context.Context, r *pb.PostAccountRequest) (*pb.PostAccountResponse, error) {
	acc, err := s.service.Post(ctx, r.Name)
	if err != nil {
		return nil, err
	}
	return &pb.PostAccountResponse{Account: &pb.Account{Id: acc.ID, Name: acc.Name}}, nil
}

func (s *grpcServer) GetAccount(ctx context.Context, r *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	acc, err := s.service.GetOne(ctx, r.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetAccountResponse{Account: &pb.Account{Id: acc.ID, Name: acc.Name}}, nil
}

func (s *grpcServer) GetAccounts(ctx context.Context, r *pb.GetAccountsRequest) (*pb.GetAccountsResponse, error) {
	accs, err := s.service.GetMany(ctx, r.Skip, r.Take)
	if err != nil {
		return nil, err
	}

	accounts := make([]*pb.Account, 0, len(accs))
	for _, acc := range accs {
		accounts = append(accounts, &pb.Account{Id: acc.ID, Name: acc.Name})
	}

	return &pb.GetAccountsResponse{Accounts: accounts}, nil
}
