package server

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/bobgromozeka/yp-diploma2/internal/server/grpc"
	"github.com/bobgromozeka/yp-diploma2/pkg/helpers"
)

func Run() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	db, dbErr := sql.Open("sqlite3", "yp-diploma2.db")
	if dbErr != nil {
		log.Fatalln("db open error: ", dbErr)
	}
	bootstrap(db)

	grpcServer := grpc.NewServer(
		db, grpc.ServerConfig{
			Addr: ":14444", // TODO add address configuration from flags
		},
	)

	helpers.SetupGracefulShutdown(cancelFunc)

	serverStartErr := grpcServer.Start(ctx)
	if serverStartErr != nil {
		log.Fatalln("server start error: ", serverStartErr)
	}
}
