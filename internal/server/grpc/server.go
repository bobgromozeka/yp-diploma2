package grpc

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/user"
	"github.com/bobgromozeka/yp-diploma2/internal/server/grpc/services"
)

type ServerConfig struct {
	Addr string
}

type Server struct {
	db   *sql.DB
	conf ServerConfig
}

func NewServer(db *sql.DB, c ServerConfig) *Server {
	return &Server{
		db:   db,
		conf: c,
	}
}

func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.conf.Addr)
	if err != nil {
		return err
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	reflection.Register(grpcServer)

	user.RegisterUserServer(grpcServer, services.NewUserService(s.db))

	go func() {
		<-ctx.Done()
		fmt.Println("Stopping grpc server......")
		grpcServer.GracefulStop()
	}()

	fmt.Printf("Starting gRPC server on addr: [%s]......\n", s.conf.Addr)

	return grpcServer.Serve(lis)
}
