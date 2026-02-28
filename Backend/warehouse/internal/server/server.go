package server

import (
	"log/slog"
	"net/http"
	"warehouse/internal/config"
	"warehouse/internal/repo"
	"warehouse/internal/server/products"

	"github.com/go-chi/chi"
)

type Server struct {
	log  *slog.Logger
	cfg  *config.HTTP
	repo *repo.Repo
}

func (s *Server) CreateServer() func() {
	router := chi.NewRouter()

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ping"))
	})

	router.Get("/products", products.GetProducts(s.repo, s.log))
	router.Get("/product/{id}", products.GetProduct(s.repo, s.log))

	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		router.ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:         s.cfg.Addr,
		Handler:      corsHandler,
		ReadTimeout:  s.cfg.Timeout,
		WriteTimeout: s.cfg.Timeout,
		IdleTimeout:  s.cfg.IdleTimeout,
	}

	go func() {
		s.log.Info("HTTP server starting", slog.String("addr", s.cfg.Addr))

		if err := srv.ListenAndServe(); err != nil {
			s.log.Error("HTTP server failed", slog.Any("error", err))
			return
		}
	}()

	return func() {
		if err := srv.Close(); err != nil {
			s.log.Error("failed to close server", slog.Any("err", err))
		}
	}
}

func NewRestAPI(cfg *config.HTTP, repo *repo.Repo, log *slog.Logger) *Server {
	return &Server{
		cfg:  cfg,
		repo: repo,
		log:  log,
	}
}
