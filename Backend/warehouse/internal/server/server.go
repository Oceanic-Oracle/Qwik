package server

import (
    "context"
    "log/slog"
    "net/http"
    "warehouse/internal/config"
    "warehouse/internal/repo"
    "warehouse/internal/server/optimization"
    "warehouse/internal/server/products"
    reviews "warehouse/internal/server/review"
    "warehouse/internal/server/shelf"
    "warehouse/internal/server/warehouse"
    "warehouse/internal/service"

    "github.com/go-chi/chi"
)

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-Requested-With")
        w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
        w.Header().Set("Access-Control-Allow-Credentials", "true")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

type loggedKey struct{}

func loggingMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.Context().Value(loggedKey{}) == nil {
                log.Info("Incoming request",
                    slog.String("method", r.Method),
                    slog.String("path", r.URL.Path),
                    slog.String("remote_addr", r.RemoteAddr),
                    slog.String("user_agent", r.UserAgent()),
                    slog.Bool("auth", r.Header.Get("Authorization") != ""),
                )
                r = r.WithContext(context.WithValue(r.Context(), loggedKey{}, true))
            }
            next.ServeHTTP(w, r)
        })
    }
}

type Server struct {
    log       *slog.Logger
    cfg       *config.HTTP
    repo      *repo.Repo
    optimizer *service.OptimizationService
}

func (s *Server) CreateServer() func() {
    router := chi.NewRouter()

    router.Use(corsMiddleware)
    router.Use(loggingMiddleware(s.log))

    router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/plain")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("pong"))
    })

    router.Get("/products", products.GetProducts(s.repo, s.log))
    router.Get("/product/{id}", products.GetProduct(s.repo, s.log))
    router.Post("/products", products.CreateProduct(s.repo, s.log))

    router.Post("/review/create", reviews.CreateReview(s.repo, s.log))

    router.Post("/products/auto-stock", shelf.AutoRestockHandler(s.repo, s.log))
    router.Post("/shelves/withdraw", shelf.WithdrawHandler(s.repo, s.log))

    router.Post("/api/optimize", optimization.OptimizeHandler(s.optimizer, s.log))
    router.Post("/api/optimize/all", optimization.OptimizeAllHandler(s.optimizer, s.log))
    router.Get("/api/warehouse/map", warehouse.GetWarehouseMapHandler(s.repo, s.log))

    router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
        s.log.Warn("Method not allowed", slog.String("method", r.Method), slog.String("path", r.URL.Path))
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    })

    router.NotFound(func(w http.ResponseWriter, r *http.Request) {
        s.log.Warn("Route not found", slog.String("method", r.Method), slog.String("path", r.URL.Path))
        http.Error(w, "Not found", http.StatusNotFound)
    })

    srv := &http.Server{
        Addr:         s.cfg.Addr,
        Handler:      router,
        ReadTimeout:  s.cfg.Timeout,
        WriteTimeout: s.cfg.Timeout,
        IdleTimeout:  s.cfg.IdleTimeout,
    }

    go func() {
        s.log.Info("HTTP server starting", slog.String("addr", s.cfg.Addr))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            s.log.Error("HTTP server failed", slog.Any("error", err))
        }
    }()

    return func() {
        s.log.Info("Shutting down HTTP server")
        if err := srv.Close(); err != nil {
            s.log.Error("failed to close server", slog.Any("err", err))
        }
    }
}

func NewRestAPI(
    cfg *config.HTTP,
    repo *repo.Repo,
    log *slog.Logger,
    optimizer *service.OptimizationService,
) *Server {
    return &Server{
        cfg:       cfg,
        repo:      repo,
        log:       log,
        optimizer: optimizer,
    }
}