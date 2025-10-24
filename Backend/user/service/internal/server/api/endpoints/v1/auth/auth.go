package auth

import (
	"auth/internal/config"
	"auth/internal/grpc-client/mailer"
	"auth/internal/repo"
	"auth/internal/repo/auth"
	"auth/internal/storage"
	"auth/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	CodeTheme = `Код подтверждения`
	CodeMsg   = `Ваш код подтверждения: %s`
)

func VerificationCode(redis *redis.Client, mailer *mailer.SenderClient, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Received request", "method", r.Method, "url", r.URL.Path)

		body := &VerifyReq{}

		if err := json.NewDecoder(r.Body).Decode(body); err != nil {
			log.Warn("failed to decode request", "error", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		code := utils.GenVerifyCode()
		sessionCode := uuid.New().String()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		
		data, err := json.Marshal(storage.VeriFyCode{
			Email: body.Email,
			VerifyCode: code,
		})
		log.Debug("Storing in Redis", "key", "verifycode:"+sessionCode, "value", string(data))
		if err != nil {
			log.Warn("failed to marshal verify code", "err", err)
			http.Error(w, "failed to generate code", http.StatusInternalServerError)
			return
		}

		if err := redis.Set(ctx, "verifycode:"+sessionCode, data, 15*time.Minute).Err(); err != nil {
			log.Warn("failed to save code", "err", err)
			http.Error(w, "failed to save code", http.StatusInternalServerError)
			return
		}

		if _, err := mailer.SendMsg(ctx, 
			body.Email,
			CodeTheme,
			fmt.Sprintf(CodeMsg, code),
		); err != nil {
			log.Warn("failed to send message", "err", err)
			http.Error(w, "failed to send message", http.StatusInternalServerError)
			return
		}

		response := &VerifyRes{
			SessionCode: sessionCode,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func Authentication(pgRepo *repo.Repo, cfg *config.Jwt, log *slog.Logger) http.HandlerFunc {
	CreateJwtToken := utils.CreateJwtTokenFunc(cfg, log)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Received request", "method", r.Method, "url", r.URL.Path)

		body := &AuthenticationReq{}

		if err := json.NewDecoder(r.Body).Decode(body); err != nil {
			log.Warn("failed to decode request", "error", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pgRes, err := pgRepo.User.GetUser(ctx, body.Login)
		if err != nil {
			log.Warn("failed to get user from db", "err", err)
			http.Error(w, "Incorrect login or password", http.StatusBadRequest)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(pgRes.Password), []byte(body.Password)); err != nil {
			log.Warn("Incorrect password", "err", err)
			http.Error(w, "Incorrect login or password", http.StatusBadRequest)
			return
		}

		jwt, err := CreateJwtToken(&utils.UserJwt{Id: pgRes.Id, Login: body.Login, Password: body.Password})
		if err != nil {
			log.Warn("JWT token generation failed", "err", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response := &AuthenticationRes{
			Jwt: jwt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func Registration(redis *redis.Client, pgRepo *repo.Repo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Received request", "method", r.Method, "url", r.URL.Path)

		body := &RegistrationReq{}

		if err := json.NewDecoder(r.Body).Decode(body); err != nil {
			log.Warn("failed to decode request", "error", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		redisJson, err := redis.Get(ctx, "verifycode:"+body.SessionCode).Bytes()
		if err != nil {
			log.Warn("incorrect session code", "error", err)
			http.Error(w, "incorrect code", http.StatusBadRequest)
			return
		}

		var redisBody storage.VeriFyCode
		if err := json.Unmarshal(redisJson, &redisBody); err != nil {
			log.Warn("failed to unmarshal json from redis", "error", err)
			http.Error(w, "failed to get verify code", http.StatusInternalServerError)
			return
		}
		log.Debug("Storing in Redis", "key", "verifycode:"+body.SessionCode, "redis value", string(redisJson), "user body", body)

		if redisBody.VerifyCode != body.VerifyCode || redisBody.Email != body.Email {
			http.Error(w, "incorrect code", http.StatusInternalServerError)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Warn("failed to hash password", "err", err)
			http.Error(w, "bad password", http.StatusInternalServerError)
			return
		}
		pgReq := &auth.CreateUserReq{
			Login: body.Login,
			Password: string(hashedPassword),
			Email: body.Email,
		}

		pgRes, err := pgRepo.User.Create(ctx, pgReq)
		if err != nil {
			log.Warn("failed to registration new user", "error", err)
			http.Error(w, "failed to registration", http.StatusInternalServerError)
			return
		}

		response := RegistrationRes{
			Id: pgRes.Id.String(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
