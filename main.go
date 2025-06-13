package main

import (
	"log"
	"log/slog"
	"os"
	"s3manager/cmd"
	"s3manager/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cnf, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration " + err.Error())
	}
	slog.Info("Configuration loaded successfully")
	log.Printf("Configuration: %+v", cnf)

	if err := cmd.Execute(cnf); err != nil {
		log.Printf("Failed to execute command " + err.Error())
		os.Exit(1)
	}
}
