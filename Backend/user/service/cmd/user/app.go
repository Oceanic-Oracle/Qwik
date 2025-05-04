package main

import (
	"auth/internal/config"
	"auth/internal/logger"
	"auth/internal/server/api"
	"auth/internal/utils"
)

func main() {
	cfg := config.MustLoad()
	
	log := logger.SetupLogger(cfg.Env)

	//Initialize server
	srv := api.NewRestApi(cfg, log)
	srv.CreateServer()
	defer srv.Close()
	
	//Graceful shutdown
	utils.Shutdown(log)

	log.Info("Server stoped")
}