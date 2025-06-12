package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	ApiURL     string
	Token      string
	AccessKey  string
	SecretKey  string
	BucketName string
	Region     string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found, using environment variables only")
		panic(err)
	}

	config := &Config{
		ApiURL:     getEnv("API_URL", ""),
		Token:      getEnv("TOKEN", ""),
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
