package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppEnv   string `envconfig:"APP_ENV" default:"development"`
	HTTPPort string `envconfig:"HTTP_PORT" default:"8080"`

	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	JWTSecret     string `envconfig:"JWT_SECRET" required:"true"`
	JWTExpiration string `envconfig:"JWT_EXPIRATION" default:"24h"`

	RedisURL string `envconfig:"REDIS_URL"`

	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	CORSAllowedOrigins string `envconfig:"CORS_ALLOWED_ORIGINS" default:"http://localhost:5173,http://localhost:3000"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
