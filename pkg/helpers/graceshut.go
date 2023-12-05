package helpers

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func SetupGracefulShutdown(cancelFunc context.CancelFunc) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig
		cancelFunc()
	}()
}
