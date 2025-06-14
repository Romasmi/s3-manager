package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"s3manager/internal/models"
	"time"
)

func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func PrintJSON(data interface{}) error {
	jsonOutput, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonOutput))
	return nil
}

func PrintError(err error, command string) {
	errorResp := models.ErrorResponse{
		Error:     err.Error(),
		Timestamp: time.Now().Format(time.RFC3339),
		Command:   command,
	}
	err = PrintJSON(errorResp)
	if err != nil {
		slog.Error("Failed to print error in JSON format", "error", err)
		fmt.Println("Error: ", errorResp)
		return
	}
}

func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}
