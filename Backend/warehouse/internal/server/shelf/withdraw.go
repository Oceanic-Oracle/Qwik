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

func WithdrawHandler(repo *repo.Repo, log *slog.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var req dto.StockRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        if req.ShelfID == "" || req.ProductID == "" || req.Quantity <= 0 {
            http.Error(w, "shelf_id, product_id and positive quantity are required", http.StatusBadRequest)
            return
        }

        res, err := repo.Shelf.WithdrawStock(r.Context(), req.ShelfID, req.ProductID, req.Quantity)
        if err != nil {
            if errors.Is(err, shelf.ErrNotEnoughStock) {
                http.Error(w, err.Error(), http.StatusConflict)
                return
            }
            log.Error("failed to withdraw", "error", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(res)
    }
}