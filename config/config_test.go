package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default_value")
	if result != "test_value" {
		t.Errorf("getEnv() = %s, want %s", result, "test_value")
	}

	result = getEnv("NON_EXISTENT_VAR", "default_value")
	if result != "default_value" {
		t.Errorf("getEnv() = %s, want %s", result, "default_value")
	}

	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")

	result = getEnv("EMPTY_VAR", "default_value")
	if result != "default_value" {
		t.Errorf("getEnv() = %s, want %s", result, "default_value")
	}
}

func TestLoad(t *testing.T) {
	originalVars := map[string]string{
		"API_URL":     os.Getenv("API_URL"),
		"ACCESS_KEY":  os.Getenv("ACCESS_KEY"),
		"SECRET_KEY":  os.Getenv("SECRET_KEY"),
		"BUCKET_NAME": os.Getenv("BUCKET_NAME"),
		"REGION":      os.Getenv("REGION"),
	}

	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	testVars := map[string]string{
		"API_URL":     "https://test-api.example.com",
		"ACCESS_KEY":  "test-access-key",
		"SECRET_KEY":  "test-secret-key",
		"BUCKET_NAME": "test-bucket",
		"REGION":      "test-region",
	}

	for key, value := range testVars {
		os.Setenv(key, value)
	}

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.ApiURL != testVars["API_URL"] {
		t.Errorf("config.ApiURL = %s, want %s", config.ApiURL, testVars["API_URL"])
	}

	if config.AccessKey != testVars["ACCESS_KEY"] {
		t.Errorf("config.AccessKey = %s, want %s", config.AccessKey, testVars["ACCESS_KEY"])
	}

	if config.SecretKey != testVars["SECRET_KEY"] {
		t.Errorf("config.SecretKey = %s, want %s", config.SecretKey, testVars["SECRET_KEY"])
	}

	if config.BucketName != testVars["BUCKET_NAME"] {
		t.Errorf("config.BucketName = %s, want %s", config.BucketName, testVars["BUCKET_NAME"])
	}

	if config.Region != testVars["REGION"] {
		t.Errorf("config.Region = %s, want %s", config.Region, testVars["REGION"])
	}

	for key := range testVars {
		os.Unsetenv(key)
	}

	config, err = Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.ApiURL != "" {
		t.Errorf("config.ApiURL = %s, want %s", config.ApiURL, "")
	}

	if config.AccessKey != "" {
		t.Errorf("config.AccessKey = %s, want %s", config.AccessKey, "")
	}

	if config.SecretKey != "" {
		t.Errorf("config.SecretKey = %s, want %s", config.SecretKey, "")
	}

	if config.BucketName != "" {
		t.Errorf("config.BucketName = %s, want %s", config.BucketName, "")
	}

	if config.Region != "" {
		t.Errorf("config.Region = %s, want %s", config.Region, "")
	}
}
