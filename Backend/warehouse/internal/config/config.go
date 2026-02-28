package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	PATH = "../.env"
)

type Config struct {
	Env          string `env:"ENV"`
	HTTP         HTTP
	PgStorage    PgStorage
	RedisStorage RedisStorage
}

type HTTP struct {
	Addr        string        `env:"SERVER_ADDR"`
	Timeout     time.Duration `env:"SERVER_TIMEOUT_SECONDS"`
	IdleTimeout time.Duration `env:"SERVER_IDLE_TIMEOUT_SECONDS"`
	MaxConn     string        `env:"SERVER_MAX_CONN"`
}

type PgStorage struct {
	PostgresUrlsWrite string `env:"POSTGRES_URLS_WRITE"`
	PostgresUrlsRead  string `env:"POSTGRES_URLS_READ"`
}

type RedisStorage struct {
	Host     string `env:"REDIS_HOST"`
	Port     string `env:"REDIS_PORT"`
	Password string `env:"REDIS_PASSWORD"`
}

func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadConfig(PATH, &cfg); err != nil {
		log.Fatalf("can not read config: %s", err)
	}
	return &cfg
}
