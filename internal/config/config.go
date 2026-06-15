package config

import (
	"log"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	OTel     OTelConfig
	MinIO    MinIOConfig
}

type AppConfig struct {
	Env     string `env:"APP_ENV" envDefault:"development"`
	Port    string `env:"APP_PORT" envDefault:"8080"`
	BaseURL string `env:"APP_BASE_URL" envDefault:"http://localhost:8080"`
}

type DatabaseConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"POSTGRES_USER" envDefault:"postgres"`
	Password string `env:"POSTGRES_PASSWORD"`
	Name     string `env:"POSTGRES_DB" envDefault:"votesystem"`
	MaxConns int32  `env:"DB_MAX_CONNS" envDefault:"10"`
	MinConns int32  `env:"DB_MIN_CONNS" envDefault:"2"`
}

func (d DatabaseConfig) DSN() string {
	return "host=" + d.Host +
		" port=" + d.Port +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.Name +
		" sslmode=disable"
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST" envDefault:"localhost"`
	Port     string `env:"REDIS_PORT" envDefault:"6379"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

func (r RedisConfig) Addr() string {
	return r.Host + ":" + r.Port
}

type MinIOConfig struct {
	Endpoint     string `env:"MINIO_ENDPOINT" envDefault:"localhost:9000"`
	AccessKey    string `env:"MINIO_ACCESS_KEY"`
	SecretKey    string `env:"MINIO_SECRET_KEY"`
	Bucket       string `env:"MINIO_BUCKET" envDefault:"votesystem"`
	UseSSL       bool   `env:"MINIO_USE_SSL" envDefault:"false"`
	PublicBaseURL string `env:"MINIO_PUBLIC_BASE_URL" envDefault:""`
}

type JWTConfig struct {
	Secret          string `env:"JWT_SECRET,required"`
	AccessExpMins   int    `env:"JWT_ACCESS_EXP_MINUTES" envDefault:"15"`
	RefreshExpHours int    `env:"JWT_REFRESH_EXP_HOURS" envDefault:"168"`
}

type OTelConfig struct {
	Endpoint    string  `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"localhost:4317"`
	ServiceName string  `env:"OTEL_SERVICE_NAME" envDefault:"votesystem"`
	SamplerRate float64 `env:"OTEL_SAMPLER_RATE" envDefault:"1.0"`
}

func Load() *Config {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: failed to load .env file: %v", err)
		}
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	return cfg
}
