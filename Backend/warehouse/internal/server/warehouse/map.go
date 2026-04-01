package warehouse

import (
    "encoding/json"
    "log/slog"
    "net/http"
    "warehouse/internal/repo"
)

func GetWarehouseMapHandler(repo *repo.Repo, log *slog.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        warehouseMap, err := repo.Shelf.GetWarehouseMap(r.Context())
        if err != nil {
            log.Error("failed to get warehouse map", "error", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(warehouseMap)
    }
}