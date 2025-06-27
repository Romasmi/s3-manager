package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// Integration tests for download command
// These tests require a real S3 connection and are skipped by default
// To run these tests, set the environment variable S3_INTEGRATION_TEST=true

func TestDownloadCommand(t *testing.T) {
	if os.Getenv("S3_INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test; set S3_INTEGRATION_TEST=true to run")
	}

	// Create a temporary directory to download files to
	tempDir, err := os.MkdirTemp("", "download-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set environment variables for S3 connection
	os.Setenv("BUCKET_NAME", os.Getenv("TEST_BUCKET_NAME"))
	os.Setenv("REGION", os.Getenv("TEST_REGION"))
	os.Setenv("API_URL", os.Getenv("TEST_API_URL"))
	os.Setenv("ACCESS_KEY", os.Getenv("TEST_ACCESS_KEY"))
	os.Setenv("SECRET_KEY", os.Getenv("TEST_SECRET_KEY"))
	defer func() {
		os.Unsetenv("BUCKET_NAME")
		os.Unsetenv("REGION")
		os.Unsetenv("API_URL")
		os.Unsetenv("ACCESS_KEY")
		os.Unsetenv("SECRET_KEY")
	}()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the download command
	// Note: This test assumes that the "test-upload" folder exists in the bucket
	// and contains at least one file (created by the upload test)
	downloadCmd.SetArgs([]string{
		"test-upload",
		"--destination", tempDir,
		"--confirm",
	})
	err = downloadCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("Download command failed: %v", err)
	}

	// Check if output contains expected information
	if !strings.Contains(output, "test-upload") {
		t.Errorf("Output doesn't contain source path: %s", output)
	}

	if !strings.Contains(output, tempDir) {
		t.Errorf("Output doesn't contain destination path: %s", output)
	}

	if !strings.Contains(output, os.Getenv("TEST_BUCKET_NAME")) {
		t.Errorf("Output doesn't contain bucket name: %s", output)
	}

	// Check if a file was actually downloaded
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	if len(files) == 0 {
		t.Errorf("No files were downloaded to %s", tempDir)
	}
}
