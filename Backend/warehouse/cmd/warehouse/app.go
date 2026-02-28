package main

import (
	"os"
	"os/signal"
	"syscall"
	"warehouse/internal/config"
	"warehouse/internal/repo"
	"warehouse/internal/server"
	"warehouse/pkg"
)

func main() {
	cfg := config.MustLoad()
	log := pkg.SetupLogger(cfg.Env)

	pool := pkg.GetPgConnectionPool(cfg.PgStorage, log)
	rep := repo.NewRepo(pool, log)

	srv := server.NewRestAPI(&cfg.HTTP, rep, log)
	srvClose := srv.CreateServer()
	defer srvClose()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
}