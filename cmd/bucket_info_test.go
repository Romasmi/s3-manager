package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// Integration tests for bucket_info command
// These tests require a real S3 connection and are skipped by default
// To run these tests, set the environment variable S3_INTEGRATION_TEST=true

func TestBucketInfoCommand(t *testing.T) {
	if os.Getenv("S3_INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test; set S3_INTEGRATION_TEST=true to run")
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

	bucketInfoCmd.SetArgs([]string{})
	err := bucketInfoCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("Bucket info command failed: %v", err)
	}

	if !strings.Contains(output, os.Getenv("TEST_BUCKET_NAME")) {
		t.Errorf("Output doesn't contain bucket name: %s", output)
	}

	if !strings.Contains(output, "region") {
		t.Errorf("Output doesn't contain region: %s", output)
	}

	if !strings.Contains(output, "object_count") {
		t.Errorf("Output doesn't contain object_count: %s", output)
	}

	if !strings.Contains(output, "total_size") {
		t.Errorf("Output doesn't contain total_size: %s", output)
	}
}
