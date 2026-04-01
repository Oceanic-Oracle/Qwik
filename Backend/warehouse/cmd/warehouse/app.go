package main

import (
    "os"
    "os/signal"
    "syscall"
    "warehouse/internal/config"
    "warehouse/internal/repo"
    "warehouse/internal/server"
    "warehouse/internal/service"
    "warehouse/pkg"
)

// MockTurnoverProvider — заглушка для тестов/демо
type MockTurnoverProvider struct{}

func (m *MockTurnoverProvider) GetTurnoverRate(productID string) (float64, error) {
    hash := 0
    for i := 0; i < len(productID); i++ {
        hash = int(productID[i]) + ((hash << 5) - hash)
    }
    return 10.0 + float64((hash%100)), nil
}

func (m *MockTurnoverProvider) GetDemandVariability(productID string) (float64, error) {
    return 0.25, nil
}

func main() {
    cfg := config.MustLoad()
    log := pkg.SetupLogger(cfg.Env)

    pool := pkg.GetPgConnectionPool(cfg.PgStorage, log)
    redisClient := pkg.GetRedisConnectionPool(cfg.RedisStorage, log)

    rep := repo.NewRepo(pool, redisClient, log)

    optimizer := service.NewOptimizationService(
        rep.Shelf,
        cfg.Optimizer,
        log,
        &MockTurnoverProvider{},
    )

    srv := server.NewRestAPI(&cfg.HTTP, rep, log, optimizer)
    srvClose := srv.CreateServer()
    defer srvClose()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
    <-stop
}