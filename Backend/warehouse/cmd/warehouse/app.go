package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"warehouse/internal/config"
	"warehouse/internal/graph"
	"warehouse/internal/repo"
	"warehouse/internal/server"
	"warehouse/pkg"
)

func main() {
	cfg := config.MustLoad()
	log := pkg.SetupLogger(cfg.Env)

	pool := pkg.GetPgConnectionPool(cfg.PgStorage, log)
	rep := repo.NewRepo(pool, log)
	graphResolver := &graph.Resolver{
		Repo: rep,
		Log: log,
	}

	http.Handle("/query", server.NewGraphQLHandler(graphResolver))

	srv := http.Server{Addr: cfg.Htppserver.Addr}
	go func() {
		log.Info("GraphQL server starting", "addr", cfg.Htppserver.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
}