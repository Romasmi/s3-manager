package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Integration tests for upload command
// These tests require a real S3 connection and are skipped by default
// To run these tests, set the environment variable S3_INTEGRATION_TEST=true

func TestUploadCommand(t *testing.T) {
	if os.Getenv("S3_INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test; set S3_INTEGRATION_TEST=true to run")
	}

	tempFile, err := os.CreateTemp("", "upload-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := []byte("test content for upload command")
	if _, err := tempFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

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

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	uploadCmd.SetArgs([]string{
		tempFile.Name(),
		"--destination", "test-upload",
		"--no-archive",
		"--confirm",
	})
	err = uploadCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("Upload command failed: %v", err)
	}

	if !strings.Contains(output, filepath.Base(tempFile.Name())) {
		t.Errorf("Output doesn't contain file name: %s", output)
	}

	if !strings.Contains(output, "test-upload") {
		t.Errorf("Output doesn't contain destination path: %s", output)
	}

	if !strings.Contains(output, os.Getenv("TEST_BUCKET_NAME")) {
		t.Errorf("Output doesn't contain bucket name: %s", output)
	}
}

func TestIsDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dir-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile, err := os.CreateTemp("", "file-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	if !isDirectory(tempDir) {
		t.Errorf("isDirectory(%s) = false, want true", tempDir)
	}

	if isDirectory(tempFile.Name()) {
		t.Errorf("isDirectory(%s) = true, want false", tempFile.Name())
	}

	if isDirectory(filepath.Join(tempDir, "non-existent")) {
		t.Errorf("isDirectory(non-existent) = true, want false")
	}
}

func TestGetDestinationDisplay(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		expected    string
	}{
		{"Empty destination", "", "bucket root"},
		{"Root destination", "/", "/"},
		{"Folder destination", "folder", "folder"},
		{"Nested folder destination", "folder/subfolder", "folder/subfolder"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDestinationDisplay(tt.destination)
			if result != tt.expected {
				t.Errorf("getDestinationDisplay(%s) = %s, want %s", tt.destination, result, tt.expected)
			}
		})
	}
}

func TestCreateDryRunResult(t *testing.T) {
	paths := []string{"/path/to/file1.txt", "/path/to/file2.txt"}
	destination := "test-folder"
	bucketName := "test-bucket"

	result1 := createDryRunResult(paths, destination, true, bucketName)
	resultMap1, ok := result1.(map[string]interface{})
	if !ok {
		t.Fatalf("createDryRunResult() did not return a map")
	}

	if resultMap1["bucket_name"] != bucketName {
		t.Errorf("bucket_name = %v, want %v", resultMap1["bucket_name"], bucketName)
	}

	if resultMap1["destination_path"] != destination {
		t.Errorf("destination_path = %v, want %v", resultMap1["destination_path"], destination)
	}

	items1, ok := resultMap1["items"].([]interface{})
	if !ok {
		t.Fatalf("items is not a slice")
	}

	if len(items1) != 1 {
		t.Errorf("items length = %d, want %d", len(items1), 1)
	}

	result2 := createDryRunResult(paths, destination, false, bucketName)
	resultMap2, ok := result2.(map[string]interface{})
	if !ok {
		t.Fatalf("createDryRunResult() did not return a map")
	}

	items2, ok := resultMap2["items"].([]interface{})
	if !ok {
		t.Fatalf("items is not a slice")
	}

	if len(items2) != 2 {
		t.Errorf("items length = %d, want %d", len(items2), 2)
	}
}
