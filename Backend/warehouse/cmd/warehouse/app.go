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
    return 50.0, nil // заглушка
}

func (m *MockTurnoverProvider) GetDemandVariability(productID string) (float64, error) {
    return 0.25, nil // заглушка
}

func main() {
    cfg := config.MustLoad()
    log := pkg.SetupLogger(cfg.Env)

    pool := pkg.GetPgConnectionPool(cfg.PgStorage, log)
    
    // ИНИЦИАЛИЗАЦИЯ REDIS
    redisClient := pkg.GetRedisConnectionPool(cfg.RedisStorage, log)

    // Передаем Redis в репозиторий для распределенных блокировок
    rep := repo.NewRepo(pool, redisClient, log)

    // Инициализация сервиса оптимизации
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