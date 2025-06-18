package s3client

import (
	"context"
	"os"
	"s3manager/config"
	"testing"
	"time"
)

// Integration tests for S3 client
// These tests require a real S3 connection and are skipped by default
// To run these tests, set the environment variable S3_INTEGRATION_TEST=true

func TestGetBucketInfo(t *testing.T) {
	if os.Getenv("S3_INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test; set S3_INTEGRATION_TEST=true to run")
	}

	cfg := &config.Config{
		BucketName: os.Getenv("TEST_BUCKET_NAME"),
		Region:     os.Getenv("TEST_REGION"),
		ApiURL:     os.Getenv("TEST_API_URL"),
		AccessKey:  os.Getenv("TEST_ACCESS_KEY"),
		SecretKey:  os.Getenv("TEST_SECRET_KEY"),
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	info, err := client.GetBucketInfo(context.Background())
	if err != nil {
		t.Fatalf("GetBucketInfo() error = %v", err)
	}

	if info.BucketName != cfg.BucketName {
		t.Errorf("BucketName = %s, want %s", info.BucketName, cfg.BucketName)
	}
}

func TestDeleteOldFiles(t *testing.T) {
	if os.Getenv("S3_INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test; set S3_INTEGRATION_TEST=true to run")
	}

	cfg := &config.Config{
		BucketName: os.Getenv("TEST_BUCKET_NAME"),
		Region:     os.Getenv("TEST_REGION"),
		ApiURL:     os.Getenv("TEST_API_URL"),
		AccessKey:  os.Getenv("TEST_ACCESS_KEY"),
		SecretKey:  os.Getenv("TEST_SECRET_KEY"),
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	result, err := client.DeleteOldFiles(context.Background(), "test", 30, true)
	if err != nil {
		t.Fatalf("DeleteOldFiles() error = %v", err)
	}

	if result.BucketName != cfg.BucketName {
		t.Errorf("BucketName = %s, want %s", result.BucketName, cfg.BucketName)
	}

	if result.Folder != "test" {
		t.Errorf("Folder = %s, want %s", result.Folder, "test")
	}

	if result.DaysOld != 30 {
		t.Errorf("DaysOld = %d, want %d", result.DaysOld, 30)
	}
}

func TestUploadFiles(t *testing.T) {
	if os.Getenv("S3_INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test; set S3_INTEGRATION_TEST=true to run")
	}

	cfg := &config.Config{
		BucketName: os.Getenv("TEST_BUCKET_NAME"),
		Region:     os.Getenv("TEST_REGION"),
		ApiURL:     os.Getenv("TEST_API_URL"),
		AccessKey:  os.Getenv("TEST_ACCESS_KEY"),
		SecretKey:  os.Getenv("TEST_SECRET_KEY"),
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tempFile, err := os.CreateTemp("", "s3client-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := []byte("test content for S3 upload")
	if _, err := tempFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	destinationPath := "test-" + time.Now().Format("20060102-150405")
	result, err := client.UploadFiles(context.Background(), []string{tempFile.Name()}, destinationPath, false)
	if err != nil {
		t.Fatalf("UploadFiles() error = %v", err)
	}

	if result.BucketName != cfg.BucketName {
		t.Errorf("BucketName = %s, want %s", result.BucketName, cfg.BucketName)
	}

	if result.DestinationPath != destinationPath {
		t.Errorf("DestinationPath = %s, want %s", result.DestinationPath, destinationPath)
	}

	if len(result.Items) != 1 {
		t.Errorf("Items length = %d, want %d", len(result.Items), 1)
	}

	if result.TotalFiles != 1 {
		t.Errorf("TotalFiles = %d, want %d", result.TotalFiles, 1)
	}

	if result.TotalSizeBytes != int64(len(content)) {
		t.Errorf("TotalSizeBytes = %d, want %d", result.TotalSizeBytes, len(content))
	}
}
