package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
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
	MigrationPath      string
	DatabaseURL        string
	MidtransServerKey  string
	MidtransEnv        string
	SupabaseURL        string
	SupabaseKey        string
	SupabaseBucket     string
	FrontendURL        string
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return fallback
}
func getEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	panic(fmt.Sprintf("Environment variable %s empty!", key))
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		slog.Warn(".env file not found")
	}
	return &Config{
		AppPort:            getEnv("APP_PORT", "8080"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnvOrPanic("DB_PASSWORD"),
		DBName:             getEnv("DB_NAME", "fynt_db"),
		JWTSecret:          getEnvOrPanic("JWT_SECRET"),
		RefreshTokenSecret: getEnvOrPanic("REFRESH_TOKEN_SECRET"),
		LogLevel:           getEnv("LOG_LEVEL", "INFO"),
		MigrationPath:      getEnv("MIGRATION_PATH", "file://infrastructure/database/migrations"),
		DatabaseURL:        getEnvOrPanic("DB_URL"),
		MidtransServerKey:  getEnvOrPanic("MIDTRANS_SERVER_KEY"),
		MidtransEnv:        getEnv("MIDTRANS_ENV", "sandbox"),
		SupabaseURL:        getEnv("SUPABASE_URL", "https://smltrmtpbnnuvnhpusjp.supabase.co"),
		SupabaseKey:        getEnvOrPanic("SUPABASE_KEY"),
		SupabaseBucket:     getEnv("SUPABASE_BUCKET", "fynt-documents"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}
