package optimization

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	"warehouse/internal/domain"
	"warehouse/internal/dto"
	"warehouse/internal/service"
)

// OptimizeHandler обрабатывает POST /api/optimize
func OptimizeHandler(svc *service.OptimizationService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		var req dto.OptimizeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Warn("не удалось декодировать запрос оптимизации", "error", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.WarehouseID == "" {
			http.Error(w, "warehouse_id is required", http.StatusBadRequest)
			return
		}

		// TODO: Здесь должна быть логика загрузки товаров по ID из БД
		// Для демо создаём заглушки
		products := make([]*domain.Product, len(req.ProductIDs))
		for i, pid := range req.ProductIDs {
			products[i] = &domain.Product{
				ID:                pid,
				AllocatedCapacity: 0.5, // заглушка
				Tags:              []string{"general"},
			}
		}

		results, err := svc.OptimizeAllocations(ctx, products, req.WarehouseID)
		if err != nil {
			log.Error("ошибка оптимизации", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Конвертация в DTO
		response := dto.OptimizeResponse{
			Results: make([]dto.AllocationResult, len(results)),
		}
		for i, res := range results {
			response.Results[i] = dto.AllocationResult{
				ProductID: res.ProductID,
				ShelfID:   res.ShelfID,
				Score:     res.Score,
				Assigned:  res.Assigned,
				Reason:    res.Reason,
			}
		}

		if req.DryRun {
			response.Message = "dry run mode: no changes applied"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("ошибка кодирования ответа", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}