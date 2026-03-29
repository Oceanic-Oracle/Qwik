package shelf

import (
    "encoding/json"
    "log/slog"
    "net/http"
    "warehouse/internal/dto"
    "warehouse/internal/repo"
    "warehouse/internal/repo/shelf"

    "errors"
)

// AutoRestockHandler автоматически размещает товар на лучшей полке
func AutoRestockHandler(repo *repo.Repo, log *slog.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var req dto.AutoStockRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        if req.ProductID == "" || req.Quantity <= 0 {
            http.Error(w, "product_id and positive quantity are required", http.StatusBadRequest)
            return
        }

        res, err := repo.Shelf.AutoAddStock(r.Context(), req.ProductID, req.Quantity)
        if err != nil {
            if errors.Is(err, shelf.ErrNoSpace) {
                http.Error(w, err.Error(), http.StatusConflict)
                return
            }
            if errors.Is(err, shelf.ErrProductNotFound) {
                http.Error(w, err.Error(), http.StatusNotFound)
                return
            }
            if errors.Is(err, shelf.ErrNoVolume) {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }
            // ИСПРАВЛЕНИЕ: Обработка новой ошибки блокировки полки
            if errors.Is(err, shelf.ErrShelfLocked) {
                http.Error(w, err.Error(), http.StatusLocked) // 423 Locked
                return
            }
            log.Error("failed to auto restock", "error", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(res)
    }
}