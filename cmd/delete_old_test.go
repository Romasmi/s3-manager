package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// Integration tests for delete_old command
// These tests require a real S3 connection and are skipped by default
// To run these tests, set the environment variable S3_INTEGRATION_TEST=true

func TestDeleteOldCommand(t *testing.T) {
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

	deleteOldCmd.SetArgs([]string{
		"--folder", "test",
		"--days", "30",
		"--dry-run",
	})
	err := deleteOldCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("Delete old command failed: %v", err)
	}

	if !strings.Contains(output, os.Getenv("TEST_BUCKET_NAME")) {
		t.Errorf("Output doesn't contain bucket name: %s", output)
	}

	if !strings.Contains(output, "test") {
		t.Errorf("Output doesn't contain folder name: %s", output)
	}

	if !strings.Contains(output, "30") {
		t.Errorf("Output doesn't contain days old: %s", output)
	}
}

func TestDaysValidation(t *testing.T) {
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		w.Close()
		os.Stdout = oldStdout
	}()

	deleteOldCmd.SetArgs([]string{
		"--days", "0",
		"--folder", "test",
		"--confirm",
	})
	err := deleteOldCmd.Execute()

	if err != nil {
		t.Errorf("deleteOldCmd.Execute() with days=0 returned error: %v", err)
	}

	deleteOldCmd.SetArgs([]string{
		"--days", "-1",
		"--folder", "test",
		"--confirm",
	})
	err = deleteOldCmd.Execute()

	if err != nil {
		t.Errorf("deleteOldCmd.Execute() with days=-1 returned error: %v", err)
	}
}
