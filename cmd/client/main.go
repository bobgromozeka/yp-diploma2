package main

import (
	"context"
	"log"

	"github.com/bobgromozeka/yp-diploma2/internal/client"
	"github.com/bobgromozeka/yp-diploma2/pkg/helpers"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	helpers.SetupGracefulShutdown(cancel)

	if clientErr := client.Run(ctx); clientErr != nil {
		log.Fatalln(clientErr)
	}
}
