package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"s3manager/internal/models"
	"strings"
	"testing"
	"time"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"Zero bytes", 0, "0 B"},
		{"Bytes", 500, "500 B"},
		{"Kilobytes", 1500, "1.5 KB"},
		{"Megabytes", 1500000, "1.4 MB"},
		{"Gigabytes", 1500000000, "1.4 GB"},
		{"Terabytes", 1500000000000, "1.4 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestPrintJSON(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	testData := map[string]string{"key": "value"}

	err := PrintJSON(testData)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Errorf("PrintJSON() returned error: %v", err)
	}

	var result map[string]string
	err = json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Errorf("PrintJSON() produced invalid JSON: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("PrintJSON() output = %v, want %v", result, testData)
	}
}

func TestPrintError(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	testErr := errors.New("test error")
	testCmd := "test-command"

	PrintError(testErr, testCmd)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "test error") {
		t.Errorf("PrintError() output doesn't contain error message: %s", output)
	}

	if !strings.Contains(output, "test-command") {
		t.Errorf("PrintError() output doesn't contain command: %s", output)
	}

	var result models.ErrorResponse
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Errorf("PrintError() produced invalid JSON: %v", err)
	}

	if result.Error != "test error" {
		t.Errorf("PrintError() error = %s, want %s", result.Error, "test error")
	}

	if result.Command != "test-command" {
		t.Errorf("PrintError() command = %s, want %s", result.Command, "test-command")
	}
}

func TestFormatTime(t *testing.T) {
	testTime := time.Date(2023, 5, 15, 10, 30, 0, 0, time.UTC)
	expected := "2023-05-15T10:30:00Z" // RFC3339 format

	result := FormatTime(testTime)
	if result != expected {
		t.Errorf("FormatTime() = %s, want %s", result, expected)
	}
}
