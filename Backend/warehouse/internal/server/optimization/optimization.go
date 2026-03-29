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

        var req dto.OptimizeRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        if len(req.Items) == 0 {
            http.Error(w, "items are required", http.StatusBadRequest)
            return
        }

        for _, item := range req.Items {
            if item.Quantity <= 0 {
                http.Error(w, "quantity must be > 0", http.StatusBadRequest)
                return
            }
        }

        results, err := svc.OptimizeAllocations(ctx, req.Items)
        if err != nil {
            log.Error("ошибка оптимизации", "error", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        response := dto.OptimizeResponse{Results: make([]dto.AllocationResult, len(results))}
        for i, res := range results {
            response.Results[i] = dto.AllocationResult{
                ProductID: res.ProductID,
                ShelfID:   res.ShelfID,
                Assigned:  res.Assigned,
                Reason:    res.Reason,
                Score:     res.Score,
            }
        }

        if req.DryRun {
            response.Message = "dry run mode: no changes applied"
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response)
    }
}