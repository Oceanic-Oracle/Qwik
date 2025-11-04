package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	PATH = "../.env"
)

type Config struct {
	Env              string `env:"ENV"`
	Htppserver       Httpserver
	PgStorage        PgStorage
	RedisStorage     RedisStorage
}

type Httpserver struct {
	Addr         string `env:"ADDR"`
	Timeout      string `env:"TIMEOUT"`
	IddleTimeout string `env:"IDLE_TIMEOUT"`
	MaxConn      string `env:"MAX_CONN"`
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