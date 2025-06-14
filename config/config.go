package config

import (
	"github.com/joho/godotenv"
	"log/slog"
	"os"
)

type Config struct {
	ApiURL     string
	AccessKey  string
	SecretKey  string
	BucketName string
	Region     string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found, using environment variables only")
	}

	config := &Config{
		ApiURL:     getEnv("API_URL", ""),
		AccessKey:  getEnv("ACCESS_KEY", ""),
		SecretKey:  getEnv("SECRET_KEY", ""),
		BucketName: getEnv("BUCKET_NAME", ""),
		Region:     getEnv("REGION", ""),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
