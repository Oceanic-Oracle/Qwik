package products

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "net/http"
    "strconv"
    "time"
    "warehouse/internal/dto"
    "warehouse/internal/repo"
    "warehouse/internal/repo/product"
)

func CreateProduct(repo *repo.Repo, log *slog.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()

        var name string
        var description string
        var visibility bool
        var price int64
        var width, height, depth, weight, volume *float64
        previewURL := ""

        // Проверяем Content-Type
        contentType := r.Header.Get("Content-Type")
        
        if len(contentType) >= 19 && contentType[:19] == "multipart/form-data" {
            // Обработка multipart/form-data (с файлом)
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
                _ = header
            }

            name = r.FormValue("name")
            description = r.FormValue("description")
            visibility = r.FormValue("visibility") == "true"

            if _, err := fmt.Sscan(r.FormValue("price"), &price); err != nil || price < 0 {
                http.Error(w, "Invalid or missing price", http.StatusBadRequest)
                return
            }

            // Парсим размеры продукта из form-data
            if val := r.FormValue("width"); val != "" {
                if f, err := strconv.ParseFloat(val, 64); err == nil {
                    width = &f
                }
            }

            if val := r.FormValue("height"); val != "" {
                if f, err := strconv.ParseFloat(val, 64); err == nil {
                    height = &f
                }
            }

            if val := r.FormValue("depth"); val != "" {
                if f, err := strconv.ParseFloat(val, 64); err == nil {
                    depth = &f
                }
            }

            if val := r.FormValue("weight"); val != "" {
                if f, err := strconv.ParseFloat(val, 64); err == nil {
                    weight = &f
                }
            }

            if val := r.FormValue("volume"); val != "" {
                if f, err := strconv.ParseFloat(val, 64); err == nil {
                    volume = &f
                }
            }
        } else if contentType == "application/json" {
            // Обработка JSON (без файла)
            var reqDTO struct {
                Name        string   `json:"name"`
                Description string   `json:"description,omitempty"`
                Price       int64    `json:"price"`
                Visibility  bool     `json:"visibility"`
                Width       *float64 `json:"width,omitempty"`
                Height      *float64 `json:"height,omitempty"`
                Depth       *float64 `json:"depth,omitempty"`
                Weight      *float64 `json:"weight,omitempty"`
                Volume      *float64 `json:"volume,omitempty"`
            }
            
            if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
                log.Warn("Failed to decode JSON", slog.Any("error", err))
                http.Error(w, "Invalid JSON format", http.StatusBadRequest)
                return
            }
            
            name = reqDTO.Name
            description = reqDTO.Description
            price = reqDTO.Price
            visibility = reqDTO.Visibility
            width = reqDTO.Width
            height = reqDTO.Height
            depth = reqDTO.Depth
            weight = reqDTO.Weight
            volume = reqDTO.Volume
        } else {
            log.Warn("Unsupported Content-Type", slog.String("content-type", contentType))
            http.Error(w, "Unsupported Content-Type. Use multipart/form-data or application/json", http.StatusBadRequest)
            return
        }

        if name == "" {
            http.Error(w, "Product name is required", http.StatusBadRequest)
            return
        }

        if price <= 0 {
            http.Error(w, "Valid price is required", http.StatusBadRequest)
            return
        }

        // --- НОВОЕ: Валидация габаритов (строго положительные числа) ---
        if width != nil && *width <= 0 {
            http.Error(w, "Width must be a positive number", http.StatusBadRequest)
            return
        }
        if height != nil && *height <= 0 {
            http.Error(w, "Height must be a positive number", http.StatusBadRequest)
            return
        }
        if depth != nil && *depth <= 0 {
            http.Error(w, "Depth must be a positive number", http.StatusBadRequest)
            return
        }
        // -----------------------------------------------------------

        req := &product.CreateProduct{
            PreviewURL:  previewURL,
            Name:        name,
            Description: description,
            Price:       price,
            Width:       width,
            Height:      height,
            Depth:       depth,
            Weight:      weight,
            Volume:      volume,
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
            Width:       createdProduct.Width,
            Height:      createdProduct.Height,
            Depth:       createdProduct.Depth,
            Weight:      createdProduct.Weight,
            Volume:      createdProduct.Volume,
            CreatedAt:   createdProduct.CreatedAt,
            Visibility:  createdProduct.Visibility,
            Avg:         createdProduct.Avg,
            Count:       createdProduct.Count,
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