package utils

import (
	"auth/internal/config"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type UserJwt struct {
	Id       string
	Login    string
	Password string

	jwt.RegisteredClaims
}

func CreateJwtTokenFunc(cfg *config.Jwt, log *slog.Logger) func(*UserJwt) (string, error) {
	return func(userdto *UserJwt) (tokenString string, err error) {
		jwtTime, err := time.ParseDuration(cfg.JwtTime)
		if err != nil {
			log.Error("Error parsing JWT time duration", slog.Any("err", err))
			return "", err
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserJwt{
			Id:       userdto.Id,
			Login:    userdto.Login,
			Password: userdto.Password,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtTime)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				Issuer:    "Qwik",
			},
		})

		tokenString, err = token.SignedString([]byte(cfg.JwtKey))
		if err != nil {
			log.Error("Error to create token", slog.Any("err", err))
			return "", err
		}

		return tokenString, nil
	}
}

func ParseJWTFunc(cfg *config.Jwt, log *slog.Logger) func(jwtoken string) (*UserJwt, error) {
	return func(jwtoken string) (*UserJwt, error) {
		token, err := jwt.ParseWithClaims(jwtoken, &UserJwt{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JwtKey), nil
		})

		if err != nil {
			log.Error("Error parsing JWT token", slog.Any("err", err))
			return nil, err
		}

		if claims, ok := token.Claims.(*UserJwt); ok && token.Valid {
			userDto := &UserJwt{
				Login:    claims.Login,
				Password: claims.Password,
			}
			return userDto, nil
		}

		log.Error("Invalid JWT token")
		return nil, fmt.Errorf("invalid JWT token")
	}
}
