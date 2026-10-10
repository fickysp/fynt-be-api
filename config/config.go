package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppPort            string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	JWTSecret          string
	RefreshTokenSecret string
	LogLevel           string
}

func getEnv(key, fallback string) string {
	if os.Getenv(key) != "" {
		return key
	}
	return fallback
}
func getEnvOrPanic(key string) string {
	if os.Getenv(key) != "" {
		return key
	}
	panic(fmt.Sprintf("FATAL: Environment variable %s WAJIB diisi!", key))
}

func Load() *Config {
	return &Config{
		AppPort:            getEnv("APP_PORT", "8080"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnvOrPanic("DB_PASSWORD"),
		DBName:             getEnv("DB_NAME", "fynt-db"),
		JWTSecret:          getEnvOrPanic("JWT_SECRET"),
		RefreshTokenSecret: getEnvOrPanic("REFRESH_TOKEN_SECRET"),
		LogLevel:           getEnv("LOG_LEVEL", "INFO"),
	}
}
