package main

import (
	"log"
	"os"
	"s3manager/cmd"
	"s3manager/config"
)

func main() {
	cnf, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := cmd.Execute(cnf); err != nil {
		log.Printf("Error executing command: %v", err)
		os.Exit(1)
	}
}
