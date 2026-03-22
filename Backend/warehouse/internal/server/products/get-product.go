package products

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	"warehouse/internal/dto"
	"warehouse/internal/repo"

	"github.com/go-chi/chi"
)

func GetProduct(repo *repo.Repo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		productID := chi.URLParam(r, "id")

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		prodModel, reviewsModel, err := repo.Product.GetProductById(ctx, productID)
		if err != nil {
			log.Error("Failed to get product", slog.String("product_id", productID), slog.Any("error", err))
			if err.Error() == "product not found" {
				http.Error(w, "Product not found", http.StatusNotFound)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		prodDTO := dto.Product{
			Id:          prodModel.Id,
			PreviewUrl:  prodModel.PreviewURL,
			Name:        prodModel.Name,
			Description: prodModel.Description,
			Price:       prodModel.Price,
			CreatedAt:   prodModel.CreatedAt,
			Visibility:  prodModel.Visibility,

			Count:   prodModel.Count,
			Avg:     prodModel.Avg,
			Reviews: make([]dto.Review, len(reviewsModel)),
		}

		for i, rev := range reviewsModel {
			prodDTO.Reviews[i] = dto.Review{
				Id:          rev.Id,
				Login:       rev.Login,
				Grade:       rev.Grade,
				Description: rev.Description,
				CreatedAt:   rev.CreatedAt,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(prodDTO); err != nil {
			log.Error("Failed to encode response", slog.Any("error", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}
