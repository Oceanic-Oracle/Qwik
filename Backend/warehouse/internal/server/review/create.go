package reviews

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	"warehouse/internal/repo"
	"warehouse/internal/repo/review"

	"github.com/google/uuid"
)

func CreateReview(repo *repo.Repo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Получаем JWT из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		var login *string

		if authHeader != "" {
			// TODO: Отправить JWT в микросервис пользователей для получения login
			// Пока заглушка - для демо можно установить тестовый login
			// login = extractLoginFromJWT(authHeader)

			// Временно, для тестирования:
			testLogin := "test_user" // В реальности получаем из сервиса пользователей
			login = &testLogin
		} else {
			login = nil // Анонимный пользователь
		}

		// Парсим JSON тело запроса
		var req struct {
			ProductID   string  `json:"product_id"`
			Grade       int     `json:"grade"`
			Description string `json:"description"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Warn("Failed to decode request body", slog.Any("error", err))
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		if _, err := uuid.Parse(req.ProductID); err != nil {
			http.Error(w, "Invalid product ID format", http.StatusBadRequest)
			return
		}

		// Валидация оценки
		if req.Grade < 1 || req.Grade > 5 {
			http.Error(w, "Grade must be between 1 and 5", http.StatusBadRequest)
			return
		}

		// Создаем объект отзыва
		reviewReq := review.Review{
			Login:       login,
			Grade:       req.Grade,
			Description: req.Description,
		}

		// Вызываем метод репозитория
		createdReview, err := repo.Review.CreateReview(ctx, req.ProductID, reviewReq)
		if err != nil {
			log.Error("Failed to create review",
				slog.Any("error", err),
				slog.String("product_id", req.ProductID),
				slog.Any("login", login))

			// Проверяем на конкретные ошибки
			if err.Error() == "user already reviewed this product" {
				http.Error(w, "You have already reviewed this product", http.StatusConflict)
				return
			}

			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Формируем ответ
		responseDTO := struct{
			Id int `json:"id"`
		}{
			Id: createdReview,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(responseDTO); err != nil {
			log.Error("Failed to encode review response", slog.Any("error", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Info("Review created successfully",
			slog.Int("review_id", createdReview),
			slog.String("product_id", req.ProductID),
			slog.Any("login", login))
	}
}
