package api

import (
	"auth/internal/config"
	"auth/internal/grpc-client/mailer"
	"auth/internal/repo"
	"auth/internal/server"
	"auth/internal/server/api/endpoints/ping"
	"auth/internal/server/api/endpoints/v1/auth"
	"auth/internal/server/api/endpoints/v1/profile"
	"auth/internal/server/middleware"
	pgstorage "auth/internal/storage/postgres"
	redisstorage "auth/internal/storage/redis"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	log *slog.Logger
	cfg *config.Config

	redisStorage *redis.Client
	pgRepo       *repo.Repo
	mailerClient *mailer.SenderClient

	limitMiddleware *middleware.IPRateLimiter

	//Do not touch
	srv *http.Server
}

func (s *Server) CreateServer() {
	router := chi.NewRouter()

	router.Use(s.limitMiddleware.Limit)
	router.Route("/ping", func(r chi.Router) {
		r.Get("/", ping.Pong(s.log))
	})
	router.Route("/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Post("/verify", auth.VerificationCode(s.redisStorage, s.mailerClient, s.log))
			r.Post("/login", auth.Authentication(s.pgRepo, &s.cfg.Jwt, s.log))
			r.Post("/registration", auth.Registration(s.redisStorage, s.pgRepo, s.log))
		})
		r.Route("/profile", func(r chi.Router) {
			r.Get("/", profile.GetProfile(s.pgRepo, &s.cfg.Jwt, s.log))
		})
	})

	idleTimeout, _ := strconv.Atoi(s.cfg.Htppserver.IddleTimeout)
	timeout, _ := strconv.Atoi(s.cfg.Htppserver.Timeout)
	srv := &http.Server{
		Addr:         s.cfg.Htppserver.Addr,
		Handler:      router,
		ReadTimeout:  time.Duration(timeout) * time.Second,
		WriteTimeout: time.Duration(timeout) * time.Second,
		IdleTimeout:  time.Duration(idleTimeout) * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			s.log.Error("failed to start server")
			return
		}
	}()

	s.srv = srv
}

func (s *Server) Close() error {
	s.limitMiddleware.Stop()
	return s.srv.Close()
}

func NewRestApi(cfg *config.Config, log *slog.Logger) server.Server {
	return &Server{
		cfg: cfg,
		log: log,

		redisStorage: redisstorage.GetConnectionPool(cfg.RedisStorage, log),
		pgRepo:       repo.NewRepo(pgstorage.GetConnectionPool(cfg.PgStorage, "disable", log), log),
		mailerClient: func() *mailer.SenderClient {
			res, err := mailer.NewSenderClient(cfg.MailerGrpcClient)
			if err != nil {
				log.Error("CRIRICAL: No connection to mailer", "err", err)
				return res
			}
			return res
		}(),
		limitMiddleware: func() *middleware.IPRateLimiter {
			maxConn, _ := strconv.Atoi(cfg.Htppserver.MaxConn)
			return middleware.NewIPRateLimiter(maxConn, time.Minute, 10)
		}(),
	}
}
