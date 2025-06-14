package main

import (
	"log/slog"
	"os"
	"s3manager/cmd"
	"s3manager/config"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	cnf, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	if err := cmd.Execute(cnf); err != nil {
		slog.Error("Failed to execute command", "error", err)
		os.Exit(1)
	}
}
