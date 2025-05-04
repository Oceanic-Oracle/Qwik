package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	PATH = "../internal/config/.env"
)

type Config struct {
	Env    string `env:"ENV"`
	Addr   string `env:"ADDR"`
	Sender Sender
}

type Sender struct {
	SmtpHost     string `env:"SMTPHOST"`
	SmtpPort     string `env:"SMTPPORT"`
	SmtpUsername string `env:"SMTPUSERNAME"`
	SmtpPassword string `env:"SMTPPASSWORD"`
}

func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadConfig(PATH, &cfg); err != nil {
		log.Fatalf("can not read config: %s", err)
	}
	return &cfg
}
