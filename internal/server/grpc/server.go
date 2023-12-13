package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/datakeeper"
	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/user"
	"github.com/bobgromozeka/yp-diploma2/internal/server/grpc/interceptors"
	"github.com/bobgromozeka/yp-diploma2/internal/server/grpc/services"
	"github.com/bobgromozeka/yp-diploma2/internal/server/storage"
)

// ServerConfig server configuration params
type ServerConfig struct {
	Addr string
}

// Server struct with app dependencies
type Server struct {
	uStorage  storage.UserStorage
	dkStorage storage.DataKeeperStorage
	conf      ServerConfig
}

// NewServer returns pointer to new server struct
func NewServer(us storage.UserStorage, dks storage.DataKeeperStorage, c ServerConfig) *Server {
	return &Server{
		uStorage:  us,
		dkStorage: dks,
		conf:      c,
	}
}

// Start starts grpc server with all needed interceptors and options from config
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.conf.Addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.AuthnUnary),
		grpc.StreamInterceptor(interceptors.AuthnStream),
	)

	reflection.Register(grpcServer)

	user.RegisterUserServer(grpcServer, services.NewUserService(s.uStorage))
	datakeeper.RegisterDataKeeperServer(grpcServer, services.NewDataKeeperService(s.dkStorage))

	go func() {
		<-ctx.Done()
		fmt.Println("Stopping grpc server......")
		grpcServer.GracefulStop()
	}()

	fmt.Printf("Starting gRPC server on addr: [%s]......\n", s.conf.Addr)

	return grpcServer.Serve(lis)
}
