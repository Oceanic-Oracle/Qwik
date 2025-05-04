package utils

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func Shutdown(log *slog.Logger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	log.Info("Server started")

	<-stop
}
