package server

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/bobgromozeka/yp-diploma2/internal/server/grpc"
	"github.com/bobgromozeka/yp-diploma2/internal/server/storage"
	"github.com/bobgromozeka/yp-diploma2/pkg/helpers"
)

// Run starts server application.
// 1. Connects to database
// 2. Creates tables if needed
// 3. Creates storages
// 4. Runs gRPC server
func Run(addr string) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	db, dbErr := sql.Open("sqlite3", "yp-diploma2.db")
	if dbErr != nil {
		log.Fatalln("db open error: ", dbErr)
	}
	storage.Bootstrap(db)

	storagesFactory := storage.NewSQLiteStoragesFactory(db)

	uStorage := storagesFactory.CreateUserStorage()
	dkStorage := storagesFactory.CreateDataKeeperStorage()

	grpcServer := grpc.NewServer(
		uStorage, dkStorage, grpc.ServerConfig{
			Addr: addr, // TODO add address configuration from flags
		},
	)

	helpers.SetupGracefulShutdown(cancelFunc)

	serverStartErr := grpcServer.Start(ctx)
	if serverStartErr != nil {
		log.Fatalln("server start error: ", serverStartErr)
	}
}
