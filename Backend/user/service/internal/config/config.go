package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	PATH = "../internal/config/.env"
)

type Config struct {
	Env              string `env:"ENV"`
	Htppserver       Httpserver
	PgStorage        PgStorage
	RedisStorage     RedisStorage
	MailerGrpcClient MailerGrpcClient
	Jwt              Jwt
}

type Httpserver struct {
	Addr         string `env:"ADDR"`
	Timeout      string `env:"TIMEOUT"`
	IddleTimeout string `env:"IDLE_TIMEOUT"`
	MaxConn      string `env:"MAX_CONN"`
}

type PgStorage struct {
	Username string `env:"POSTGRES_USERNAME"`
	Password string `env:"POSTGRES_PASSWORD"`
	Database string `env:"POSTGRES_DB"`
	Port     string `env:"POSTGRES_PORT"`
	Hosts    string `env:"POSTGRES_HOST"`
}

type RedisStorage struct {
	Host     string `env:"REDIS_HOST"`
	Port     string `env:"REDIS_PORT"`
	Password string `env:"REDIS_PASSWORD"`
}

type MailerGrpcClient struct {
	Host string `env:"MAILER_HOST"`
	Port string `env:"MAILER_ADDR"`
}

type Jwt struct {
	JwtTime string `env:"JWT_TIME"`
	JwtKey  string `env:"JWT_SECRET_KEY"`
}

func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadConfig(PATH, &cfg); err != nil {
		log.Fatalf("can not read config: %s", err)
	}
	return &cfg
}
