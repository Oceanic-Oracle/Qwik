package profile

import (
	"auth/internal/config"
	"auth/internal/repo"
	"auth/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	//"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GetProfile(pgRepo *repo.Repo, cfg *config.Jwt, log *slog.Logger) http.HandlerFunc {
	ParseJWT := utils.ParseJWTFunc(cfg, log)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Received request", "method", r.Method, "url", r.URL.Path)
    
		authHeader := r.Header.Get("Authorization")
		log.Debug("Authorization header", "header", authHeader)
		
		if authHeader == "" {
			log.Error("Unauthorized: missing token")
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(authHeader)
		if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
			tokenString = strings.TrimSpace(tokenString[7:])
		}

		if tokenString == "" {
			log.Error("Unauthorized: invalid token format")
			http.Error(w, "Unauthorized: invalid token format", http.StatusUnauthorized)
			return
		}

		parseToken, err := ParseJWT(tokenString)
		if err != nil {
			log.Warn("JWT parsing failed", "err", err)
			
			switch {
			case errors.Is(err, jwt.ErrTokenMalformed):
				http.Error(w, "Unauthorized: invalid token format", http.StatusUnauthorized)
			case errors.Is(err, jwt.ErrTokenExpired):
				http.Error(w, "Unauthorized: token expired", http.StatusUnauthorized)
			case errors.Is(err, jwt.ErrTokenNotValidYet):
				http.Error(w, "Unauthorized: token not active yet", http.StatusUnauthorized)
			default:
				http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			}
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pgRes, err := pgRepo.Profile.GetProfile(ctx, parseToken.Login)
		if err != nil {
			log.Warn("failed to get user from db", "err", err)
			http.Error(w, "Incorrect user data", http.StatusBadRequest)
			return
		}

		response := &GetProfileRes{
			Surname:    pgRes.Surname,
			Name:       pgRes.Name,
			Patronymic: pgRes.Patronymic,
			CreatedAt:  pgRes.CreatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
