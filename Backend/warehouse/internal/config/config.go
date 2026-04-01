package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const PATH = "../.env"

type Config struct {
	Env          string `env:"ENV"`
	HTTP         HTTP
	PgStorage    PgStorage
	RedisStorage RedisStorage
	Optimizer    OptimizerConfig `env-prefix:"OPT_"` // Префикс для переменных оптимизатора
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
	URL      string `env:"REDIS_URL"`
}

// OptimizerConfig содержит параметры алгоритма оптимизации
type OptimizerConfig struct {
	// Веса для формулы скоринга полки: I_shelf = W1*P_level + W2*P_priority - W3*Distance
	WeightLevel    float64 `env:"WEIGHT_LEVEL" env-default:"0.4"`    // Влияние уровня полки
	WeightPriority float64 `env:"WEIGHT_PRIORITY" env-default:"0.4"` // Влияние приоритета
	WeightDistance float64 `env:"WEIGHT_DISTANCE" env-default:"0.2"` // Влияние расстояния

	// Пороги ABC-классификации (на основе принципа Парето)
	ABCCutoff float64 `env:"ABC_CUTOFF" env-default:"0.8"`  // Верхние 80% по ценности = класс A
	BCxCutoff float64 `env:"BCX_CUTOFF" env-default:"0.95"` // Следующие 15% = класс B, остаток = C

	// XYZ-классификация (вариабельность спроса)
	XYZCutoffLow  float64 `env:"XYZ_CUTOFF_LOW" env-default:"0.1"`  // CV < 0.1 = X (стабильный)
	XYZCutoffHigh float64 `env:"XYZ_CUTOFF_HIGH" env-default:"0.5"` // CV > 0.5 = Z (нестабильный)

	// TTL кэша для проверки совместимости тегов
	CompatCacheTTL time.Duration `env:"COMPAT_CACHE_TTL" env-default:"5m"`
}

func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadConfig(PATH, &cfg); err != nil {
		log.Fatalf("не удалось прочитать конфиг: %s", err)
	}
	return &cfg
}
