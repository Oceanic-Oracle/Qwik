package main

import (
	"mailer/internal/config"
	"mailer/internal/logger"
	grpcserver "mailer/internal/server/grpc-server"
	"mailer/internal/utils"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	srv, err := grpcserver.NewServer(log, cfg)
	if err != nil {
		panic(err)
	}
	defer srv.Stop()

	utils.Shutdown(log)

	log.Info("Server stoped")
}