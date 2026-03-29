package products

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"warehouse/internal/dto"
	"warehouse/internal/repo"
)

func GetProducts(repo *repo.Repo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		visibilityParam := r.URL.Query().Get("visibility")
		
		var visibility *bool
		if visibilityParam != "" {
			visibilityBool, err := strconv.ParseBool(visibilityParam)
			if err != nil {
				log.Error("Invalid visibility parameter", slog.Any("error", err))
				http.Error(w, "Invalid visibility parameter. Use true or false", http.StatusBadRequest)
				return
			}
			visibility = &visibilityBool
		}

		productsModel, err := repo.Product.GetProducts(ctx, visibility)
		if err != nil {
			log.Error("Failed to get products", slog.Any("error", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		productsDTO := make([]dto.Product, len(productsModel))
		for i, prod := range productsModel {
			productsDTO[i] = dto.Product{
				Id:          prod.Id,
				PreviewUrl:  prod.PreviewURL,
				Name:        prod.Name,
				Description: prod.Description,
				Price:       prod.Price,
				CreatedAt:   prod.CreatedAt,
				Count:       prod.Count,
				Visibility:  prod.Visibility,
				Avg:         prod.Avg,
				Reviews:     []dto.Review{},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(productsDTO); err != nil {
			log.Error("Failed to encode products response", slog.Any("error", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}
