package optimization

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	"warehouse/internal/dto"
	"warehouse/internal/service"
)

func OptimizeHandler(svc *service.OptimizationService, log *slog.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
        defer cancel()

        var req struct {
            Items  []dto.OptimizeItem `json:"items"`
            DryRun bool                `json:"dry_run"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        results, err := svc.OptimizeAllocations(ctx, req.Items)
        if err != nil {
            log.Error("ошибка оптимизации", "error", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(results)
    }
}

func OptimizeAllHandler(svc *service.OptimizationService, log *slog.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
        defer cancel()

        log.Info("Запущен глобальный рефакторинг склада...")
        results, err := svc.OptimizeAllAllocations(ctx)
        if err != nil {
            log.Error("ошибка глобальной оптимизации", "error", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(results)
    }
}