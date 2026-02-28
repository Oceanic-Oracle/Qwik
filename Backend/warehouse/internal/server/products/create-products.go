package products

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"warehouse/internal/dto"
	"warehouse/internal/repo"
	"warehouse/internal/repo/product"
)

func CreateProduct(repo *repo.Repo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			log.Warn("Failed to parse multipart form", slog.Any("error", err))
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("image")
		if err != nil && err != http.ErrMissingFile {
			log.Warn("Failed to read uploaded file", slog.Any("error", err))
			http.Error(w, "Failed to process image", http.StatusBadRequest)
			return
		}
		if file != nil {
			defer file.Close()
			// TODO: Загрузка файла в S3-хранилище
			previewURL := "" 
			_ = header       
			_ = previewURL   
		}

		name := r.FormValue("name")
		description := r.FormValue("description")
		visibility := r.FormValue("visibility") == "true"

		var price int64
		if _, err := fmt.Sscan(r.FormValue("price"), &price); err != nil || price < 0 {
			http.Error(w, "Invalid or missing price", http.StatusBadRequest)
			return
		}

		if name == "" {
			http.Error(w, "Product name is required", http.StatusBadRequest)
			return
		}

		req := &product.CreateProduct{
			PreviewURL:  "",
			Name:        name,
			Description: description,
			Price:       price,
			Visibility:  visibility,
		}

		createdProduct, err := repo.Product.CreateProduct(ctx, req)
		if err != nil {
			log.Error("Failed to create product", slog.Any("error", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		responseDTO := dto.Product{
			Id:          createdProduct.Id,
			PreviewUrl:  createdProduct.PreviewURL,
			Name:        createdProduct.Name,
			Description: createdProduct.Description,
			Price:       createdProduct.Price,
			CreatedAt:   createdProduct.CreatedAt,
			Visibility:  createdProduct.Visibility,
			Avg:         createdProduct.Avg,
			Reviews:     []dto.Review{},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(responseDTO); err != nil {
			log.Error("Failed to encode product response", slog.Any("error", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Info("Product created successfully",
			slog.String("product_id", createdProduct.Id),
			slog.String("name", createdProduct.Name))
	}
}